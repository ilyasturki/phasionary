package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"phasionary/internal/domain"
	"phasionary/internal/ui"
)

func renderProjectLine(name string, selected bool) string {
	prefix := "  "
	if selected {
		prefix = "> "
	}
	line := fmt.Sprintf("%s■ %s", prefix, name)
	if selected {
		return ui.SelectedStyle.Render(line)
	}
	return ui.HeaderStyle.Render(line)
}

func (m model) renderEditProjectLine() string {
	prefix := "> "
	icon := "■ "

	edited := m.editValue
	if edited == "" {
		edited = " "
	}
	runes := []rune(edited)
	cursor := m.editCursor
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(runes) {
		cursor = len(runes)
	}
	left := string(runes[:cursor])
	right := string(runes[cursor:])
	cursorChar := " "
	if cursor < len(runes) {
		cursorChar = string(runes[cursor])
		right = string(runes[cursor+1:])
	}
	cursorStyle := ui.SelectedStyle
	return fmt.Sprintf(
		"%s%s%s%s%s",
		prefix,
		ui.HeaderStyle.Render(icon),
		ui.HeaderStyle.Render(left),
		cursorStyle.Render(cursorChar),
		ui.HeaderStyle.Render(right),
	)
}

func renderCategoryLine(name string, selected bool, width int) string {
	prefix := "  "
	if selected {
		prefix = "> "
	}

	if width <= 0 {
		line := fmt.Sprintf("%s%s", prefix, name)
		if selected {
			return ui.SelectedStyle.Render(line)
		}
		return ui.CategoryStyle.Render(line)
	}

	const prefixWidth = 2
	availableWidth := width - prefixWidth
	if availableWidth < 1 {
		availableWidth = 1
	}

	wrapped := ansi.Wrap(name, availableWidth, "")
	lines := strings.Split(wrapped, "\n")
	indent := strings.Repeat(" ", prefixWidth)

	var result []string
	for i, line := range lines {
		if i == 0 {
			result = append(result, prefix+line)
		} else {
			result = append(result, indent+line)
		}
	}

	fullText := strings.Join(result, "\n")
	if selected {
		return ui.SelectedStyle.Render(fullText)
	}
	return ui.CategoryStyle.Render(fullText)
}

func renderTaskLine(task domain.Task, selected bool, width int) string {
	prefix := "  "
	if selected {
		prefix = "> "
	}
	priorityIcon := ui.PriorityIcon(task.Priority)

	if !selected {
		status := formatStatus(task.Status)
		icon := ""
		if priorityIcon != "" {
			icon = ui.PriorityStyle(task.Priority).Render(priorityIcon) + " "
		}
		titleStyle := ui.PriorityStyle(task.Priority)

		if width <= 0 {
			title := titleStyle.Render(task.Title)
			return fmt.Sprintf("%s[%s] %s%s", prefix, status, icon, title)
		}

		prefixPart := fmt.Sprintf("%s[%s] %s", prefix, status, icon)
		overhead := ansi.StringWidth(prefixPart)
		availableWidth := width - overhead
		if availableWidth < 1 {
			availableWidth = 1
		}

		wrapped := ansi.Wrap(task.Title, availableWidth, "")
		wrapLines := strings.Split(wrapped, "\n")
		indent := strings.Repeat(" ", overhead)

		var result []string
		for i, line := range wrapLines {
			styledLine := titleStyle.Render(line)
			if i == 0 {
				result = append(result, fmt.Sprintf("%s[%s] %s%s", prefix, status, icon, styledLine))
			} else {
				result = append(result, indent+styledLine)
			}
		}
		return strings.Join(result, "\n")
	}

	// Selected task
	statusText := statusLabel(task.Status)
	priorityStyle := ui.SelectedPriorityStyle(task.Priority)
	statusStyle := ui.SelectedStatusStyle(task.Status)
	icon := ""
	if priorityIcon != "" {
		icon = priorityStyle.Render(priorityIcon + " ")
	}

	if width <= 0 {
		title := priorityStyle.Render(task.Title)
		return ui.SelectedStyle.Render(prefix+"[") +
			statusStyle.Render(statusText) +
			ui.SelectedStyle.Render("] ") +
			icon +
			title
	}

	// Calculate visible overhead: "> [" + statusText + "] " + icon
	iconText := ""
	if priorityIcon != "" {
		iconText = priorityIcon + " "
	}
	overhead := ansi.StringWidth(prefix + "[" + statusText + "] " + iconText)
	availableWidth := width - overhead
	if availableWidth < 1 {
		availableWidth = 1
	}

	wrapped := ansi.Wrap(task.Title, availableWidth, "")
	wrapLines := strings.Split(wrapped, "\n")
	indent := strings.Repeat(" ", overhead)

	var result []string
	for i, line := range wrapLines {
		styledTitle := priorityStyle.Render(line)
		if i == 0 {
			firstLine := ui.SelectedStyle.Render(prefix+"[") +
				statusStyle.Render(statusText) +
				ui.SelectedStyle.Render("] ") +
				icon +
				styledTitle
			result = append(result, firstLine)
		} else {
			styledIndent := ui.SelectedStyle.Render(indent)
			result = append(result, styledIndent+styledTitle)
		}
	}
	return strings.Join(result, "\n")
}

