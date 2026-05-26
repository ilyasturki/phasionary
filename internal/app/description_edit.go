package app

import (
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"

	"phasionary/internal/app/selection"
	"phasionary/internal/domain"
	"phasionary/internal/ui"
)

// startDescriptionInlineEdit opens the in-app textarea editor for the task's
// description at (catIdx, taskIdx). Use this instead of an external editor.
func (m *model) startDescriptionInlineEdit(catIdx, taskIdx int) tea.Cmd {
	if catIdx < 0 || catIdx >= len(m.project.Categories) {
		return nil
	}
	cat := &m.project.Categories[catIdx]
	if taskIdx < 0 || taskIdx >= len(cat.Tasks) {
		return nil
	}
	task := cat.Tasks[taskIdx]

	ta := textarea.New()
	ta.SetValue(task.Description)
	ta.ShowLineNumbers = false
	ta.Prompt = ""
	ta.CharLimit = 0
	if w := m.ui.Screen.Width; w > 4 {
		ta.SetWidth(w - 4)
	}
	ta.SetHeight(descriptionEditorVisibleHeight(m.ui.Screen.Height))
	cmd := ta.Focus()
	ta.CursorEnd()

	m.ui.DescriptionEdit = DescriptionEditState{
		textarea:      ta,
		categoryIndex: catIdx,
		taskIndex:     taskIdx,
		original:      task.Description,
		creating:      task.Description == "",
	}
	if !m.ui.Modes.ToDescriptionEdit() {
		return nil
	}
	return cmd
}

func descriptionEditorVisibleHeight(screenHeight int) int {
	// Leave room for project line, headers, and surrounding context.
	const chrome = 8
	h := screenHeight - chrome
	switch {
	case h < 3:
		return 3
	case h > 12:
		return 12
	default:
		return h
	}
}

func (m *model) handleDescriptionEditKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.cancelDescriptionEdit()
		return m, nil
	case "ctrl+s", "ctrl+enter":
		m.finishDescriptionEdit()
		return m, nil
	}
	var cmd tea.Cmd
	m.ui.DescriptionEdit.textarea, cmd = m.ui.DescriptionEdit.textarea.Update(msg)
	return m, cmd
}

func (m *model) cancelDescriptionEdit() {
	wasCreating := m.ui.DescriptionEdit.creating
	taskID := m.descriptionEditTaskID()
	m.ui.DescriptionEdit = DescriptionEditState{}
	m.ui.Modes.ToNormal()
	m.ui.Screen.StatusMsg = ""
	// If we opened the editor for a task that had no description, the
	// FocusDescription row may not exist; restore selection to the parent task.
	if wasCreating && taskID != "" {
		m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
			if p.Kind != selection.FocusTask || p.CategoryIndex < 0 || p.CategoryIndex >= len(m.project.Categories) {
				return false
			}
			c := m.project.Categories[p.CategoryIndex]
			return p.TaskIndex >= 0 && p.TaskIndex < len(c.Tasks) && c.Tasks[p.TaskIndex].ID == taskID
		})
	}
}

func (m *model) finishDescriptionEdit() {
	state := m.ui.DescriptionEdit
	if state.categoryIndex < 0 || state.categoryIndex >= len(m.project.Categories) {
		m.cancelDescriptionEdit()
		return
	}
	cat := &m.project.Categories[state.categoryIndex]
	if state.taskIndex < 0 || state.taskIndex >= len(cat.Tasks) {
		m.cancelDescriptionEdit()
		return
	}
	task := &cat.Tasks[state.taskIndex]
	taskID := task.ID

	newDesc := strings.Trim(strings.ReplaceAll(state.textarea.Value(), "\r\n", "\n"), "\n")
	if strings.TrimSpace(newDesc) == "" {
		newDesc = ""
	}

	if newDesc != state.original {
		m.recordHistory()
		task.Description = newDesc
		task.UpdatedAt = domain.NowTimestamp()
		m.storeTaskUpdate()
		switch {
		case state.original != "" && newDesc == "":
			m.ui.Screen.StatusMsg = "Description cleared"
		case state.original == "" && newDesc != "":
			m.ui.Screen.StatusMsg = "Description added"
		default:
			m.ui.Screen.StatusMsg = "Description updated"
		}
	} else {
		m.ui.Screen.StatusMsg = ""
	}

	m.ui.DescriptionEdit = DescriptionEditState{}
	m.ui.Modes.ToNormal()
	m.rebuildPositions()

	// Restore selection: prefer the description row if it still exists, else the task row.
	if newDesc != "" && m.ui.Screen.ExpandDescriptions {
		m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
			if p.Kind != selection.FocusDescription || p.CategoryIndex < 0 || p.CategoryIndex >= len(m.project.Categories) {
				return false
			}
			c := m.project.Categories[p.CategoryIndex]
			return p.TaskIndex >= 0 && p.TaskIndex < len(c.Tasks) && c.Tasks[p.TaskIndex].ID == taskID
		})
	} else {
		m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
			if p.Kind != selection.FocusTask || p.CategoryIndex < 0 || p.CategoryIndex >= len(m.project.Categories) {
				return false
			}
			c := m.project.Categories[p.CategoryIndex]
			return p.TaskIndex >= 0 && p.TaskIndex < len(c.Tasks) && c.Tasks[p.TaskIndex].ID == taskID
		})
	}
	m.ensureVisible()
}

func (m model) descriptionEditView() string {
	state := m.ui.DescriptionEdit
	title := "Edit Description"
	if state.creating {
		title = "Add Description"
	}
	body := state.textarea.View()
	hint := "ctrl+s save  ·  esc cancel  ·  enter newline"
	lines := []string{
		ui.DialogTitleStyle.Render(title),
		"",
		body,
		"",
		ui.DialogHintStyle.Render(hint),
	}
	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}

func (m *model) descriptionEditTaskID() string {
	s := m.ui.DescriptionEdit
	if s.categoryIndex < 0 || s.categoryIndex >= len(m.project.Categories) {
		return ""
	}
	cat := m.project.Categories[s.categoryIndex]
	if s.taskIndex < 0 || s.taskIndex >= len(cat.Tasks) {
		return ""
	}
	return cat.Tasks[s.taskIndex].ID
}
