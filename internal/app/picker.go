package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const pickerVisibleItems = 10

func (m *model) openProjectPicker() {
	projects, err := m.deps.Store.ListProjects()
	if err != nil {
		m.ui.StatusMsg = fmt.Sprintf("Error loading projects: %v", err)
		return
	}

	currentIdx := 0
	for i, p := range projects {
		if p.ID == m.project.ID {
			currentIdx = i
			break
		}
	}

	m.ui.Picker = ProjectPickerState{
		projects:     projects,
		selected:     currentIdx,
		scrollOffset: 0,
	}
	m.ui.Picker.ensureVisible()
	m.ui.Modes.ToProjectPicker()
}

func (m model) handleProjectPickerKey(msg tea.KeyMsg) (model, tea.Cmd) {
	if m.ui.Picker.isAdding {
		return m.handlePickerAddKey(msg)
	}
	switch msg.String() {
	case "j", "down":
		m.ui.Picker.moveSelection(1)
	case "k", "up":
		m.ui.Picker.moveSelection(-1)
	case "enter":
		if m.ui.Picker.isOnAddButton() {
			m.ui.Picker.startAdding()
		} else {
			m.selectProject()
		}
	case "esc", "q":
		m.ui.Picker.reset()
		m.ui.Modes.ToNormal()
	}
	return m, nil
}

func (m model) handlePickerAddKey(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.createProjectFromPicker()
		return m, nil
	case "esc":
		m.ui.Picker.cancelAdding()
		return m, nil
	}
	var cmd tea.Cmd
	m.ui.Picker.input, cmd = m.ui.Picker.input.Update(msg)
	return m, cmd
}

func (m *model) createProjectFromPicker() {
	name := strings.TrimSpace(m.ui.Picker.input.Value())
	if name == "" {
		m.ui.Picker.cancelAdding()
		return
	}

	project, err := m.deps.Store.CreateProject(name)
	if err != nil {
		m.ui.StatusMsg = fmt.Sprintf("Error: %v", err)
		return
	}

	m.project = project
	m.ui.Filter = NewFilterState()
	positions := rebuildPositions(project.Categories, &m.ui.Filter)
	initialSelection := findFirstTaskIndex(positions)
	m.ui.Selection.SetPositions(toSelectionPositions(positions))
	m.ui.Selection.SetSelected(initialSelection)
	m.ui.ScrollOffset = 0

	m.ensureVisible()
	m.ui.StatusMsg = fmt.Sprintf("Created project: %s", project.Name)
	m.ui.Picker.reset()
	m.ui.Modes.ToNormal()
}

func (m *model) selectProject() {
	if m.ui.Picker.selected < 0 || m.ui.Picker.selected >= len(m.ui.Picker.projects) {
		m.ui.Picker.reset()
		m.ui.Modes.ToNormal()
		return
	}

	selectedProject := m.ui.Picker.projects[m.ui.Picker.selected]
	if selectedProject.ID == m.project.ID {
		m.ui.Picker.reset()
		m.ui.Modes.ToNormal()
		return
	}

	project, err := m.deps.Store.LoadProject(selectedProject.ID)
	if err != nil {
		m.ui.StatusMsg = fmt.Sprintf("Error loading project: %v", err)
		m.ui.Picker.reset()
		m.ui.Modes.ToNormal()
		return
	}

	m.project = project
	m.ui.Filter = NewFilterState()
	positions := rebuildPositions(project.Categories, &m.ui.Filter)
	initialSelection := findFirstTaskIndex(positions)
	m.ui.Selection.SetPositions(toSelectionPositions(positions))
	m.ui.Selection.SetSelected(initialSelection)
	m.ui.ScrollOffset = 0

	m.ensureVisible()
	m.ui.StatusMsg = fmt.Sprintf("Switched to: %s", project.Name)
	m.ui.Picker.reset()
	m.ui.Modes.ToNormal()
}

func (p *ProjectPickerState) moveSelection(delta int) {
	total := p.totalItems()
	p.selected += delta
	if p.selected < 0 {
		p.selected = 0
	}
	if p.selected >= total {
		p.selected = total - 1
	}
	p.ensureVisible()
}

func (p *ProjectPickerState) ensureVisible() {
	if p.selected < p.scrollOffset {
		p.scrollOffset = p.selected
	}
	if p.selected >= p.scrollOffset+pickerVisibleItems {
		p.scrollOffset = p.selected - pickerVisibleItems + 1
	}
}
