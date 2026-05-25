package app

func (m model) computeRowMap() []int {
	if m.ui.Screen.Height <= 0 {
		return nil
	}

	rowMap := make([]int, m.ui.Screen.Height)
	for i := range rowMap {
		rowMap[i] = -1
	}

	layout := m.buildLayout()
	viewport := NewViewport(layout, m.ui.Screen.Height, m.layoutConfig())
	viewport.ComputeVisibility(m.ui.Screen.ScrollOffset)

	for i := 0; i < m.ui.Screen.Height; i++ {
		rowMap[i] = viewport.RowToPosition(i)
	}

	return rowMap
}
