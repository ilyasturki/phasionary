package app

import "phasionary/internal/domain"

func (m *model) deleteSelected() {
	if m.editing {
		return
	}
	pos, ok := m.selectedPosition()
	if !ok || pos.Kind == focusProject {
		return
	}
	m.confirmDelete = true
}

func (m *model) confirmDeleteAction() {
	m.confirmDelete = false
	position, ok := m.selectedPosition()
	if !ok {
		return
	}
	switch position.Kind {
	case focusTask:
		m.deleteTask(position)
	case focusCategory:
		m.deleteCategory(position)
	}
}

func (m *model) deleteTask(position focusPosition) {
	catIndex := position.CategoryIndex
	taskIndex := position.TaskIndex
	taskID := m.categories[catIndex].Tasks[taskIndex].ID

	// Remove from view categories
	m.categories[catIndex].Tasks = append(
		m.categories[catIndex].Tasks[:taskIndex],
		m.categories[catIndex].Tasks[taskIndex+1:]...,
	)

	// Remove from project categories (match by ID)
	projTasks := m.project.Categories[catIndex].Tasks
	for i, t := range projTasks {
		if t.ID == taskID {
			m.project.Categories[catIndex].Tasks = append(projTasks[:i], projTasks[i+1:]...)
			break
		}
	}

	m.rebuildAndClamp()
	m.storeTaskUpdate()
}

func (m *model) deleteCategory(position focusPosition) {
	catIndex := position.CategoryIndex

	// Remove from view categories
	m.categories = append(m.categories[:catIndex], m.categories[catIndex+1:]...)

	// Remove from project categories
	m.project.Categories = append(m.project.Categories[:catIndex], m.project.Categories[catIndex+1:]...)

	m.rebuildAndClamp()
	m.storeTaskUpdate()
}

func (m *model) rebuildAndClamp() {
	m.positions = rebuildPositions(m.categories)
	if m.selected >= len(m.positions) {
		m.selected = len(m.positions) - 1
	}
	if m.selected < 0 && len(m.positions) > 0 {
		m.selected = 0
	}
	m.ensureVisible()
}

func (m *model) toggleSelectedTask() {
	if m.editing {
		return
	}
	position, ok := m.selectedPosition()
	if !ok || position.Kind != focusTask {
		return
	}
	category := &m.categories[position.CategoryIndex]
	task := &category.Tasks[position.TaskIndex]
	nextStatus := nextTaskStatus(task.Status)
	if nextStatus == task.Status {
		return
	}
	updateTaskStatus(task, nextStatus)
	m.syncTaskToProject(position, *task)
	m.storeTaskUpdate()
}

func nextTaskStatus(current string) string {
	switch current {
	case domain.StatusTodo:
		return domain.StatusInProgress
	case domain.StatusInProgress:
		return domain.StatusCompleted
	case domain.StatusCompleted:
		return domain.StatusTodo
	case domain.StatusCancelled:
		return domain.StatusTodo
	default:
		return domain.StatusTodo
	}
}

func updateTaskStatus(task *domain.Task, status string) {
	task.Status = status
	task.UpdatedAt = domain.NowTimestamp()
	if status == domain.StatusCompleted {
		task.CompletionDate = domain.NowTimestamp()
		return
	}
	task.CompletionDate = ""
}

func (m *model) increasePriority() {
	if m.editing {
		return
	}
	position, ok := m.selectedPosition()
	if !ok || position.Kind != focusTask {
		return
	}
	category := &m.categories[position.CategoryIndex]
	task := &category.Tasks[position.TaskIndex]
	newPriority := nextPriorityUp(task.Priority)
	if newPriority == task.Priority {
		return
	}
	task.Priority = newPriority
	task.UpdatedAt = domain.NowTimestamp()
	m.syncTaskToProject(position, *task)
	m.storeTaskUpdate()
}

func (m *model) decreasePriority() {
	if m.editing {
		return
	}
	position, ok := m.selectedPosition()
	if !ok || position.Kind != focusTask {
		return
	}
	category := &m.categories[position.CategoryIndex]
	task := &category.Tasks[position.TaskIndex]
	newPriority := nextPriorityDown(task.Priority)
	if newPriority == task.Priority {
		return
	}
	task.Priority = newPriority
	task.UpdatedAt = domain.NowTimestamp()
	m.syncTaskToProject(position, *task)
	m.storeTaskUpdate()
}

func (m *model) moveTaskDown() {
	if m.editing {
		return
	}
	position, ok := m.selectedPosition()
	if !ok || position.Kind != focusTask {
		return
	}
	catIndex := position.CategoryIndex
	taskIndex := position.TaskIndex
	tasks := m.categories[catIndex].Tasks
	if taskIndex >= len(tasks)-1 {
		return
	}
	// Swap in view
	tasks[taskIndex], tasks[taskIndex+1] = tasks[taskIndex+1], tasks[taskIndex]
	// Swap in project model
	pt := m.project.Categories[catIndex].Tasks
	pt[taskIndex], pt[taskIndex+1] = pt[taskIndex+1], pt[taskIndex]
	// Rebuild positions and follow the moved task
	m.positions = rebuildPositions(m.categories)
	m.selected++
	m.ensureVisible()
	m.storeTaskUpdate()
}

func (m *model) moveTaskUp() {
	if m.editing {
		return
	}
	position, ok := m.selectedPosition()
	if !ok || position.Kind != focusTask {
		return
	}
	catIndex := position.CategoryIndex
	taskIndex := position.TaskIndex
	if taskIndex <= 0 {
		return
	}
	tasks := m.categories[catIndex].Tasks
	// Swap in view
	tasks[taskIndex], tasks[taskIndex-1] = tasks[taskIndex-1], tasks[taskIndex]
	// Swap in project model
	pt := m.project.Categories[catIndex].Tasks
	pt[taskIndex], pt[taskIndex-1] = pt[taskIndex-1], pt[taskIndex]
	// Rebuild positions and follow the moved task
	m.positions = rebuildPositions(m.categories)
	m.selected--
	m.ensureVisible()
	m.storeTaskUpdate()
}

func nextPriorityUp(current string) string {
	switch current {
	case domain.PriorityLow:
		return domain.PriorityMedium
	case domain.PriorityMedium:
		return domain.PriorityHigh
	case domain.PriorityHigh:
		return domain.PriorityHigh
	default:
		return domain.PriorityMedium
	}
}

func nextPriorityDown(current string) string {
	switch current {
	case domain.PriorityHigh:
		return domain.PriorityMedium
	case domain.PriorityMedium:
		return domain.PriorityLow
	case domain.PriorityLow:
		return domain.PriorityLow
	default:
		return domain.PriorityMedium
	}
}
