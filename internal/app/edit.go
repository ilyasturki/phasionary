package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"phasionary/internal/app/selection"
	"phasionary/internal/domain"
)

func (m *model) startEditing() {
	position, ok := m.selectedPosition()
	if !ok {
		return
	}
	switch position.Kind {
	case selection.FocusProject:
		m.ui.Modes.ToEdit()
		m.ui.Edit = newEditState(m.project.Name, false, "", selection.FocusProject)
	case selection.FocusTask:
		task := m.project.Categories[position.CategoryIndex].Tasks[position.TaskIndex]
		m.ui.Modes.ToEdit()
		m.ui.Edit = newEditState(task.Title, false, "", selection.FocusTask)
	case selection.FocusCategory:
		category := m.project.Categories[position.CategoryIndex]
		m.ui.Modes.ToEdit()
		m.ui.Edit = newEditState(category.Name, false, "", selection.FocusCategory)
	}
}

func (m *model) startAddingTask() {
	position, ok := m.selectedPosition()
	if !ok {
		return
	}
	catIndex := position.CategoryIndex
	if catIndex < 0 || catIndex >= len(m.project.Categories) {
		return
	}
	newTask, err := domain.NewTask("")
	if err != nil {
		return
	}

	insertAtTop := m.ui.LastSortAscending == nil || *m.ui.LastSortAscending
	var taskIndex int
	if insertAtTop {
		m.project.Categories[catIndex].InsertTask(0, newTask)
		taskIndex = 0
	} else {
		m.project.Categories[catIndex].AddTask(newTask)
		taskIndex = len(m.project.Categories[catIndex].Tasks) - 1
	}

	m.rebuildPositions()
	m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
		return p.Kind == selection.FocusTask && p.CategoryIndex == catIndex && p.TaskIndex == taskIndex
	})
	m.ui.Modes.ToEdit()
	m.ui.Edit = newEditState("", true, newTask.ID, selection.FocusTask)
	m.ensureVisible()
}

func (m *model) startAddingCategory() {
	position, ok := m.selectedPosition()
	if !ok {
		return
	}
	insertIndex := position.CategoryIndex + 1
	newCat, err := domain.NewCategory("")
	if err != nil {
		return
	}
	m.project.InsertCategory(insertIndex, newCat)
	m.rebuildPositions()
	m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
		return p.Kind == selection.FocusCategory && p.CategoryIndex == insertIndex
	})
	m.ui.Modes.ToEdit()
	m.ui.Edit = newEditState("", true, newCat.ID, selection.FocusCategory)
	m.ensureVisible()
}

func (m *model) removeNewCategory() {
	if m.ui.Edit.newItemID == "" {
		return
	}
	catIndex := -1
	for i, cat := range m.project.Categories {
		if cat.ID == m.ui.Edit.newItemID {
			catIndex = i
			break
		}
	}
	if catIndex < 0 {
		return
	}
	_ = m.project.RemoveCategory(catIndex)
	m.rebuildPositions()
	m.ensureVisible()
}

func (m *model) handleEditKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "enter":
		m.finishEditing()
		return nil
	case "esc":
		m.cancelEditing()
		return nil
	}
	var cmd tea.Cmd
	m.ui.Edit.input, cmd = m.ui.Edit.input.Update(msg)
	sanitizeInput(&m.ui.Edit.input)
	return cmd
}

func (m *model) finishEditing() {
	if !m.ui.Modes.IsEdit() {
		return
	}
	position, ok := m.selectedPosition()
	if !ok {
		m.cancelEditing()
		return
	}
	trimmed := strings.TrimSpace(m.ui.Edit.input.Value())
	if trimmed == "" {
		m.cancelEditing()
		return
	}
	switch position.Kind {
	case selection.FocusProject:
		m.project.Name = trimmed
		m.project.UpdatedAt = domain.NowTimestamp()
		m.storeTaskUpdate()
	case selection.FocusTask:
		task := &m.project.Categories[position.CategoryIndex].Tasks[position.TaskIndex]
		if task.Title != trimmed || m.ui.Edit.isAdding {
			task.Title = trimmed
			task.UpdatedAt = domain.NowTimestamp()

			if m.ui.Edit.isAdding {
				taskID := task.ID
				ascending := m.ui.LastSortAscending == nil || *m.ui.LastSortAscending
				sortCategoryTasks(m.project.Categories[position.CategoryIndex].Tasks, ascending)
				m.rebuildPositions()
				m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
					return p.Kind == selection.FocusTask &&
						p.CategoryIndex == position.CategoryIndex &&
						m.project.Categories[p.CategoryIndex].Tasks[p.TaskIndex].ID == taskID
				})
			}

			m.storeTaskUpdate()
		}
	case selection.FocusCategory:
		m.finishCategoryEditing(position, trimmed)
	default:
		m.cancelEditing()
		return
	}
	m.ui.Modes.ToNormal()
	m.ui.Edit.reset()
}

func (m *model) finishCategoryEditing(position selection.Position, name string) {
	if err := m.project.RenameCategory(position.CategoryIndex, name); err != nil {
		if m.ui.Edit.isAdding {
			m.removeNewCategory()
		}
		return
	}
	m.storeTaskUpdate()
}

func (m *model) cancelEditing() {
	if m.ui.Edit.isAdding {
		switch m.ui.Edit.itemType {
		case selection.FocusTask:
			m.removeNewTask()
		case selection.FocusCategory:
			m.removeNewCategory()
		}
	}
	m.ui.Modes.ToNormal()
	m.ui.Edit.reset()
}

func (m model) forwardToInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.ui.Modes.IsEdit() {
		var cmd tea.Cmd
		m.ui.Edit.input, cmd = m.ui.Edit.input.Update(msg)
		sanitizeInput(&m.ui.Edit.input)
		return m, cmd
	}
	if m.ui.Picker.isAdding {
		var cmd tea.Cmd
		m.ui.Picker.input, cmd = m.ui.Picker.input.Update(msg)
		sanitizeInput(&m.ui.Picker.input)
		return m, cmd
	}
	return m, nil
}

func (m *model) removeNewTask() {
	if m.ui.Edit.newItemID == "" {
		return
	}
	for cIndex := range m.project.Categories {
		for tIndex, task := range m.project.Categories[cIndex].Tasks {
			if task.ID == m.ui.Edit.newItemID {
				_ = m.project.Categories[cIndex].RemoveTask(tIndex)
				break
			}
		}
	}
	m.rebuildPositions()
	m.ensureVisible()
}
