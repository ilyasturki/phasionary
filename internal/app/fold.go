package app

func (m *model) toggleFold() {
	pos, ok := m.selectedPosition()
	if !ok {
		return
	}

	var categoryID string
	switch pos.Kind {
	case focusCategory:
		categoryID = m.project.Categories[pos.CategoryIndex].ID
	case focusTask:
		categoryID = m.project.Categories[pos.CategoryIndex].ID
		m.ui.Selection.SetSelected(m.findCategoryPositionIndex(pos.CategoryIndex))
	default:
		return
	}

	m.ui.Fold.Toggle(categoryID)
	m.rebuildPositions()
	m.ensureVisible()
}

func (m *model) foldAll() {
	categoryIDs := make([]string, len(m.project.Categories))
	for i, cat := range m.project.Categories {
		categoryIDs[i] = cat.ID
	}
	m.ui.Fold.FoldAll(categoryIDs)

	pos, ok := m.selectedPosition()
	if ok && pos.Kind == focusTask {
		m.ui.Selection.SetSelected(m.findCategoryPositionIndex(pos.CategoryIndex))
	}

	m.rebuildPositions()
	m.ensureVisible()
}

func (m *model) unfoldAll() {
	m.ui.Fold.UnfoldAll()
	m.rebuildPositions()
	m.ensureVisible()
}

func (m *model) findCategoryPositionIndex(categoryIndex int) int {
	for i, pos := range m.positions() {
		if pos.Kind == focusCategory && pos.CategoryIndex == categoryIndex {
			return i
		}
	}
	return 0
}
