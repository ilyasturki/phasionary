package app

import (
	"fmt"
	"strings"

	"phasionary/internal/ui"
)

func (m model) projectPickerView() string {
	lines := []string{ui.DialogTitleStyle.Render("Select Project:"), ""}

	total := m.ui.Picker.totalItems()
	visibleEnd := m.ui.Picker.scrollOffset + pickerVisibleItems
	if visibleEnd > total {
		visibleEnd = total
	}

	if m.ui.Picker.scrollOffset > 0 {
		lines = append(lines, ui.DialogHintStyle.Render("  ↑ more above"))
	}

	for i := m.ui.Picker.scrollOffset; i < visibleEnd; i++ {
		isSelected := i == m.ui.Picker.selected
		if i == len(m.ui.Picker.projects) {
			lines = append(lines, m.renderAddProjectLine(isSelected))
			continue
		}
		p := m.ui.Picker.projects[i]
		prefix := "  "
		if isSelected {
			prefix = "> "
		}
		name := p.Name
		if p.ID == m.project.ID {
			name += " (current)"
		}
		line := prefix + name
		if isSelected {
			line = ui.SelectedStyle.Render(line)
		} else if p.ID == m.project.ID {
			line = prefix + p.Name + ui.DialogHintStyle.Render(" (current)")
		}
		lines = append(lines, line)
	}

	if visibleEnd < total {
		lines = append(lines, ui.DialogHintStyle.Render("  ↓ more below"))
	}

	hintText := "j/k navigate | J/K reorder | enter select | d delete | esc cancel"
	if m.ui.Picker.isAdding {
		hintText = "enter create | esc cancel"
	}
	lines = append(lines, "", ui.DialogHintStyle.Render(hintText))

	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}

func (m model) renderAddProjectLine(isSelected bool) string {
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
		return ui.SelectedStyle.Render(line)
	}
	return ui.MutedStyle.Render(line)
}
