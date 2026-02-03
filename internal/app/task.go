package app

import (
	"sort"

	"phasionary/internal/app/components"
	"phasionary/internal/app/modes"
	"phasionary/internal/app/selection"
	"phasionary/internal/domain"
)

func (m *model) deleteSelected() {
	if !m.ui.Modes.CanPerformAction(modes.ActionDeleteItem) {
		return
	}
	pos, ok := m.selectedPosition()
	if !ok || pos.Kind == focusProject {
		return
	}
	m.ui.Modes.ToConfirmDelete()
}

func (m *model) confirmDeleteAction() {
	m.ui.Modes.ToNormal()
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
	_ = m.project.Categories[catIndex].RemoveTask(taskIndex)
	m.rebuildAndClamp()
	m.storeTaskUpdate()
}

func (m *model) deleteCategory(position focusPosition) {
	catIndex := position.CategoryIndex
	_ = m.project.RemoveCategory(catIndex)
	m.rebuildAndClamp()
	m.storeTaskUpdate()
}

func (m *model) rebuildAndClamp() {
	m.rebuildPositions()
	m.ensureVisible()
}

func (m *model) toggleSelectedTask() {
	if !m.ui.Modes.CanPerformAction(modes.ActionToggleTask) {
		return
	}
	position, ok := m.selectedPosition()
	if !ok || position.Kind != focusTask {
		return
	}
	task := &m.project.Categories[position.CategoryIndex].Tasks[position.TaskIndex]
	if task.CycleStatus() {
		m.storeTaskUpdate()
	}
}

func (m *model) increasePriority() {
	if !m.ui.Modes.CanPerformAction(modes.ActionChangePriority) {
		return
	}
	position, ok := m.selectedPosition()
	if !ok || position.Kind != focusTask {
		return
	}
	task := &m.project.Categories[position.CategoryIndex].Tasks[position.TaskIndex]
	if task.IncreasePriority() {
		m.storeTaskUpdate()
	}
}

func (m *model) decreasePriority() {
	if !m.ui.Modes.CanPerformAction(modes.ActionChangePriority) {
		return
	}
	position, ok := m.selectedPosition()
	if !ok || position.Kind != focusTask {
		return
	}
	task := &m.project.Categories[position.CategoryIndex].Tasks[position.TaskIndex]
	if task.DecreasePriority() {
		m.storeTaskUpdate()
	}
}

func (m *model) openEstimatePicker() {
	if !m.ui.Modes.CanPerformAction(modes.ActionChangeEstimate) {
		return
	}
	position, ok := m.selectedPosition()
	if !ok || position.Kind == focusProject {
		return
	}

	var currentEstimate int
	if position.Kind == focusTask {
		currentEstimate = m.project.Categories[position.CategoryIndex].Tasks[position.TaskIndex].EstimateMinutes
	} else {
		currentEstimate = m.project.Categories[position.CategoryIndex].EstimateMinutes
	}

	m.ui.EstimatePicker = components.NewEstimatePickerState(currentEstimate)
	m.ui.Modes.ToEstimatePicker()
}

func (m *model) selectEstimate(minutes int) {
	position, ok := m.selectedPosition()
	if !ok || position.Kind == focusProject {
		return
	}

	if position.Kind == focusTask {
		task := &m.project.Categories[position.CategoryIndex].Tasks[position.TaskIndex]
		task.SetEstimate(minutes)
	} else {
		category := &m.project.Categories[position.CategoryIndex]
		category.SetEstimate(minutes)
	}
	m.storeTaskUpdate()
}

func (m *model) moveTaskDown() {
	if !m.ui.Modes.CanPerformAction(modes.ActionMoveItem) {
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
	m.ui.Selection.MoveBy(1)
	m.ensureVisible()
	m.storeTaskUpdate()
}

func (m *model) moveTaskUp() {
	if !m.ui.Modes.CanPerformAction(modes.ActionMoveItem) {
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
	m.ui.Selection.MoveBy(-1)
	m.ensureVisible()
	m.storeTaskUpdate()
}

func (m *model) moveCategoryDown() {
	if !m.ui.Modes.CanPerformAction(modes.ActionMoveItem) {
		return
	}
	position, ok := m.selectedPosition()
	if !ok || position.Kind != focusCategory {
		return
	}
	catIndex := position.CategoryIndex
	if catIndex >= len(m.project.Categories)-1 {
		return
	}
	m.project.Categories[catIndex], m.project.Categories[catIndex+1] =
		m.project.Categories[catIndex+1], m.project.Categories[catIndex]
	m.rebuildPositions()
	m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
		return p.Kind == selection.FocusCategory && p.CategoryIndex == catIndex+1
	})
	m.ensureVisible()
	m.storeTaskUpdate()
}

