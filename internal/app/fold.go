package app

import "phasionary/internal/app/selection"

func (m *model) toggleFold() {
	pos, ok := m.selectedPosition()
	if !ok {
		return
	}

	var categoryID string
	switch pos.Kind {
	case selection.FocusCategory:
		categoryID = m.project.Categories[pos.CategoryIndex].ID
	case selection.FocusTask:
		categoryID = m.project.Categories[pos.CategoryIndex].ID
		m.ui.Selection.SetSelected(m.findCategoryPositionIndex(pos.CategoryIndex))
	default:
		return
	}

	m.ui.Fold.Toggle(categoryID)
	m.saveFoldState()
	m.rebuildPositions()
	m.ensureVisible()
}

func (m *model) foldAll() {
	categoryIDs := make([]string, len(m.project.Categories))
	for i, cat := range m.project.Categories {
		categoryIDs[i] = cat.ID
	}
	m.ui.Fold.FoldAll(categoryIDs)
	m.saveFoldState()

	pos, ok := m.selectedPosition()
	if ok && pos.Kind == selection.FocusTask {
		m.ui.Selection.SetSelected(m.findCategoryPositionIndex(pos.CategoryIndex))
	}

	m.rebuildPositions()
	m.ensureVisible()
}

func (m *model) unfoldAll() {
	m.ui.Fold.UnfoldAll()
	m.saveFoldState()
	m.rebuildPositions()
	m.ensureVisible()
}

func (m *model) saveFoldState() {
	_ = m.deps.StateManager.SetFoldedCategories(m.project.ID, m.ui.Fold.FoldedIDs())
}

func (m *model) findCategoryPositionIndex(categoryIndex int) int {
	for i, pos := range m.positions() {
		if pos.Kind == selection.FocusCategory && pos.CategoryIndex == categoryIndex {
			return i
		}
	}
	return 0
}
