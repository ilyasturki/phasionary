package app

import (
	"fmt"
	"strings"

	"phasionary/internal/app/selection"
	"phasionary/internal/ui"
)

func (m model) confirmDeleteView() string {
	if m.ui.ConfirmDelete.Kind == ConfirmDeleteProject {
		return m.confirmDeleteProjectView()
	}

	position, ok := m.selectedPosition()
	if !ok || position.Kind == selection.FocusProject {
		return ""
	}
	var message string
	if position.Kind == selection.FocusTask {
		task := m.project.Categories[position.CategoryIndex].Tasks[position.TaskIndex]
		message = fmt.Sprintf("Delete task %q?", truncateText(task.Title, 30))
	} else {
		cat := m.project.Categories[position.CategoryIndex]
		message = fmt.Sprintf("Delete category %q and %d tasks?", truncateText(cat.Name, 30), len(cat.Tasks))
	}
	lines := []string{
		message,
		"",
		ui.DialogHintStyle.Render("y/enter confirm | n/esc cancel"),
	}
	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}

func (m model) confirmDeleteProjectView() string {
	var projectName string
	for _, p := range m.ui.Picker.projects {
		if p.ID == m.ui.ConfirmDelete.ProjectID {
			projectName = p.Name
			break
		}
	}
	lines := []string{
		fmt.Sprintf("Delete project %q?", truncateText(projectName, 30)),
		"",
		ui.DialogHintStyle.Render("y/enter confirm | n/esc cancel"),
	}
	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}
