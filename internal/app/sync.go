package app

import "phasionary/internal/domain"

func (m *model) syncTaskToProject(position focusPosition, task domain.Task) {
	if position.CategoryIndex < 0 || position.CategoryIndex >= len(m.project.Categories) {
		return
	}
	projectCategory := &m.project.Categories[position.CategoryIndex]
	for index := range projectCategory.Tasks {
		if projectCategory.Tasks[index].ID == task.ID {
			projectCategory.Tasks[index] = task
			return
		}
	}
	projectCategory.Tasks = append(projectCategory.Tasks, task)
}

func (m *model) storeTaskUpdate() {
	if m.store == nil {
		return
	}
	_ = m.store.SaveProject(m.project)
}

func rebuildPositions(categories []categoryView) []focusPosition {
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

