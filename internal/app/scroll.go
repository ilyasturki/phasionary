package app

// footerHeight is the number of lines used by footer:
// 1 blank line (\n\n, where first \n ends body) + status line + shortcuts line = 3
const footerHeight = 3

// availableHeight returns the number of lines available for content rendering
func (m model) availableHeight() int {
	if m.height <= footerHeight {
		return 1
	}
	return m.height - footerHeight
}

// elementHeight returns the rendered line count for a given position index
func (m model) elementHeight(posIndex int) int {
	if posIndex < 0 || posIndex >= len(m.positions) {
		return 1
	}
	pos := m.positions[posIndex]
	switch pos.Kind {
	case focusProject:
		return m.countProjectLines()
	case focusCategory:
		return m.countCategoryLines(m.categories[pos.CategoryIndex].Name)
	case focusTask:
		return m.countTaskLines(m.categories[pos.CategoryIndex].Tasks[pos.TaskIndex])
	default:
		return 1
	}
}

// ensureVisible adjusts scrollOffset so that the selected element is visible
func (m *model) ensureVisible() {
	if len(m.positions) == 0 || m.selected < 0 {
		m.scrollOffset = 0
		return
	}

	// Clamp selected to valid range
	if m.selected >= len(m.positions) {
		m.selected = len(m.positions) - 1
	}

	// If selected is above scroll offset, scroll up to show it
	if m.selected < m.scrollOffset {
		m.scrollOffset = m.selected
		return
	}

	// Calculate if selected is below visible area
	availHeight := m.availableHeight()

	// Reserve space for scroll indicators (must match View() logic)
	if m.scrollOffset > 0 {
		availHeight-- // "more above" indicator
	}
	availHeight-- // "more below" indicator
	if availHeight < 1 {
		availHeight = 1
	}

	usedHeight := 0

	// Calculate height from scrollOffset to selected (inclusive)
	// Each element occupies elementHeight visual rows.
	// spacingBefore adds blank lines between elements.
	for i := m.scrollOffset; i <= m.selected; i++ {
		h := m.elementHeight(i)
		// Add inter-element spacing (blank lines)
		if i > m.scrollOffset {
			h += m.spacingBefore(i)
		}
		usedHeight += h
	}

	// If selected extends below visible area, scroll down
	for usedHeight > availHeight && m.scrollOffset < m.selected {
		// Remove the topmost element and its following spacing
		removeHeight := m.elementHeight(m.scrollOffset)
		if m.scrollOffset+1 <= m.selected {
			removeHeight += m.spacingBefore(m.scrollOffset + 1)
		}
		usedHeight -= removeHeight
		m.scrollOffset++

		// After scrolling, we now have a "more above" indicator
		// Only reduce availHeight once when we first start scrolling
		if m.scrollOffset == 1 {
			availHeight--
			if availHeight < 1 {
				availHeight = 1
			}
		}
	}
}

// spacingBefore returns blank lines before a position (for layout spacing).
// This must match the addBlankLines() calls in View().
func (m model) spacingBefore(posIndex int) int {
	if posIndex <= 0 || posIndex >= len(m.positions) {
		return 0
	}
	pos := m.positions[posIndex]
	prevPos := m.positions[posIndex-1]

	switch pos.Kind {
	case focusProject:
		return 0
	case focusCategory:
		// After project: 2 blank lines
		if prevPos.Kind == focusProject {
			return 2
		}
		// Between categories: 1 blank line
		return 1
	case focusTask:
		if prevPos.Kind == focusCategory {
			// First task after category header: 1 blank line
			return 1
		}
		// Consecutive tasks: no blank lines between them
		return 0
	}
	return 0
}
