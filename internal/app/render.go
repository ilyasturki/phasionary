package app

import (
	"fmt"
	"strings"

	"phasionary/internal/domain"
	"phasionary/internal/ui"
)

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
	return ui.StatusLineStyle.Render("Shortcuts: j/k move | a add task | A add category | enter edit | space status | h/l priority | y copy | ? help | q quit")
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
