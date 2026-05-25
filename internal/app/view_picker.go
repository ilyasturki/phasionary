package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"phasionary/internal/ui"
)

const (
	pickerWidthFraction   = 0.5
	pickerContentMinWidth = 50
	pickerContentMaxWidth = 70
)

var pickerDialogChromeWidth = ui.HelpDialogStyle.GetHorizontalPadding() + ui.HelpDialogStyle.GetHorizontalBorderSize()

func (m model) pickerContentWidth() int {
	total := m.ui.Screen.Width
	if total <= 0 {
		return pickerContentMinWidth
	}
	target := int(float64(total) * pickerWidthFraction)
	content := target - pickerDialogChromeWidth
	if content < pickerContentMinWidth {
		content = pickerContentMinWidth
	}
	if content > pickerContentMaxWidth {
		content = pickerContentMaxWidth
	}
	if maxContent := total - pickerDialogChromeWidth - 2; content > maxContent && maxContent > 10 {
		content = maxContent
	}
	if content < 10 {
		content = 10
	}
	return content
}

func padToWidth(line string, width int) string {
	w := lipgloss.Width(line)
	if w >= width {
		return line
	}
	return line + strings.Repeat(" ", width-w)
}

func (m model) projectPickerView() string {
	contentWidth := m.pickerContentWidth()

	lines := []string{
		ui.DialogTitleStyle.Render("Select Project"),
		"",
	}

	total := m.ui.Picker.totalItems()
	visibleEnd := m.ui.Picker.scrollOffset + pickerVisibleItems
	if visibleEnd > total {
		visibleEnd = total
	}

	if m.ui.Picker.scrollOffset > 0 {
		lines = append(lines, ui.DialogHintStyle.Render(scrollMoreAbove))
	}

	addButtonIdx := len(m.ui.Picker.projects)
	for i := m.ui.Picker.scrollOffset; i < visibleEnd; i++ {
		isSelected := i == m.ui.Picker.selected
		if i == addButtonIdx {
			lines = append(lines,
				ui.DialogHintStyle.Render(strings.Repeat("─", contentWidth)),
				m.renderAddProjectLine(isSelected, contentWidth),
			)
			continue
		}
		lines = append(lines, m.renderPickerRow(i, isSelected, contentWidth))
	}

	if visibleEnd < total {
		lines = append(lines, ui.DialogHintStyle.Render(scrollMoreBelow))
	}

	hintText := "J/K reorder · ⏎ select · d delete · esc cancel"
	if m.ui.Picker.isAdding {
		hintText = "⏎ create · esc cancel"
	}
	lines = append(lines, "", ui.DialogHintStyle.Render(hintText))

	// lipgloss.Width includes horizontal padding but not the border.
	return ui.HelpDialogStyle.Width(contentWidth + ui.HelpDialogStyle.GetHorizontalPadding()).Render(strings.Join(lines, "\n"))
}

func (m model) renderPickerRow(i int, isSelected bool, contentWidth int) string {
	p := m.ui.Picker.projects[i]
	prefix := "  "
	if isSelected {
		prefix = "> "
	}

	isCurrent := p.ID == m.project.ID

	if isSelected {
		line := prefix + p.Name
		if isCurrent {
			line += " (current)"
		}
		return ui.SelectedStyle.Render(padToWidth(line, contentWidth))
	}

	if isCurrent {
		return prefix + p.Name + ui.DialogHintStyle.Render(" (current)")
	}
	return prefix + p.Name
}

func (m model) renderAddProjectLine(isSelected bool, contentWidth int) string {
	prefix := "  "
	if isSelected {
		prefix = "> "
	}

	if m.ui.Picker.isAdding && isSelected {
		split := splitAtCursor(m.ui.Picker.input.Value(), m.ui.Picker.input.Position())
		return fmt.Sprintf("%s+ %s%s%s",
			prefix,
			split.left,
			ui.GetCursorStyle(m.ui.Screen.WindowFocused).Render(split.cursorCh),
			split.right,
		)
	}

	line := prefix + "+ New Project"
	if isSelected {
		return ui.SelectedStyle.Render(padToWidth(line, contentWidth))
	}
	return ui.MutedStyle.Render(line)
}
