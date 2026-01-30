package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

const pickerVisibleItems = 10

func (m *model) openProjectPicker() {
	projects, err := m.store.ListProjects()
	if err != nil {
		m.statusMsg = fmt.Sprintf("Error loading projects: %v", err)
		return
	}
	if len(projects) == 0 {
		m.statusMsg = "No projects available"
		return
	}

	currentIdx := 0
	for i, p := range projects {
		if p.ID == m.project.ID {
			currentIdx = i
			break
		}
	}

	m.picker = ProjectPickerState{
		projects:     projects,
		selected:     currentIdx,
		scrollOffset: 0,
	}
	m.picker.ensureVisible()
	m.mode = ModeProjectPicker
}

func (m model) handleProjectPickerKey(msg tea.KeyMsg) model {
	switch msg.String() {
	case "j", "down":
		m.picker.moveSelection(1)
	case "k", "up":
		m.picker.moveSelection(-1)
	case "enter":
		m.selectProject()
	case "esc", "q":
		m.picker.reset()
		m.mode = ModeNormal
	}
	return m
}

func (m *model) selectProject() {
	if m.picker.selected < 0 || m.picker.selected >= len(m.picker.projects) {
		m.picker.reset()
		m.mode = ModeNormal
		return
	}

	selectedProject := m.picker.projects[m.picker.selected]
	if selectedProject.ID == m.project.ID {
		m.picker.reset()
		m.mode = ModeNormal
		return
	}

	project, err := m.store.LoadProject(selectedProject.ID)
	if err != nil {
		m.statusMsg = fmt.Sprintf("Error loading project: %v", err)
		m.picker.reset()
		m.mode = ModeNormal
		return
	}

	m.project = project
	m.positions = rebuildPositions(project.Categories)
	m.scrollOffset = 0

	if len(m.positions) > 0 {
		m.selected = 0
		for i, pos := range m.positions {
			if pos.Kind == focusTask {
				m.selected = i
				break
			}
		}
	} else {
		m.selected = -1
	}

	m.ensureVisible()
	m.statusMsg = fmt.Sprintf("Switched to: %s", project.Name)
	m.picker.reset()
	m.mode = ModeNormal
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