func statusLabel(status string) string {
	switch status {
	case domain.StatusInProgress:
		return "in progress"
	case domain.StatusCompleted:
		return "completed"
	case domain.StatusCancelled:
		return "cancelled"
	default:
		return "todo"
	}
}

func formatStatus(status string) string {
	return ui.StatusStyle(status).Render(statusLabel(status))
}

func (m model) renderEditCategoryLine() string {
	prefix := "> "
	const prefixWidth = 2

	// Show placeholder when adding a new category with empty value
	if m.addingCategory && m.editValue == "" {
		cursorStyle := ui.SelectedStyle
		placeholder := ui.MutedStyle.Render("Enter category name...")
		styledText := cursorStyle.Render(" ") + placeholder

		if m.width > 0 {
			availableWidth := m.width - prefixWidth
			if availableWidth < 1 {
				availableWidth = 1
			}
			wrapped := ansi.Wrap(styledText, availableWidth, "")
			lines := strings.Split(wrapped, "\n")
			indent := strings.Repeat(" ", prefixWidth)
			for i, line := range lines {
				if i == 0 {
					lines[i] = prefix + line
				} else {
					lines[i] = indent + line
				}
			}
			return strings.Join(lines, "\n")
		}
		return fmt.Sprintf("%s%s", prefix, styledText)
	}

	edited := m.editValue
	if edited == "" {
		edited = " "
	}
	runes := []rune(edited)
	cursor := m.editCursor
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(runes) {
		cursor = len(runes)
	}
	left := string(runes[:cursor])
	right := string(runes[cursor:])
	cursorChar := " "
	if cursor < len(runes) {
		cursorChar = string(runes[cursor])
		right = string(runes[cursor+1:])
	}
	cursorStyle := ui.SelectedStyle

	if m.width > 0 {
		availableWidth := m.width - prefixWidth
		if availableWidth < 1 {
			availableWidth = 1
		}
		wrapped := ansi.Wrap(edited, availableWidth, "")
		wrapLines := strings.Split(wrapped, "\n")
		indent := strings.Repeat(" ", prefixWidth)

		pos := 0
		var result []string
		for i, line := range wrapLines {
			lineRunes := []rune(line)
			lineLen := len(lineRunes)
			var styledLine string

			if cursor >= pos && cursor < pos+lineLen {
				offset := cursor - pos
				l := string(lineRunes[:offset])
				c := string(lineRunes[offset])
				r := string(lineRunes[offset+1:])
				styledLine = ui.CategoryStyle.Render(l) + cursorStyle.Render(c) + ui.CategoryStyle.Render(r)
			} else if cursor == pos+lineLen {
				styledLine = ui.CategoryStyle.Render(line) + cursorStyle.Render(" ")
			} else {
				styledLine = ui.CategoryStyle.Render(line)
			}

			if i == 0 {
				result = append(result, prefix+styledLine)
			} else {
				result = append(result, indent+styledLine)
			}
			pos += lineLen + 1
		}
		return strings.Join(result, "\n")
	}

	return fmt.Sprintf(
		"%s%s%s%s",
		prefix,
		ui.CategoryStyle.Render(left),
		cursorStyle.Render(cursorChar),
		ui.CategoryStyle.Render(right),
	)
}

