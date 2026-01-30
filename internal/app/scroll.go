package app

func (m model) availableHeight() int {
	config := DefaultLayoutConfig()
	if m.height <= config.FooterHeight {
		return 1
	}
	return m.height - config.FooterHeight
}

func (m *model) ensureVisible() {
	if len(m.positions) == 0 || m.selected < 0 {
		m.scrollOffset = 0
		return
	}

	if m.selected >= len(m.positions) {
		m.selected = len(m.positions) - 1
	}

	layout := m.buildLayout()
	viewport := NewViewport(layout, m.height, DefaultLayoutConfig())
	viewport.ScrollOffset = m.scrollOffset
	m.scrollOffset = viewport.EnsureVisible(m.selected)
}

func (m *model) centerOnSelected() {
	if len(m.positions) == 0 || m.selected < 0 {
		return
	}

	layout := m.buildLayout()
	viewport := NewViewport(layout, m.height, DefaultLayoutConfig())
	m.scrollOffset = viewport.CenterOnPosition(m.selected)
}
