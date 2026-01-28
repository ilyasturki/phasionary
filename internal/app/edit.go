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
	case focusTask:
		task := m.categories[position.CategoryIndex].Tasks[position.TaskIndex]
		m.editing = true
		m.editValue = task.Title
		m.editCursor = len([]rune(m.editValue))
	case focusCategory:
		category := m.categories[position.CategoryIndex]
		m.editing = true
		m.editValue = category.Name
		m.editCursor = len([]rune(m.editValue))
	}
}

func (m *model) startAddingTask() {
	position, ok := m.selectedPosition()
	if !ok {
		return
	}
	catIndex := position.CategoryIndex
	if catIndex < 0 || catIndex >= len(m.categories) {
		return
	}

	// Create new task with empty title (defaults to todo status)
	newTask, err := domain.NewTask("")
	if err != nil {
		return
	}

	// Append task to both view and project
	m.categories[catIndex].Tasks = append(m.categories[catIndex].Tasks, newTask)
	m.project.Categories[catIndex].Tasks = append(m.project.Categories[catIndex].Tasks, newTask)

	// Rebuild positions
	m.positions = rebuildPositions(m.categories)

	// Find and select the new task position
	taskIndex := len(m.categories[catIndex].Tasks) - 1
	for i, pos := range m.positions {
		if pos.Kind == focusTask && pos.CategoryIndex == catIndex && pos.TaskIndex == taskIndex {
			m.selected = i
			break
		}
	}

	// Enter edit mode for the new task
	m.editing = true
	m.addingTask = true
	m.newTaskID = newTask.ID
	m.editValue = ""
	m.editCursor = 0
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

	// Insert into view categories
	m.categories = append(m.categories, categoryView{})
	copy(m.categories[insertIndex+1:], m.categories[insertIndex:])
	m.categories[insertIndex] = categoryView{Name: "", Tasks: nil}

	// Insert into project categories
	m.project.Categories = append(m.project.Categories, domain.Category{})
	copy(m.project.Categories[insertIndex+1:], m.project.Categories[insertIndex:])
	m.project.Categories[insertIndex] = newCat

	// Rebuild positions and select the new category
	m.positions = rebuildPositions(m.categories)
	for i, pos := range m.positions {
		if pos.Kind == focusCategory && pos.CategoryIndex == insertIndex {
			m.selected = i
			break
		}
	}

	m.editing = true
	m.addingCategory = true
	m.newCategoryID = newCat.ID
	m.editValue = ""
	m.editCursor = 0
}

func (m *model) removeNewCategory() {
	if m.newCategoryID == "" {
		return
	}

	// Find the category index by ID
	catIndex := -1
	for i, cat := range m.project.Categories {
		if cat.ID == m.newCategoryID {
			catIndex = i
			break
		}
	}
	if catIndex < 0 {
		return
	}

	// Remove from view categories
	m.categories = append(m.categories[:catIndex], m.categories[catIndex+1:]...)

	// Remove from project categories
	m.project.Categories = append(m.project.Categories[:catIndex], m.project.Categories[catIndex+1:]...)

	// Rebuild positions and clamp selection
	m.positions = rebuildPositions(m.categories)
	if m.selected >= len(m.positions) {
		m.selected = len(m.positions) - 1
	}
	if m.selected < 0 && len(m.positions) > 0 {
		m.selected = 0
	}
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
	if !m.editing {
		return
	}
	position, ok := m.selectedPosition()
	if !ok {
		m.cancelEditing()
		return
	}
	trimmed := strings.TrimSpace(m.editValue)
	if trimmed == "" {
		m.cancelEditing()
		return
	}
	switch position.Kind {
	case focusTask:
		category := &m.categories[position.CategoryIndex]
		task := &category.Tasks[position.TaskIndex]
		if task.Title != trimmed || m.addingTask {
			task.Title = trimmed
			task.UpdatedAt = domain.NowTimestamp()
			m.syncTaskToProject(position, *task)
			m.storeTaskUpdate()
		}
		m.refreshtaskview(position)
	case focusCategory:
		m.finishCategoryEditing(position, trimmed)
	default:
		m.cancelEditing()
		return
	}
	m.editing = false
	m.editValue = ""
	m.editCursor = 0
	m.addingTask = false
	m.newTaskID = ""
	m.addingCategory = false
	m.newCategoryID = ""
}

func (m *model) finishCategoryEditing(position focusPosition, name string) {
	// Check for duplicate name (case-insensitive) among other categories
	for i, cat := range m.categories {
		if i != position.CategoryIndex && strings.EqualFold(cat.Name, name) {
			// Duplicate found — remove phantom category if adding
			if m.addingCategory {
				m.removeNewCategory()
			}
			return
		}
	}
	m.categories[position.CategoryIndex].Name = name
	m.project.Categories[position.CategoryIndex].Name = name
	m.project.UpdatedAt = domain.NowTimestamp()
	m.storeTaskUpdate()
}

func (m *model) cancelEditing() {
	if m.addingTask {
		m.removeNewTask()
	}
	if m.addingCategory {
		m.removeNewCategory()
	}
	m.editing = false
	m.editValue = ""
	m.editCursor = 0
	m.addingTask = false
	m.newTaskID = ""
	m.addingCategory = false
	m.newCategoryID = ""
}

func (m *model) removeNewTask() {
	if m.newTaskID == "" {
		return
	}

	// Find and remove from view categories
	for cIndex := range m.categories {
		tasks := m.categories[cIndex].Tasks
		for tIndex, task := range tasks {
			if task.ID == m.newTaskID {
				m.categories[cIndex].Tasks = append(tasks[:tIndex], tasks[tIndex+1:]...)
				break
			}
		}
	}

	// Find and remove from project categories
	for cIndex := range m.project.Categories {
		tasks := m.project.Categories[cIndex].Tasks
		for tIndex, task := range tasks {
			if task.ID == m.newTaskID {
				m.project.Categories[cIndex].Tasks = append(tasks[:tIndex], tasks[tIndex+1:]...)
				break
			}
		}
	}

	// Rebuild positions and adjust selection
	m.positions = rebuildPositions(m.categories)
	if m.selected >= len(m.positions) {
		m.selected = len(m.positions) - 1
	}
	if m.selected < 0 && len(m.positions) > 0 {
		m.selected = 0
	}
}

func (m *model) moveEditCursor(delta int) {
	runes := []rune(m.editValue)
	next := m.editCursor + delta
	if next < 0 {
		next = 0
	}
	if next > len(runes) {
		next = len(runes)
	}
	m.editCursor = next
}

func (m *model) insertEditRunes(runesToInsert []rune) {
	if len(runesToInsert) == 0 {
		return
	}
	runes := []rune(m.editValue)
	cursor := m.editCursor
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
	m.editValue = string(updated)
	m.editCursor = cursor + len(runesToInsert)
}

func (m *model) deleteEditRune(offset int) {
	runes := []rune(m.editValue)
	if len(runes) == 0 {
		return
	}
	index := m.editCursor + offset
	if offset < 0 {
		index = m.editCursor - 1
	}
	if index < 0 || index >= len(runes) {
		return
	}
	updated := append([]rune{}, runes[:index]...)
	updated = append(updated, runes[index+1:]...)
	m.editValue = string(updated)
	if offset < 0 {
		m.editCursor = index
	} else if m.editCursor > len(updated) {
		m.editCursor = len(updated)
	}
}
