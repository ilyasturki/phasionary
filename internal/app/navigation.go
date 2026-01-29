package app

func (m *model) moveSelection(delta int) {
	if m.editing || len(m.positions) == 0 {
		return
	}
	next := m.selected + delta
	if next < 0 {
		next = 0
	}
	if next >= len(m.positions) {
		next = len(m.positions) - 1
	}
	m.selected = next
	m.ensureVisible()
}

func (m model) selectedPosition() (focusPosition, bool) {
	if m.selected < 0 || m.selected >= len(m.positions) {
		return focusPosition{}, false
	}
	return m.positions[m.selected], true
}

// moveSelectionByPage moves selection by a fraction of the visible height.
// factor=0.5 moves half page, factor=1.0 moves full page.
// Negative values move up, positive move down.
func (m *model) moveSelectionByPage(factor float64) {
	if m.editing || len(m.positions) == 0 {
		return
	}
	pageSize := int(float64(m.availableHeight()) * factor)
	if pageSize == 0 {
		if factor > 0 {
			pageSize = 1
		} else {
			pageSize = -1
		}
	}
	m.moveSelection(pageSize)
}

// jumpToFirst moves selection to the first position (project header).
func (m *model) jumpToFirst() {
	if m.editing || len(m.positions) == 0 {
		return
	}
	m.selected = 0
	m.ensureVisible()
}

// jumpToLast moves selection to the last position.
func (m *model) jumpToLast() {
	if m.editing || len(m.positions) == 0 {
		return
	}
	m.selected = len(m.positions) - 1
	m.ensureVisible()
}
