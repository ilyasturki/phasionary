package app

import "phasionary/internal/domain"

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
		task.Section = domain.SectionPast
		return
	}
	if status == domain.StatusCancelled {
		task.Section = domain.SectionPast
		return
	}
	if task.Section == domain.SectionPast {
		task.Section = domain.SectionCurrent
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
	m.refreshtaskview(position)
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
	m.refreshtaskview(position)
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
