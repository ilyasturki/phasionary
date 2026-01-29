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
