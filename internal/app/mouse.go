package app

func (m model) computeRowMap() []int {
	if m.ui.Height <= 0 {
		return nil
	}

	rowMap := make([]int, m.ui.Height)
	for i := range rowMap {
		rowMap[i] = -1
	}

	layout := m.buildLayout()
	viewport := NewViewport(layout, m.ui.Height, DefaultLayoutConfig())
	viewport.ComputeVisibility(m.ui.ScrollOffset)

	for i := 0; i < m.ui.Height; i++ {
		rowMap[i] = viewport.RowToPosition(i)
	}

	return rowMap
}
