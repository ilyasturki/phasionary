package app

import (
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
	if !ok || pos.Kind == selection.FocusProject {
		return
	}
	if pos.Kind == selection.FocusDescription {
		m.clearDescription(pos)
		return
	}
	m.ui.ConfirmDelete = ConfirmDeleteState{Kind: ConfirmDeleteSelection}
	m.ui.Modes.ToConfirmDelete()
}

func (m *model) clearDescription(pos selection.Position) {
	if pos.CategoryIndex < 0 || pos.CategoryIndex >= len(m.project.Categories) {
		return
	}
	cat := &m.project.Categories[pos.CategoryIndex]
	if pos.TaskIndex < 0 || pos.TaskIndex >= len(cat.Tasks) {
		return
	}
	task := &cat.Tasks[pos.TaskIndex]
	if task.Description == "" {
		return
	}
	taskID := task.ID
	task.Description = ""
	task.UpdatedAt = domain.NowTimestamp()
	m.storeTaskUpdate()
	m.rebuildPositions()
	m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
		if p.Kind != selection.FocusTask || p.CategoryIndex < 0 || p.CategoryIndex >= len(m.project.Categories) {
			return false
		}
		c := m.project.Categories[p.CategoryIndex]
		return p.TaskIndex >= 0 && p.TaskIndex < len(c.Tasks) && c.Tasks[p.TaskIndex].ID == taskID
	})
	m.ensureVisible()
	m.ui.Screen.StatusMsg = "Description cleared"
}

func (m *model) confirmDeleteAction() {
	m.ui.ConfirmDelete.reset()
	m.ui.Modes.ToNormal()
	position, ok := m.selectedPosition()
	if !ok {
		return
	}
	switch position.Kind {
	case selection.FocusTask:
		m.deleteTask(position)
	case selection.FocusCategory:
		m.deleteCategory(position)
	}
}

func (m *model) deleteTask(position selection.Position) {
	catIndex := position.CategoryIndex
	taskIndex := position.TaskIndex

	task := m.project.Categories[catIndex].Tasks[taskIndex]
	taskCopy := task
	m.ui.Clipboard.Task = &taskCopy
	m.ui.Clipboard.IsCut = false
	m.ui.Clipboard.SourceID = ""

	_ = m.project.Categories[catIndex].RemoveTask(taskIndex)
	m.rebuildAndClamp()
	m.storeTaskUpdate()
}

func (m *model) deleteCategory(position selection.Position) {
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
	if !ok || position.Kind != selection.FocusTask {
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
	if !ok || position.Kind != selection.FocusTask {
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
	if !ok || position.Kind != selection.FocusTask {
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
	if !ok || position.Kind == selection.FocusProject {
		return
	}

	var currentEstimate int
	if position.Kind == selection.FocusTask {
		currentEstimate = m.project.Categories[position.CategoryIndex].Tasks[position.TaskIndex].EstimateMinutes
	} else {
		currentEstimate = m.project.Categories[position.CategoryIndex].EstimateMinutes
	}

	m.ui.EstimatePicker = components.NewEstimatePickerState(currentEstimate)
	m.ui.Modes.ToEstimatePicker()
}

func (m *model) selectEstimate(minutes int) {
	position, ok := m.selectedPosition()
	if !ok || position.Kind == selection.FocusProject {
		return
	}

	if position.Kind == selection.FocusTask {
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
	if !ok || position.Kind != selection.FocusTask {
		return
	}
	catIndex := position.CategoryIndex
	taskIndex := position.TaskIndex
	tasks := m.project.Categories[catIndex].Tasks
	if taskIndex < len(tasks)-1 {
		tasks[taskIndex], tasks[taskIndex+1] = tasks[taskIndex+1], tasks[taskIndex]
		m.rebuildPositions()
		m.ui.Selection.MoveBy(1)
		m.ensureVisible()
		m.storeTaskUpdate()
		return
	}
	if catIndex >= len(m.project.Categories)-1 {
		return
	}
	task := tasks[taskIndex]
	_ = m.project.Categories[catIndex].RemoveTask(taskIndex)
	dstCatIndex := catIndex + 1
	m.project.Categories[dstCatIndex].InsertTask(0, task)
	m.project.UpdatedAt = domain.NowTimestamp()
	m.rebuildPositions()
	m.selectTaskOrCategory(dstCatIndex, 0)
	m.ensureVisible()
	m.storeTaskUpdate()
}

func (m *model) moveTaskUp() {
	if !m.ui.Modes.CanPerformAction(modes.ActionMoveItem) {
		return
	}
	position, ok := m.selectedPosition()
	if !ok || position.Kind != selection.FocusTask {
		return
	}
	catIndex := position.CategoryIndex
	taskIndex := position.TaskIndex
	if taskIndex > 0 {
		tasks := m.project.Categories[catIndex].Tasks
		tasks[taskIndex], tasks[taskIndex-1] = tasks[taskIndex-1], tasks[taskIndex]
		m.rebuildPositions()
		m.ui.Selection.MoveBy(-1)
		m.ensureVisible()
		m.storeTaskUpdate()
		return
	}
	if catIndex <= 0 {
		return
	}
	dstCatIndex := catIndex - 1
	_ = m.project.MoveTask(catIndex, taskIndex, dstCatIndex)
	dstTaskIndex := len(m.project.Categories[dstCatIndex].Tasks) - 1
	m.rebuildPositions()
	m.selectTaskOrCategory(dstCatIndex, dstTaskIndex)
	m.ensureVisible()
	m.storeTaskUpdate()
}

func (m *model) selectTaskOrCategory(catIndex, taskIndex int) {
	if m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
		return p.Kind == selection.FocusTask && p.CategoryIndex == catIndex && p.TaskIndex == taskIndex
	}) {
		return
	}
	m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
		return p.Kind == selection.FocusCategory && p.CategoryIndex == catIndex
	})
}