func (m *model) moveCategoryUp() {
	if !m.ui.Modes.CanPerformAction(modes.ActionMoveItem) {
		return
	}
	position, ok := m.selectedPosition()
	if !ok || position.Kind != focusCategory {
		return
	}
	catIndex := position.CategoryIndex
	if catIndex <= 0 {
		return
	}
	m.project.Categories[catIndex], m.project.Categories[catIndex-1] =
		m.project.Categories[catIndex-1], m.project.Categories[catIndex]
	m.rebuildPositions()
	m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
		return p.Kind == selection.FocusCategory && p.CategoryIndex == catIndex-1
	})
	m.ensureVisible()
	m.storeTaskUpdate()
}

func statusOrder(status string) int {
	switch status {
	case domain.StatusTodo:
		return 0
	case domain.StatusInProgress:
		return 1
	case domain.StatusCompleted:
		return 2
	case domain.StatusCancelled:
		return 3
	default:
		return 0
	}
}

func priorityOrder(priority string) int {
	switch priority {
	case domain.PriorityHigh:
		return 0
	case domain.PriorityMedium, "":
		return 1
	case domain.PriorityLow:
		return 2
	default:
		return 3
	}
}

func getTaskSortDate(task domain.Task) string {
	if task.UpdatedAt != "" {
		return task.UpdatedAt
	}
	return task.CreatedAt
}

func estimateOrder(estimate int) int {
	if estimate == 0 {
		return 1
	}
	return 0
}

func sortCategoryTasks(tasks []domain.Task, ascending bool) {
	sort.SliceStable(tasks, func(i, j int) bool {
		orderI := statusOrder(tasks[i].Status)
		orderJ := statusOrder(tasks[j].Status)
		if orderI != orderJ {
			if ascending {
				return orderI < orderJ
			}
			return orderI > orderJ
		}
		prioI := priorityOrder(tasks[i].Priority)
		prioJ := priorityOrder(tasks[j].Priority)
		if prioI != prioJ {
			if ascending {
				return prioI < prioJ
			}
			return prioI > prioJ
		}
		estOrderI := estimateOrder(tasks[i].EstimateMinutes)
		estOrderJ := estimateOrder(tasks[j].EstimateMinutes)
		if estOrderI != estOrderJ {
			if ascending {
				return estOrderI < estOrderJ
			}
			return estOrderI > estOrderJ
		}
		if tasks[i].EstimateMinutes != tasks[j].EstimateMinutes {
			if ascending {
				return tasks[i].EstimateMinutes < tasks[j].EstimateMinutes
			}
			return tasks[i].EstimateMinutes > tasks[j].EstimateMinutes
		}
		dateI := getTaskSortDate(tasks[i])
		dateJ := getTaskSortDate(tasks[j])
		if ascending {
			return dateI > dateJ // Ascending: newest first
		}
		return dateI < dateJ // Descending: oldest first
	})
}

func (m *model) sortTasksByStatus() {
	ascending := true
	m.ui.LastSortAscending = &ascending
	m.sortTasksByStatusOrder(true)
}

func (m *model) sortTasksByStatusReverse() {
	ascending := false
	m.ui.LastSortAscending = &ascending
	m.sortTasksByStatusOrder(false)
}

func (m *model) sortTasksByStatusOrder(ascending bool) {
	if !m.ui.Modes.CanPerformAction(modes.ActionSort) {
		return
	}

	var selectedTaskID string
	position, hasSelection := m.selectedPosition()
	if hasSelection && position.Kind == focusTask {
		selectedTaskID = m.project.Categories[position.CategoryIndex].Tasks[position.TaskIndex].ID
	}

	for i := range m.project.Categories {
		sortCategoryTasks(m.project.Categories[i].Tasks, ascending)
	}

	m.rebuildPositions()

	if selectedTaskID != "" {
		m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
			if p.Kind != selection.FocusTask {
				return false
			}
			task := m.project.Categories[p.CategoryIndex].Tasks[p.TaskIndex]
			return task.ID == selectedTaskID
		})
	}

	m.ensureVisible()
	m.storeTaskUpdate()
}
