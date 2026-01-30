package app

import "phasionary/internal/domain"

func (m *model) deleteSelected() {
	if m.mode == ModeEdit {
		return
	}
	pos, ok := m.selectedPosition()
	if !ok || pos.Kind == focusProject {
		return
	}
	m.mode = ModeConfirmDelete
}

func (m *model) confirmDeleteAction() {
	m.mode = ModeNormal
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
	tasks := m.project.Categories[catIndex].Tasks
	m.project.Categories[catIndex].Tasks = append(tasks[:taskIndex], tasks[taskIndex+1:]...)
	m.rebuildAndClamp()
	m.storeTaskUpdate()
}

func (m *model) deleteCategory(position focusPosition) {
	catIndex := position.CategoryIndex
	m.project.Categories = append(m.project.Categories[:catIndex], m.project.Categories[catIndex+1:]...)
	m.rebuildAndClamp()
	m.storeTaskUpdate()
}

func (m *model) rebuildAndClamp() {
	m.rebuildPositions()
	if m.selected >= len(m.positions) {
		m.selected = len(m.positions) - 1
	}
	if m.selected < 0 && len(m.positions) > 0 {
		m.selected = 0
	}
	m.ensureVisible()
}

func (m *model) toggleSelectedTask() {
	if m.mode == ModeEdit {
		return
	}
	position, ok := m.selectedPosition()
	if !ok || position.Kind != focusTask {
		return
	}
	task := &m.project.Categories[position.CategoryIndex].Tasks[position.TaskIndex]
	nextStatus := nextTaskStatus(task.Status)
	if nextStatus == task.Status {
		return
	}
	updateTaskStatus(task, nextStatus)
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
	_ = task.SetStatus(status)
}

func (m *model) increasePriority() {
	if m.mode == ModeEdit {
		return
	}
	position, ok := m.selectedPosition()
	if !ok || position.Kind != focusTask {
		return
	}
	task := &m.project.Categories[position.CategoryIndex].Tasks[position.TaskIndex]
	newPriority := nextPriorityUp(task.Priority)
	if newPriority == task.Priority {
		return
	}
	task.Priority = newPriority
	task.UpdatedAt = domain.NowTimestamp()
	m.storeTaskUpdate()
}

func (m *model) decreasePriority() {
	if m.mode == ModeEdit {
		return
	}
	position, ok := m.selectedPosition()
	if !ok || position.Kind != focusTask {
		return
	}
	task := &m.project.Categories[position.CategoryIndex].Tasks[position.TaskIndex]
	newPriority := nextPriorityDown(task.Priority)
	if newPriority == task.Priority {
		return
	}
	task.Priority = newPriority
	task.UpdatedAt = domain.NowTimestamp()
	m.storeTaskUpdate()
}

func (m *model) moveTaskDown() {
	if m.mode == ModeEdit {
		return
	}
	position, ok := m.selectedPosition()
	if !ok || position.Kind != focusTask {
		return
	}
	catIndex := position.CategoryIndex
	taskIndex := position.TaskIndex
	tasks := m.project.Categories[catIndex].Tasks
	if taskIndex >= len(tasks)-1 {
		return
	}
	tasks[taskIndex], tasks[taskIndex+1] = tasks[taskIndex+1], tasks[taskIndex]
	m.rebuildPositions()
	m.selected++
	m.ensureVisible()
	m.storeTaskUpdate()
}

func (m *model) moveTaskUp() {
	if m.mode == ModeEdit {
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
	tasks := m.project.Categories[catIndex].Tasks
	tasks[taskIndex], tasks[taskIndex-1] = tasks[taskIndex-1], tasks[taskIndex]
	m.rebuildPositions()
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
