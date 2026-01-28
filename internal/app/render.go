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

func renderCategoryLine(name string, selected bool) string {
	prefix := "  "
	if selected {
		prefix = "> "
	}
	line := fmt.Sprintf("%s%s", prefix, name)
	if selected {
		return ui.SelectedStyle.Render(line)
	}
	return ui.CategoryStyle.Render(line)
}

func renderTaskLine(task domain.Task, selected bool) string {
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
		title := titleStyle.Render(task.Title)
		return fmt.Sprintf("%s[%s] %s%s", prefix, status, icon, title)
	}
	statusText := statusLabel(task.Status)
	priorityStyle := ui.SelectedPriorityStyle(task.Priority)
	statusStyle := ui.SelectedStatusStyle(task.Status)
	icon := ""
	if priorityIcon != "" {
		icon = priorityStyle.Render(priorityIcon + " ")
	}
	title := priorityStyle.Render(task.Title)
	return ui.SelectedStyle.Render(prefix+"[") +
		statusStyle.Render(statusText) +
		ui.SelectedStyle.Render("] ") +
		icon +
		title
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

	// Show placeholder when adding a new category with empty value
	if m.addingCategory && m.editValue == "" {
		placeholder := ui.MutedStyle.Render("Enter category name...")
		cursorStyle := ui.SelectedStyle
		return fmt.Sprintf("%s%s%s", prefix, cursorStyle.Render(" "), placeholder)
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

	// Show placeholder text when adding a new task with empty value
	if m.addingTask && m.editValue == "" {
		placeholder := ui.MutedStyle.Render("Enter task title...")
		cursorStyle := ui.SelectedStyle
		return fmt.Sprintf(
			"%s[%s] %s%s",
			prefix,
			statusText,
			cursorStyle.Render(" "),
			placeholder,
		)
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
	titleStyle := ui.PriorityStyle(task.Priority)
	icon := ui.PriorityIcon(task.Priority)
	iconPrefix := ""
	if icon != "" {
		iconPrefix = titleStyle.Render(icon) + " "
	}
	return fmt.Sprintf(
		"%s[%s] %s%s%s%s",
		prefix,
		statusText,
		iconPrefix,
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
	summary := fmt.Sprintf("Selected: %s / %s (%s - %s)", category.Name, task.Title, task.Status, task.Section)
	return ui.StatusLineStyle.Render(summary)
}

func (m model) shortcutsLine() string {
	if m.editing {
		return ui.StatusLineStyle.Render("Shortcuts: enter save | esc cancel | arrows move cursor | ? help")
	}
	return ui.StatusLineStyle.Render("Shortcuts: j/k move | a add task | A add category | enter edit | space status | h/l priority | y copy | d delete | ? help | q quit")
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
		"Shortcuts:",
		"  ? toggle help  (q/esc to close)",
		"  up/down or j/k move selection",
		"  a add new task",
		"  A add new category",
		"  enter edit selected item",
		"  space toggle task status",
		"  h/l change priority",
		"  y copy selected text",
		"  d delete selected item",
		"  q or ctrl+c quit",
		"",
		"Editing:",
		"  enter save changes",
		"  esc cancel editing",
		"  left/right move cursor",
		"  backspace/delete remove character",
		"  type to insert text",
	}
	return ui.HelpDialogStyle.Render(ui.MutedStyle.Render(strings.Join(lines, "\n")))
}