func (m model) renderEditTaskLine(task domain.Task) string {
	prefix := "> "
	statusText := formatStatus(task.Status)

	titleStyle := ui.PriorityStyle(task.Priority)
	icon := ui.PriorityIcon(task.Priority)
	iconPrefix := ""
	iconPlain := ""
	if icon != "" {
		iconPrefix = titleStyle.Render(icon) + " "
		iconPlain = icon + " "
	}

	// Build the prefix part (everything before the editable text)
	prefixPart := fmt.Sprintf("%s[%s] %s", prefix, statusText, iconPrefix)
	overhead := ansi.StringWidth(prefix + "[" + statusLabel(task.Status) + "] " + iconPlain)

	// Show placeholder text when adding a new task with empty value
	if m.addingTask && m.editValue == "" {
		cursorStyle := ui.SelectedStyle
		placeholder := ui.MutedStyle.Render("Enter task title...")
		styledText := cursorStyle.Render(" ") + placeholder

		if m.width > 0 {
			availableWidth := m.width - overhead
			if availableWidth < 1 {
				availableWidth = 1
			}
			wrapped := ansi.Wrap(styledText, availableWidth, "")
			lines := strings.Split(wrapped, "\n")
			indent := strings.Repeat(" ", overhead)
			for i, line := range lines {
				if i == 0 {
					lines[i] = prefixPart + line
				} else {
					lines[i] = indent + line
				}
			}
			return strings.Join(lines, "\n")
		}
		return prefixPart + styledText
	}

	edited := m.editValue
	if edited == "" {
		edited = " "
	}
	runes := []rune(edited)
	cursor := m.editCursor
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(runes) {
		cursor = len(runes)
	}
	left := string(runes[:cursor])
	right := string(runes[cursor:])
	cursorChar := " "
	if cursor < len(runes) {
		cursorChar = string(runes[cursor])
		right = string(runes[cursor+1:])
	}
	cursorStyle := ui.SelectedStyle

	if m.width > 0 {
		availableWidth := m.width - overhead
		if availableWidth < 1 {
			availableWidth = 1
		}
		wrapped := ansi.Wrap(edited, availableWidth, "")
		wrapLines := strings.Split(wrapped, "\n")
		indent := strings.Repeat(" ", overhead)

		pos := 0
		var result []string
		for i, line := range wrapLines {
			lineRunes := []rune(line)
			lineLen := len(lineRunes)
			var styledLine string

			if cursor >= pos && cursor < pos+lineLen {
				offset := cursor - pos
				l := string(lineRunes[:offset])
				c := string(lineRunes[offset])
				r := string(lineRunes[offset+1:])
				styledLine = titleStyle.Render(l) + cursorStyle.Render(c) + titleStyle.Render(r)
			} else if cursor == pos+lineLen {
				styledLine = titleStyle.Render(line) + cursorStyle.Render(" ")
			} else {
				styledLine = titleStyle.Render(line)
			}

			if i == 0 {
				result = append(result, prefixPart+styledLine)
			} else {
				result = append(result, indent+styledLine)
			}
			pos += lineLen + 1
		}
		return strings.Join(result, "\n")
	}

	return fmt.Sprintf(
		"%s%s%s%s",
		prefixPart,
		titleStyle.Render(left),
		cursorStyle.Render(cursorChar),
		titleStyle.Render(right),
	)
}

