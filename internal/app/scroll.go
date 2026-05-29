package app

const wheelScrollStep = 3

func (m *model) scrollUp(amount int) {
	if m.ui.Screen.ScrollOffset <= 0 {
		m.ui.Screen.ScrollOffset = 0
		return
	}
	m.ui.Screen.ScrollOffset -= amount
	if m.ui.Screen.ScrollOffset < 0 {
		m.ui.Screen.ScrollOffset = 0
	}
}

func (m *model) scrollDown(amount int) {
	layout := m.buildLayout()
	viewport := NewViewport(layout, m.ui.Screen.Height, m.layoutConfig())
	for i := 0; i < amount; i++ {
		viewport.ComputeVisibility(m.ui.Screen.ScrollOffset)
		if !viewport.HasMoreBelow {
			return
		}
		m.ui.Screen.ScrollOffset++
	}
}

func (m model) availableHeight() int {
	config := m.layoutConfig()
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
	viewport := NewViewport(layout, m.ui.Screen.Height, m.layoutConfig())
	viewport.ScrollOffset = m.ui.Screen.ScrollOffset
	m.ui.Screen.ScrollOffset = viewport.EnsureVisible(m.selected())
}

func (m *model) centerOnSelected() {
	if m.ui.Selection.IsEmpty() || m.selected() < 0 {
		return
	}

	layout := m.buildLayout()
	viewport := NewViewport(layout, m.ui.Screen.Height, m.layoutConfig())
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
	viewport := NewViewport(layout, m.ui.Screen.Height, m.layoutConfig())
	m.ui.Screen.ScrollOffset = viewport.BottomOnPosition(m.selected())
}
