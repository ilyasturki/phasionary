package app

import "phasionary/internal/domain"

func (m *model) storeTaskUpdate() {
	if m.deps.Store == nil {
		return
	}
	if err := m.deps.Store.SaveProject(m.project); err != nil {
		m.ui.StatusMsg = "Save failed: " + err.Error()
	}
}

func rebuildPositions(categories []domain.Category) []focusPosition {
	positions := make([]focusPosition, 0)
	positions = append(positions, focusPosition{
		Kind:          focusProject,
		CategoryIndex: -1,
		TaskIndex:     -1,
	})
	for cIndex, category := range categories {
		positions = append(positions, focusPosition{
			Kind:          focusCategory,
			CategoryIndex: cIndex,
			TaskIndex:     -1,
		})
		for tIndex := range category.Tasks {
			positions = append(positions, focusPosition{
				Kind:          focusTask,
				CategoryIndex: cIndex,
				TaskIndex:     tIndex,
			})
		}
	}
	return positions
}

func (m *model) rebuildPositions() {
	positions := rebuildPositions(m.project.Categories)
	m.ui.Selection.SetPositions(toSelectionPositions(positions))
}
