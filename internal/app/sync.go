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

func (m *model) refreshtaskview(position focusPosition) {
	if position.CategoryIndex < 0 || position.CategoryIndex >= len(m.categories) {
		return
	}
	category := &m.categories[position.CategoryIndex]
	sorted := append([]domain.Task(nil), category.Tasks...)
	domain.SortTasks(sorted)
	category.Tasks = sorted
	m.positions = rebuildPositions(m.categories)
	m.selected = m.findPositionForTask(position, category.Tasks)
}

func rebuildPositions(categories []categoryView) []focusPosition {
	positions := make([]focusPosition, 0)
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

func (m *model) findPositionForTask(previous focusPosition, tasks []domain.Task) int {
	if previous.CategoryIndex < 0 || previous.CategoryIndex >= len(m.categories) {
		return m.selected
	}
	if previous.TaskIndex < 0 || previous.TaskIndex >= len(m.categories[previous.CategoryIndex].Tasks) {
		return m.selected
	}
	taskID := m.categories[previous.CategoryIndex].Tasks[previous.TaskIndex].ID
	for index, position := range m.positions {
		if position.Kind == focusTask &&
			position.CategoryIndex == previous.CategoryIndex &&
			position.TaskIndex >= 0 &&
			position.TaskIndex < len(tasks) &&
			tasks[position.TaskIndex].ID == taskID {
			return index
		}
	}
	return m.selected
}
