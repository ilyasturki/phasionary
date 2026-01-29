package app

import (
	"strings"

	"github.com/charmbracelet/x/ansi"

	"phasionary/internal/domain"
	"phasionary/internal/ui"
)

// computeRowMap builds a mapping from screen row Y coordinates to position indices.
// Returns a slice where rowMap[y] is the position index for that row, or -1 for non-selectable rows.
// This mirrors the View() logic to handle text wrapping and blank lines.
func (m model) computeRowMap() []int {
	if m.height <= 0 {
		return nil
	}

	rowMap := make([]int, m.height)
	for i := range rowMap {
		rowMap[i] = -1 // default: non-selectable
	}

	row := 0
	cursor := 0

	// Project line (first focusable item)
	projectLines := m.countProjectLines()
	for i := 0; i < projectLines && row+i < m.height; i++ {
		rowMap[row+i] = cursor
	}
	row += projectLines
	cursor++

	// Blank line after project (2 newlines = 1 blank line)
	row++

	// Categories and tasks
	for catIdx, category := range m.categories {
		// Blank line before category (except first)
		if catIdx > 0 {
			row++
		}

		// Category line
		categoryLines := m.countCategoryLines(category.Name)
		for i := 0; i < categoryLines && row+i < m.height; i++ {
			rowMap[row+i] = cursor
		}
		row += categoryLines
		cursor++

		if len(category.Tasks) == 0 {
			// "(no tasks)" placeholder - not selectable
			row++ // "(no tasks)" line
			continue
		}

		// Tasks
		for _, task := range category.Tasks {
			taskLines := m.countTaskLines(task)
			for i := 0; i < taskLines && row+i < m.height; i++ {
				rowMap[row+i] = cursor
			}
			row += taskLines
			cursor++
		}
	}

	return rowMap
}

// countProjectLines returns how many screen rows the project line takes.
func (m model) countProjectLines() int {
	// Project line is not wrapped in the current implementation
	return 1
}

// countCategoryLines returns how many screen rows a category line takes.
func (m model) countCategoryLines(name string) int {
	if m.width <= 0 {
		return 1
	}

	const prefixWidth = 2
	availableWidth := m.width - prefixWidth
	if availableWidth < 1 {
		availableWidth = 1
	}

	wrapped := ansi.Wrap(name, availableWidth, "")
	return strings.Count(wrapped, "\n") + 1
}

// countTaskLines returns how many screen rows a task line takes.
func (m model) countTaskLines(task domain.Task) int {
	if m.width <= 0 {
		return 1
	}

	prefix := "  "
	priorityIcon := ui.PriorityIcon(task.Priority)
	statusText := statusLabel(task.Status)

	iconText := ""
	if priorityIcon != "" {
		iconText = priorityIcon + " "
	}
	overhead := ansi.StringWidth(prefix + "[" + statusText + "] " + iconText)
	availableWidth := m.width - overhead
	if availableWidth < 1 {
		availableWidth = 1
	}

	wrapped := ansi.Wrap(task.Title, availableWidth, "")
	return strings.Count(wrapped, "\n") + 1
}
