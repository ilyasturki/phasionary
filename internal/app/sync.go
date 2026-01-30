package app

import "phasionary/internal/domain"

func (m *model) storeTaskUpdate() {
	if m.store == nil {
		return
	}
	if err := m.store.SaveProject(m.project); err != nil {
		m.statusMsg = "Save failed: " + err.Error()
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
	m.positions = rebuildPositions(m.project.Categories)
}
