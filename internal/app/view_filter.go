package app

import (
	"fmt"
	"strings"

	"phasionary/internal/ui"
)

func (m model) filterView() string {
	switch m.ui.Filter.View() {
	case FilterViewStatus:
		return m.filterStatusView()
	case FilterViewPriority:
		return m.filterPriorityView()
	case FilterViewCategory:
		return m.filterCategoryView()
	default:
		return m.filterHubView()
	}
}

func (m model) filterHubView() string {
	rows := []struct {
		label string
		count int
	}{
		{"Status", m.ui.Filter.StatusCount()},
		{"Priority", m.ui.Filter.PriorityCount()},
		{"Category", m.ui.Filter.CategoryCount()},
	}

	lines := []string{ui.DialogTitleStyle.Render("Filter Tasks"), ""}
	selected := m.ui.Filter.HubSelected()
	for i, row := range rows {
		marker := "  "
		if i == selected {
			marker = "> "
		}
		summary := ui.MutedStyle.Render("[—]")
		if row.count > 0 {
			summary = fmt.Sprintf("[%d active]", row.count)
		}
		line := fmt.Sprintf("%s%-10s %s", marker, row.label, summary)
		if i == selected {
			line = ui.SelectedStyle.Render(line)
		}
		lines = append(lines, line)
	}

	lines = append(lines, "")
	const clearAllLabel = "Clear all filters"
	var clearLine string
	switch {
	case selected == FilterHubClearAll:
		clearLine = ui.SelectedStyle.Render("> " + clearAllLabel)
	case !m.ui.Filter.HasActiveFilter():
		clearLine = "  " + ui.MutedStyle.Render(clearAllLabel)
	default:
		clearLine = "  " + clearAllLabel
	}
	lines = append(lines, clearLine)

	lines = append(lines, "", ui.DialogHintStyle.Render("enter open · q/esc/f close"))
	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}

func (m model) filterStatusView() string {
	lines := []string{ui.DialogTitleStyle.Render("Filter · Status"), ""}
	for i, status := range filterStatuses {
		lines = append(lines, renderCheckboxLine(
			formatStatusLabel(status),
			m.ui.Filter.IsStatusEnabled(status),
			i == m.ui.Filter.Selected(),
		))
	}
	lines = append(lines, "", ui.DialogHintStyle.Render("space toggle · esc back · f close"))
	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}

func (m model) filterPriorityView() string {
	lines := []string{ui.DialogTitleStyle.Render("Filter · Priority"), ""}
	for i, p := range filterPriorities {
		lines = append(lines, renderCheckboxLine(
			formatPriorityLabel(p),
			m.ui.Filter.IsPriorityEnabled(p),
			i == m.ui.Filter.Selected(),
		))
	}
	lines = append(lines, "", ui.DialogHintStyle.Render("space toggle · esc back · f close"))
	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}

func (m model) filterCategoryView() string {
	lines := []string{ui.DialogTitleStyle.Render("Filter · Category"), ""}
	if len(m.project.Categories) == 0 {
		lines = append(lines, ui.MutedStyle.Render("  (no categories)"))
	} else {
		for i, cat := range m.project.Categories {
			lines = append(lines, renderCheckboxLine(
				cat.Name,
				m.ui.Filter.IsCategoryEnabled(cat.ID),
				i == m.ui.Filter.Selected(),
			))
		}
	}
	lines = append(lines, "", ui.DialogHintStyle.Render("space toggle · esc back · f close"))
	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}

func renderCheckboxLine(label string, enabled, selected bool) string {
	marker := "  "
	if selected {
		marker = "> "
	}
	checkbox := "[ ]"
	if enabled {
		checkbox = "[x]"
	}
	line := fmt.Sprintf("%s%s %s", marker, checkbox, label)
	if selected {
		line = ui.SelectedStyle.Render(line)
	}
	return line
}
