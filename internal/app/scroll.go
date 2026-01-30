package app

func (m model) availableHeight() int {
	config := DefaultLayoutConfig()
	if m.ui.Height <= config.FooterHeight {
		return 1
	}
	return m.ui.Height - config.FooterHeight
}

func (m *model) ensureVisible() {
	if m.ui.Selection.IsEmpty() || m.selected() < 0 {
		m.ui.ScrollOffset = 0
		return
	}

	layout := m.buildLayout()
	viewport := NewViewport(layout, m.ui.Height, DefaultLayoutConfig())
	viewport.ScrollOffset = m.ui.ScrollOffset
	m.ui.ScrollOffset = viewport.EnsureVisible(m.selected())
}

func (m *model) centerOnSelected() {
	if m.ui.Selection.IsEmpty() || m.selected() < 0 {
		return
	}

	layout := m.buildLayout()
	viewport := NewViewport(layout, m.ui.Height, DefaultLayoutConfig())
	m.ui.ScrollOffset = viewport.CenterOnPosition(m.selected())
}
