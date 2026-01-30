package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

const pickerVisibleItems = 10

func (m *model) openProjectPicker() {
	projects, err := m.deps.Store.ListProjects()
	if err != nil {
		m.ui.StatusMsg = fmt.Sprintf("Error loading projects: %v", err)
		return
	}
	if len(projects) == 0 {
		m.ui.StatusMsg = "No projects available"
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

func (m model) handleProjectPickerKey(msg tea.KeyMsg) model {
	switch msg.String() {
	case "j", "down":
		m.ui.Picker.moveSelection(1)
	case "k", "up":
		m.ui.Picker.moveSelection(-1)
	case "enter":
		m.selectProject()
	case "esc", "q":
		m.ui.Picker.reset()
		m.ui.Modes.ToNormal()
	}
	return m
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
	positions := rebuildPositions(project.Categories)
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
	if len(p.projects) == 0 {
		return
	}
	p.selected += delta
	if p.selected < 0 {
		p.selected = 0
	}
	if p.selected >= len(p.projects) {
		p.selected = len(p.projects) - 1
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
