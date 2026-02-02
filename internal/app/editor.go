package app

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"phasionary/internal/domain"
)

type editorFinishedMsg struct {
	err      error
	tempFile string
}

type ExternalEditState struct {
	TempFilePath  string
	ItemType      focusKind
	CategoryIndex int
	TaskIndex     int
}

func (e *ExternalEditState) reset() {
	e.TempFilePath = ""
	e.ItemType = focusProject
	e.CategoryIndex = -1
	e.TaskIndex = -1
}

func getEditorCmd() string {
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	return "vim"
}

func formatTaskForEdit(task domain.Task) string {
	return task.Title
}

func formatCategoryForEdit(category domain.Category) string {
	return category.Name
}

func formatProjectForEdit(project domain.Project) string {
	return project.Name
}

func (m *model) startExternalEdit() tea.Cmd {
	pos, ok := m.selectedPosition()
	if !ok {
		return nil
	}

	var content string
	switch pos.Kind {
	case focusProject:
		content = formatProjectForEdit(m.project)
	case focusCategory:
		content = formatCategoryForEdit(m.project.Categories[pos.CategoryIndex])
	case focusTask:
		content = formatTaskForEdit(m.project.Categories[pos.CategoryIndex].Tasks[pos.TaskIndex])
	}

	tempFile, err := os.CreateTemp("", "phasionary-edit-*.txt")
	if err != nil {
		m.ui.StatusMsg = fmt.Sprintf("Failed to create temp file: %v", err)
		return nil
	}

	if _, err := tempFile.WriteString(content); err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		m.ui.StatusMsg = fmt.Sprintf("Failed to write temp file: %v", err)
		return nil
	}
	tempFile.Close()

	m.ui.ExternalEdit = ExternalEditState{
		TempFilePath:  tempFile.Name(),
		ItemType:      pos.Kind,
		CategoryIndex: pos.CategoryIndex,
		TaskIndex:     pos.TaskIndex,
	}

	m.ui.Modes.ToExternalEdit()

	editor := getEditorCmd()
	c := exec.Command(editor, tempFile.Name())

	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err: err, tempFile: tempFile.Name()}
	})
}

func (m *model) handleEditorFinished(msg editorFinishedMsg) {
	defer func() {
		os.Remove(msg.tempFile)
		m.ui.ExternalEdit.reset()
		m.ui.Modes.ToNormal()
	}()

	if msg.err != nil {
		m.ui.StatusMsg = fmt.Sprintf("Editor error: %v", msg.err)
		return
	}

	content, err := os.ReadFile(msg.tempFile)
	if err != nil {
		m.ui.StatusMsg = fmt.Sprintf("Failed to read edited file: %v", err)
		return
	}

	switch m.ui.ExternalEdit.ItemType {
	case focusProject:
		m.applyProjectEdit(string(content))
	case focusCategory:
		m.applyCategoryEdit(string(content))
	case focusTask:
		m.applyTaskEdit(string(content))
	}
}

func (m *model) applyProjectEdit(content string) {
	name := strings.TrimSpace(content)
	if name == "" {
		m.ui.StatusMsg = "Project name cannot be empty"
		return
	}

	if name == m.project.Name {
		return
	}

	m.project.Name = name
	m.project.UpdatedAt = domain.NowTimestamp()
	if err := m.deps.Store.SaveProject(m.project); err != nil {
		m.ui.StatusMsg = fmt.Sprintf("Failed to save: %v", err)
		return
	}
	m.ui.StatusMsg = "Project updated"
}

func (m *model) applyCategoryEdit(content string) {
	idx := m.ui.ExternalEdit.CategoryIndex
	if idx < 0 || idx >= len(m.project.Categories) {
		m.ui.StatusMsg = "Category no longer exists"
		return
	}

	name := strings.TrimSpace(content)
	if name == "" {
		m.ui.StatusMsg = "Category name cannot be empty"
		return
	}

	if name == m.project.Categories[idx].Name {
		return
	}

	normalizedNew := domain.NormalizeName(name)
	for i, cat := range m.project.Categories {
		if i != idx && domain.NormalizeName(cat.Name) == normalizedNew {
			m.ui.StatusMsg = "A category with that name already exists"
			return
		}
	}

	m.project.Categories[idx].Name = name
	m.project.UpdatedAt = domain.NowTimestamp()
	if err := m.deps.Store.SaveProject(m.project); err != nil {
		m.ui.StatusMsg = fmt.Sprintf("Failed to save: %v", err)
		return
	}
	m.ui.StatusMsg = "Category updated"
}

func (m *model) applyTaskEdit(content string) {
	catIdx := m.ui.ExternalEdit.CategoryIndex
	taskIdx := m.ui.ExternalEdit.TaskIndex

	if catIdx < 0 || catIdx >= len(m.project.Categories) {
		m.ui.StatusMsg = "Category no longer exists"
		return
	}
	if taskIdx < 0 || taskIdx >= len(m.project.Categories[catIdx].Tasks) {
		m.ui.StatusMsg = "Task no longer exists"
		return
	}

	title := strings.TrimSpace(content)
	if title == "" {
		m.ui.StatusMsg = "Task title cannot be empty"
		return
	}

	task := &m.project.Categories[catIdx].Tasks[taskIdx]
	if title == task.Title {
		return
	}

	task.Title = title
	if err := m.deps.Store.SaveProject(m.project); err != nil {
		m.ui.StatusMsg = fmt.Sprintf("Failed to save: %v", err)
		return
	}
	m.ui.StatusMsg = "Task updated"
}
