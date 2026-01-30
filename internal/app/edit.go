package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"phasionary/internal/domain"
)

func (m *model) startEditing() {
	position, ok := m.selectedPosition()
	if !ok {
		return
	}
	switch position.Kind {
	case focusProject:
		m.mode = ModeEdit
		m.edit = newEditState(m.project.Name, false, "", focusProject)
	case focusTask:
		task := m.project.Categories[position.CategoryIndex].Tasks[position.TaskIndex]
		m.mode = ModeEdit
		m.edit = newEditState(task.Title, false, "", focusTask)
	case focusCategory:
		category := m.project.Categories[position.CategoryIndex]
		m.mode = ModeEdit
		m.edit = newEditState(category.Name, false, "", focusCategory)
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
	m.project.Categories[catIndex].Tasks = append(m.project.Categories[catIndex].Tasks, newTask)
	m.rebuildPositions()
	taskIndex := len(m.project.Categories[catIndex].Tasks) - 1
	for i, pos := range m.positions {
		if pos.Kind == focusTask && pos.CategoryIndex == catIndex && pos.TaskIndex == taskIndex {
			m.selected = i
			break
		}
	}
	m.mode = ModeEdit
	m.edit = newEditState("", true, newTask.ID, focusTask)
	m.edit.cursor = 0
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
	m.project.Categories = append(m.project.Categories, domain.Category{})
	copy(m.project.Categories[insertIndex+1:], m.project.Categories[insertIndex:])
	m.project.Categories[insertIndex] = newCat
	m.rebuildPositions()
	for i, pos := range m.positions {
		if pos.Kind == focusCategory && pos.CategoryIndex == insertIndex {
			m.selected = i
			break
		}
	}
	m.mode = ModeEdit
	m.edit = newEditState("", true, newCat.ID, focusCategory)
	m.edit.cursor = 0
	m.ensureVisible()
}

func (m *model) removeNewCategory() {
	if m.edit.newItemID == "" {
		return
	}
	catIndex := -1
	for i, cat := range m.project.Categories {
		if cat.ID == m.edit.newItemID {
			catIndex = i
			break
		}
	}
	if catIndex < 0 {
		return
	}
	m.project.Categories = append(m.project.Categories[:catIndex], m.project.Categories[catIndex+1:]...)
	m.rebuildPositions()
	if m.selected >= len(m.positions) {
		m.selected = len(m.positions) - 1
	}
	if m.selected < 0 && len(m.positions) > 0 {
		m.selected = 0
	}
	m.ensureVisible()
}

func (m *model) handleEditKey(msg tea.KeyMsg) {
	switch msg.String() {
	case "enter":
		m.finishEditing()
	case "esc":
		m.cancelEditing()
	case "left":
		m.moveEditCursor(-1)
	case "right":
		m.moveEditCursor(1)
	case "backspace":
		m.deleteEditRune(-1)
	case "delete":
		m.deleteEditRune(0)
	case " ", "space":
		m.insertEditRunes([]rune(" "))
	default:
		if msg.Type == tea.KeyRunes {
			m.insertEditRunes(msg.Runes)
		}
	}
}

func (m *model) finishEditing() {
	if m.mode != ModeEdit {
		return
	}
	position, ok := m.selectedPosition()
	if !ok {
		m.cancelEditing()
		return
	}
	trimmed := strings.TrimSpace(m.edit.value)
	if trimmed == "" {
		m.cancelEditing()
		return
	}
	switch position.Kind {
	case focusProject:
		m.project.Name = trimmed
		m.project.UpdatedAt = domain.NowTimestamp()
		m.storeTaskUpdate()
	case focusTask:
		task := &m.project.Categories[position.CategoryIndex].Tasks[position.TaskIndex]
		if task.Title != trimmed || m.edit.isAdding {
			task.Title = trimmed
			task.UpdatedAt = domain.NowTimestamp()
			m.storeTaskUpdate()
		}
	case focusCategory:
		m.finishCategoryEditing(position, trimmed)
	default:
		m.cancelEditing()
		return
	}
	m.mode = ModeNormal
	m.edit.reset()
}

func (m *model) finishCategoryEditing(position focusPosition, name string) {
	for i, cat := range m.project.Categories {
		if i != position.CategoryIndex && strings.EqualFold(cat.Name, name) {
			if m.edit.isAdding {
				m.removeNewCategory()
			}
			return
		}
	}
	m.project.Categories[position.CategoryIndex].Name = name
	m.project.UpdatedAt = domain.NowTimestamp()
	m.storeTaskUpdate()
}

func (m *model) cancelEditing() {
	if m.edit.isAdding {
		switch m.edit.itemType {
		case focusTask:
			m.removeNewTask()
		case focusCategory:
			m.removeNewCategory()
		}
	}
	m.mode = ModeNormal
	m.edit.reset()
}

func (m *model) removeNewTask() {
	if m.edit.newItemID == "" {
		return
	}
	for cIndex := range m.project.Categories {
		tasks := m.project.Categories[cIndex].Tasks
		for tIndex, task := range tasks {
			if task.ID == m.edit.newItemID {
				m.project.Categories[cIndex].Tasks = append(tasks[:tIndex], tasks[tIndex+1:]...)
				break
			}
		}
	}
	m.rebuildPositions()
	if m.selected >= len(m.positions) {
		m.selected = len(m.positions) - 1
	}
	if m.selected < 0 && len(m.positions) > 0 {
		m.selected = 0
	}
	m.ensureVisible()
}

func (m *model) moveEditCursor(delta int) {
	runes := []rune(m.edit.value)
	next := m.edit.cursor + delta
	if next < 0 {
		next = 0
	}
	if next > len(runes) {
		next = len(runes)
	}
	m.edit.cursor = next
}

func (m *model) insertEditRunes(runesToInsert []rune) {
	if len(runesToInsert) == 0 {
		return
	}
	runes := []rune(m.edit.value)
	cursor := m.edit.cursor
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(runes) {
		cursor = len(runes)
	}
	updated := make([]rune, 0, len(runes)+len(runesToInsert))
	updated = append(updated, runes[:cursor]...)
	updated = append(updated, runesToInsert...)
	updated = append(updated, runes[cursor:]...)
	m.edit.value = string(updated)
	m.edit.cursor = cursor + len(runesToInsert)
}

func (m *model) deleteEditRune(offset int) {
	runes := []rune(m.edit.value)
	if len(runes) == 0 {
		return
	}
	index := m.edit.cursor + offset
	if offset < 0 {
		index = m.edit.cursor - 1
	}
	if index < 0 || index >= len(runes) {
		return
	}
	updated := append([]rune{}, runes[:index]...)
	updated = append(updated, runes[index+1:]...)
	m.edit.value = string(updated)
	if offset < 0 {
		m.edit.cursor = index
	} else if m.edit.cursor > len(updated) {
		m.edit.cursor = len(updated)
	}
}
