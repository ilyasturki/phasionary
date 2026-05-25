package app

func (m model) availableHeight() int {
	config := DefaultLayoutConfig()
	if m.ui.Screen.Height <= config.FooterHeight {
		return 1
	}
	return m.ui.Screen.Height - config.FooterHeight
}

func (m *model) ensureVisible() {
	if m.ui.Selection.IsEmpty() || m.selected() < 0 {
		m.ui.Screen.ScrollOffset = 0
		return
	}

	layout := m.buildLayout()
	viewport := NewViewport(layout, m.ui.Screen.Height, DefaultLayoutConfig())
	viewport.ScrollOffset = m.ui.Screen.ScrollOffset
	m.ui.Screen.ScrollOffset = viewport.EnsureVisible(m.selected())
}

func (m *model) centerOnSelected() {
	if m.ui.Selection.IsEmpty() || m.selected() < 0 {
		return
	}

	layout := m.buildLayout()
	viewport := NewViewport(layout, m.ui.Screen.Height, DefaultLayoutConfig())
	m.ui.Screen.ScrollOffset = viewport.CenterOnPosition(m.selected())
}

func (m *model) topOnSelected() {
	if m.ui.Selection.IsEmpty() || m.selected() < 0 {
		return
	}
	m.ui.Screen.ScrollOffset = m.selected()
}

func (m *model) bottomOnSelected() {
	if m.ui.Selection.IsEmpty() || m.selected() < 0 {
		return
	}

	layout := m.buildLayout()
	viewport := NewViewport(layout, m.ui.Screen.Height, DefaultLayoutConfig())
	m.ui.Screen.ScrollOffset = viewport.BottomOnPosition(m.selected())
}
