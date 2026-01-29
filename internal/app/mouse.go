package app

import (
	"strings"

	"github.com/charmbracelet/x/ansi"

	"phasionary/internal/domain"
	"phasionary/internal/ui"
)

// computeRowMap builds a mapping from screen row Y coordinates to position indices.
// Returns a slice where rowMap[y] is the position index for that row, or -1 for non-selectable rows.
// This mirrors the View() logic to handle text wrapping, blank lines, and scroll offset.
func (m model) computeRowMap() []int {
	if m.height <= 0 {
		return nil
	}

	rowMap := make([]int, m.height)
	for i := range rowMap {
		rowMap[i] = -1 // default: non-selectable
	}

	availHeight := m.availableHeight()
	row := 0
	usedHeight := 0
	cursor := 0
	hasMoreAbove := m.scrollOffset > 0

	// Reserve space for scroll indicators (must match View() logic)
	if hasMoreAbove {
		availHeight--
	}
	availHeight-- // Reserve for potential "more below"
	if availHeight < 1 {
		availHeight = 1
	}

	// Account for "more above" indicator in row position
	if hasMoreAbove {
		row++
	}

	// Helper to check if we should render an element and map its rows
	// Returns true if rendered, false if we should stop
	renderElement := func(posIndex int, elementHeight int) bool {
		if cursor < m.scrollOffset {
			cursor++
			return true // Skip, continue
		}
		if usedHeight+elementHeight > availHeight {
			return false // Stop rendering
		}
		// Account for separator line between elements (not counted in usedHeight)
		if usedHeight > 0 {
			row++
		}
		// Map all rows of this element to this position index
		for i := 0; i < elementHeight && row+i < m.height; i++ {
			rowMap[row+i] = posIndex
		}
		row += elementHeight
		usedHeight += elementHeight
		cursor++
		return true
	}

	// Helper to add blank lines
	addBlankLines := func(count int) {
		if cursor <= m.scrollOffset || usedHeight == 0 {
			return
		}
		for i := 0; i < count; i++ {
			if usedHeight+1 > availHeight {
				return
			}
			row++
			usedHeight++
		}
	}

	// Project line (first focusable item)
	projectLines := m.countProjectLines()
	if !renderElement(cursor, projectLines) {
		return rowMap
	}
	addBlankLines(2) // 2 blank lines after project

	// Categories and tasks
	for catIdx, category := range m.categories {
		if catIdx > 0 {
			addBlankLines(1) // 1 blank line between categories
		}

		// Category line
		categoryLines := m.countCategoryLines(category.Name)
		if !renderElement(cursor, categoryLines) {
			return rowMap
		}

		if len(category.Tasks) == 0 {
			// "(no tasks)" placeholder - not selectable
			if cursor > m.scrollOffset && usedHeight+1 <= availHeight {
				row++
				usedHeight++
			}
			continue
		}

		addBlankLines(1) // 1 blank line after category header

		// Tasks (consecutive tasks have no blank lines between them)
		for _, task := range category.Tasks {
			taskLines := m.countTaskLines(task)
			if !renderElement(cursor, taskLines) {
				return rowMap
			}
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