func (m *model) moveCategoryDown() {
	if !m.ui.Modes.CanPerformAction(modes.ActionMoveItem) {
		return
	}
	position, ok := m.selectedPosition()
	if !ok || position.Kind != selection.FocusCategory {
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
	if !ok || position.Kind != selection.FocusCategory {
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

func (m *model) cutSelectedTask() {
	if !m.ui.Modes.CanPerformAction(modes.ActionDeleteItem) {
		return
	}
	position, ok := m.selectedPosition()
	if !ok || position.Kind != selection.FocusTask {
		m.ui.Screen.StatusMsg = "Can only cut tasks"
		return
	}

	task := m.project.Categories[position.CategoryIndex].Tasks[position.TaskIndex]
	taskCopy := task
	m.ui.Clipboard.Task = &taskCopy
	m.ui.Clipboard.IsCut = true
	m.ui.Clipboard.SourceID = task.ID

	title := task.Title
	if len([]rune(title)) > 30 {
		title = string([]rune(title)[:30]) + "..."
	}
	m.ui.Screen.StatusMsg = "Marked for cut: " + title
}

func (m *model) pasteFromClipboard() {
	if m.pasteMulti() {
		return
	}
	m.pasteTask()
}

func (m *model) pasteTask() {
	if m.ui.Clipboard.Task == nil {
		m.ui.Screen.StatusMsg = "Nothing to paste"
		return
	}

	newID, err := domain.NewID()
	if err != nil {
		m.ui.Screen.StatusMsg = "Failed to create task ID"
		return
	}

	newTask := domain.Task{
		ID:              newID,
		Title:           m.ui.Clipboard.Task.Title,
		Status:          m.ui.Clipboard.Task.Status,
		CreatedAt:       m.ui.Clipboard.Task.CreatedAt,
		UpdatedAt:       domain.NowTimestamp(),
		Priority:        m.ui.Clipboard.Task.Priority,
		CompletionDate:  m.ui.Clipboard.Task.CompletionDate,
		EstimateMinutes: m.ui.Clipboard.Task.EstimateMinutes,
		Description:     m.ui.Clipboard.Task.Description,
	}

	position, ok := m.selectedPosition()
	var catIndex, taskIndex int

	if !ok || len(m.project.Categories) == 0 {
		m.ui.Screen.StatusMsg = "No category to paste into"
		return
	}

	switch position.Kind {
	case selection.FocusProject:
		catIndex = 0
		taskIndex = 0
	case selection.FocusCategory:
		catIndex = position.CategoryIndex
		taskIndex = 0
	case selection.FocusTask:
		catIndex = position.CategoryIndex
		taskIndex = position.TaskIndex + 1
	}

	if m.ui.Clipboard.IsCut {
		srcCat, srcIdx := m.findTaskByID(m.ui.Clipboard.SourceID)
		if srcCat == catIndex && srcIdx >= 0 && srcIdx < taskIndex {
			taskIndex--
		}
		m.removeTaskByID(m.ui.Clipboard.SourceID)
	}

	if taskIndex > len(m.project.Categories[catIndex].Tasks) {
		taskIndex = len(m.project.Categories[catIndex].Tasks)
	}

	m.project.Categories[catIndex].InsertTask(taskIndex, newTask)

	statusMsg := "Pasted!"
	if m.ui.Clipboard.IsCut {
		statusMsg = "Moved!"
	}
	m.ui.Clipboard = ClipboardState{}

	m.rebuildPositions()
	m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
		if p.Kind != selection.FocusTask {
			return false
		}
		return m.project.Categories[p.CategoryIndex].Tasks[p.TaskIndex].ID == newID
	})
	m.ensureVisible()
	m.storeTaskUpdate()
	m.ui.Screen.StatusMsg = statusMsg
}

func (m *model) removeTaskByID(id string) {
	for catIdx := range m.project.Categories {
		for taskIdx, task := range m.project.Categories[catIdx].Tasks {
			if task.ID == id {
				_ = m.project.Categories[catIdx].RemoveTask(taskIdx)
				return
			}
		}
	}
}

func (m *model) findTaskByID(id string) (int, int) {
	for catIdx := range m.project.Categories {
		for taskIdx, task := range m.project.Categories[catIdx].Tasks {
			if task.ID == id {
				return catIdx, taskIdx
			}
		}
	}
	return -1, -1
}
