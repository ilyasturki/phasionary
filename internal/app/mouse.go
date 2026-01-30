package app

func (m model) computeRowMap() []int {
	if m.height <= 0 {
		return nil
	}

	rowMap := make([]int, m.height)
	for i := range rowMap {
		rowMap[i] = -1
	}

	layout := m.buildLayout()
	viewport := NewViewport(layout, m.height, DefaultLayoutConfig())
	viewport.ComputeVisibility(m.scrollOffset)

	for i := 0; i < m.height; i++ {
		rowMap[i] = viewport.RowToPosition(i)
	}

	return rowMap
}
