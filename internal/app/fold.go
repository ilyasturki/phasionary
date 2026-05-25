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

func (m *model) toggleExpandDescriptions() {
	// If we're collapsing while focused on a description row, that row is about
	// to disappear — remember the parent task so we can land back on it instead
	// of wherever the index happens to fall.
	var (
		wasOnDescription bool
		parentTaskID     string
	)
	if pos, ok := m.selectedPosition(); ok && pos.Kind == selection.FocusDescription {
		if pos.CategoryIndex >= 0 && pos.CategoryIndex < len(m.project.Categories) {
			cat := m.project.Categories[pos.CategoryIndex]
			if pos.TaskIndex >= 0 && pos.TaskIndex < len(cat.Tasks) {
				wasOnDescription = true
				parentTaskID = cat.Tasks[pos.TaskIndex].ID
			}
		}
	}

	m.ui.Screen.ExpandDescriptions = !m.ui.Screen.ExpandDescriptions
	m.rebuildPositions()

	if wasOnDescription && !m.ui.Screen.ExpandDescriptions && parentTaskID != "" {
		m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
			if p.Kind != selection.FocusTask || p.CategoryIndex < 0 || p.CategoryIndex >= len(m.project.Categories) {
				return false
			}
			c := m.project.Categories[p.CategoryIndex]
			return p.TaskIndex >= 0 && p.TaskIndex < len(c.Tasks) && c.Tasks[p.TaskIndex].ID == parentTaskID
		})
	}

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