func (m model) statusLine() string {
	if m.statusMsg != "" {
		return ui.StatusLineStyle.Render(m.statusMsg)
	}
	position, ok := m.selectedPosition()
	if !ok {
		return ui.StatusLineStyle.Render("No items to display.")
	}
	if position.Kind == focusProject {
		summary := fmt.Sprintf("Project: %s", m.project.Name)
		return ui.StatusLineStyle.Render(summary)
	}
	category := m.categories[position.CategoryIndex]
	if position.Kind == focusCategory {
		summary := fmt.Sprintf("Category: %s (%d tasks)", category.Name, len(category.Tasks))
		return ui.StatusLineStyle.Render(summary)
	}
	task := category.Tasks[position.TaskIndex]
	summary := fmt.Sprintf("Selected: %s / %s (%s)", category.Name, task.Title, task.Status)
	return ui.StatusLineStyle.Render(summary)
}

func (m model) shortcutsLine() string {
	if m.editing {
		return ui.StatusLineStyle.Render("Shortcuts: enter save | esc cancel | arrows move cursor | ? help")
	}
	return ui.StatusLineStyle.Render("Shortcuts: j/k move | J/K reorder | a add task | A add category | enter edit | space status | h/l priority | y copy | d delete | ? help | q quit")
}

func placeOverlay(bg, fg string, width, height int) string {
	bgLines := strings.Split(bg, "\n")
	fgLines := strings.Split(fg, "\n")

	fgW := lipgloss.Width(fg)
	fgH := len(fgLines)
	startY := max(0, (height-fgH)/2)
	startX := max(0, (width-fgW)/2)

	for i, fgLine := range fgLines {
		y := startY + i
		if y >= len(bgLines) {
			break
		}
		left := ansi.Truncate(bgLines[y], startX, "")
		if w := ansi.StringWidth(left); w < startX {
			left += strings.Repeat(" ", startX-w)
		}
		right := ansi.TruncateLeft(bgLines[y], startX+fgW, "")
		bgLines[y] = left + fgLine + right
	}

	return strings.Join(bgLines, "\n")
}

func truncateText(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
}

func (m model) confirmDeleteView() string {
	position, ok := m.selectedPosition()
	if !ok || position.Kind == focusProject {
		return ""
	}
	var message string
	if position.Kind == focusTask {
		task := m.categories[position.CategoryIndex].Tasks[position.TaskIndex]
		message = fmt.Sprintf("Delete task %q?", truncateText(task.Title, 30))
	} else {
		cat := m.categories[position.CategoryIndex]
		message = fmt.Sprintf("Delete category %q and %d tasks?", truncateText(cat.Name, 30), len(cat.Tasks))
	}
	lines := []string{
		message,
		"",
		"y/enter confirm | n/esc cancel",
	}
	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}

func (m model) helpView() string {
	lines := []string{
		"Navigation:",
		"  j/k or ↑/↓    move selection",
		"  ctrl+d/u      half-page down/up",
		"  ctrl+f/b      full-page down/up",
		"  gg            jump to first item",
		"  G             jump to last item",
		"  zz            center selection on screen",
		"",
		"Actions:",
		"  a             add new task",
		"  A             add new category",
		"  enter         edit selected item",
		"  space         toggle task status",
		"  J/K           reorder task up/down",
		"  h/l           change priority",
		"  y             copy selected text",
		"  d             delete selected item",
		"  ?             toggle help",
		"  q or ctrl+c   quit",
		"",
		"Editing:",
		"  enter         save changes",
		"  esc           cancel editing",
		"  ←/→           move cursor",
		"  backspace     delete character",
	}
	return ui.HelpDialogStyle.Render(ui.MutedStyle.Render(strings.Join(lines, "\n")))
}
