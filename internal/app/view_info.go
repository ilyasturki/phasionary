package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"phasionary/internal/app/selection"
	"phasionary/internal/domain"
	"phasionary/internal/ui"
)

func (m model) infoView() string {
	pos, ok := m.selectedPosition()
	if !ok {
		return ""
	}

	var lines []string
	switch pos.Kind {
	case selection.FocusProject:
		lines = m.projectInfoLines()
	case selection.FocusCategory:
		lines = m.categoryInfoLines(pos.CategoryIndex)
	case selection.FocusTask, selection.FocusDescription:
		lines = m.taskInfoLines(pos.CategoryIndex, pos.TaskIndex)
	}

	lines = append(lines, "", ui.RenderHints([]ui.Hint{{Key: "i/esc/q", Label: "close"}}))
	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}

func (m model) taskInfoLines(catIdx, taskIdx int) []string {
	task := m.project.Categories[catIdx].Tasks[taskIdx]
	category := m.project.Categories[catIdx]

	lines := []string{
		ui.DialogTitleStyle.Render("Task Info"),
		"",
	}

	const infoMaxWidth = 60
	const titleLabel = "Title:    "
	labelWidth := len(titleLabel)
	available := infoMaxWidth - labelWidth
	wrapped := ansi.Wrap(task.Title, available, "")
	titleLines := strings.Split(wrapped, "\n")
	indent := strings.Repeat(" ", labelWidth)
	for i, line := range titleLines {
		if i == 0 {
			lines = append(lines, titleLabel+line)
		} else {
			lines = append(lines, indent+line)
		}
	}

	lines = append(lines,
		fmt.Sprintf("Status:   %s", formatStatusLabel(task.Status)),
		fmt.Sprintf("Priority: %s", formatPriorityLabel(task.Priority)),
		fmt.Sprintf("Estimate: %s", FormatEstimateLabel(task.EstimateMinutes)),
		fmt.Sprintf("Category: %s", category.Name),
		"",
		fmt.Sprintf("Created:  %s", FormatDateWithRelative(task.CreatedAt)),
		fmt.Sprintf("Updated:  %s", FormatDateWithRelative(task.UpdatedAt)),
	)

	if task.CompletionDate != "" {
		lines = append(lines, fmt.Sprintf("Completed: %s", FormatDateWithRelative(task.CompletionDate)))
	}

	if task.Description != "" {
		lines = append(lines, "", ui.DialogTitleStyle.Render("Description"), "")
		descIndent := strings.Repeat(" ", 2)
		for _, srcLine := range strings.Split(task.Description, "\n") {
			if strings.TrimSpace(srcLine) == "" {
				lines = append(lines, "")
				continue
			}
			wrapped := ansi.Wrap(srcLine, infoMaxWidth-len(descIndent), "")
			for _, wline := range strings.Split(wrapped, "\n") {
				lines = append(lines, descIndent+wline)
			}
		}
	}

	return lines
}

func (m model) categoryInfoLines(catIdx int) []string {
	category := m.project.Categories[catIdx]
	counts := category.StatusCounts()

	lines := []string{
		ui.DialogTitleStyle.Render("Category Info"),
		"",
		fmt.Sprintf("Name:     %s", category.Name),
		fmt.Sprintf("Estimate: %s", FormatEstimateLabel(category.EstimateMinutes)),
		fmt.Sprintf("Created:  %s", FormatDateWithRelative(category.CreatedAt)),
	}

	if category.UpdatedAt != "" {
		lines = append(lines, fmt.Sprintf("Updated:  %s", FormatDateWithRelative(category.UpdatedAt)))
	}

	lines = append(lines,
		"",
		fmt.Sprintf("Total Tasks: %d", len(category.Tasks)),
		"",
	)
	lines = append(lines, statusBreakdownLines(counts)...)

	return lines
}

func (m model) projectInfoLines() []string {
	counts := m.project.StatusCounts()
	lines := []string{
		ui.DialogTitleStyle.Render("Project Info"),
		"",
		fmt.Sprintf("Name:       %s", m.project.Name),
		fmt.Sprintf("Created:    %s", FormatDateWithRelative(m.project.CreatedAt)),
		fmt.Sprintf("Updated:    %s", FormatDateWithRelative(m.project.UpdatedAt)),
		"",
		fmt.Sprintf("Categories: %d", len(m.project.Categories)),
		fmt.Sprintf("Total Tasks: %d", counts.Total()),
		"",
	}
	lines = append(lines, statusBreakdownLines(counts)...)
	return lines
}

func statusBreakdownLines(counts domain.StatusCounts) []string {
	return []string{
		"Task Breakdown:",
		fmt.Sprintf("  Todo:        %d", counts.Todo),
		fmt.Sprintf("  In Progress: %d", counts.InProgress),
		fmt.Sprintf("  Completed:   %d", counts.Completed),
		fmt.Sprintf("  Cancelled:   %d", counts.Cancelled),
	}
}

func formatStatusLabel(status string) string {
	switch status {
	case domain.StatusTodo:
		return "Todo"
	case domain.StatusInProgress:
		return "In Progress"
	case domain.StatusCompleted:
		return "Completed"
	case domain.StatusCancelled:
		return "Cancelled"
	default:
		return status
	}
}

func formatPriorityLabel(priority string) string {
	switch priority {
	case domain.PriorityHigh:
		return "High"
	case domain.PriorityMedium:
		return "Medium"
	case domain.PriorityLow:
		return "Low"
	case "":
		return "None"
	default:
		return priority
	}
}
