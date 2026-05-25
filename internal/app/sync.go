package app

import (
	"phasionary/internal/app/selection"
	"phasionary/internal/domain"
)

func (m *model) storeTaskUpdate() {
	if m.deps.Store == nil {
		return
	}
	if err := m.deps.Store.SaveProject(m.project); err != nil {
		m.ui.Screen.StatusMsg = "Save failed: " + err.Error()
	}
}

func rebuildPositions(categories []domain.Category, filter *FilterState, fold *FoldState) []selection.Position {
	positions := make([]selection.Position, 0)
	positions = append(positions, selection.Position{
		Kind:          selection.FocusProject,
		CategoryIndex: -1,
		TaskIndex:     -1,
	})
	for cIndex, category := range categories {
		positions = append(positions, selection.Position{
			Kind:          selection.FocusCategory,
			CategoryIndex: cIndex,
			TaskIndex:     -1,
		})
		if fold != nil && fold.IsFolded(category.ID) {
			continue
		}
		for tIndex, task := range category.Tasks {
			if filter != nil && !filter.TaskVisible(task, category.ID) {
				continue
			}
			positions = append(positions, selection.Position{
				Kind:          selection.FocusTask,
				CategoryIndex: cIndex,
				TaskIndex:     tIndex,
			})
		}
	}
	return positions
}

func (m *model) rebuildPositions() {
	positions := rebuildPositions(m.project.Categories, &m.ui.Filter, &m.ui.Fold)
	m.ui.Selection.SetPositions(positions)
}
