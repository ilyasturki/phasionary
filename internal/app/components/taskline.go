package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"phasionary/internal/domain"
	"phasionary/internal/ui"
)

type TaskLineRenderer struct {
	width         int
	statusDisplay string
}

func NewTaskLineRenderer(width int, statusDisplay string) *TaskLineRenderer {
	return &TaskLineRenderer{
		width:         width,
		statusDisplay: statusDisplay,
	}
}

func (r *TaskLineRenderer) Render(task domain.Task, selected bool) string {
	prefix := "  "
	if selected {
		prefix = "> "
	}
	priorityIcon := ui.PriorityIcon(task.Priority)

	if selected {
		return r.renderSelected(task, prefix, priorityIcon)
	}
	return r.renderUnselected(task, prefix, priorityIcon)
}

func (r *TaskLineRenderer) renderUnselected(task domain.Task, prefix, priorityIcon string) string {
	status := r.formatStatus(task.Status, false)
	icon := ""
	if priorityIcon != "" {
		icon = ui.PriorityStyle(task.Priority).Render(priorityIcon) + " "
	}
	titleStyle := ui.PriorityStyle(task.Priority)
	prefixPart := fmt.Sprintf("%s[%s] %s", prefix, status, icon)

	if r.width <= 0 {
		return prefixPart + titleStyle.Render(task.Title)
	}

	return r.wrapTaskContent(task.Title, prefixPart, titleStyle, nil)
}

func (r *TaskLineRenderer) renderSelected(task domain.Task, prefix, priorityIcon string) string {
	statusText := r.statusLabel(task.Status)
	priorityStyle := ui.SelectedPriorityStyle(task.Priority)
	statusStyle := ui.SelectedStatusStyle(task.Status)

	icon := ""
	iconText := ""
	if priorityIcon != "" {
		icon = priorityStyle.Render(priorityIcon + " ")
		iconText = priorityIcon + " "
	}

	prefixPart := ui.SelectedStyle.Render(prefix+"[") +
		statusStyle.Render(statusText) +
		ui.SelectedStyle.Render("] ") + icon

	if r.width <= 0 {
		return prefixPart + priorityStyle.Render(task.Title)
	}

	overhead := ansi.StringWidth(prefix + "[" + statusText + "] " + iconText)
	return r.wrapSelectedContent(task.Title, prefixPart, overhead, priorityStyle)
}

func (r *TaskLineRenderer) wrapTaskContent(title, prefixPart string, titleStyle lipgloss.Style, _ interface{}) string {
	overhead := ansi.StringWidth(prefixPart)
	available := safeWidth(r.width, overhead)
	wrapped := ansi.Wrap(title, available, "")
	wrapLines := strings.Split(wrapped, "\n")
	indent := strings.Repeat(" ", overhead)

	var result []string
	for i, line := range wrapLines {
		styledLine := titleStyle.Render(line)
		if i == 0 {
			result = append(result, prefixPart+styledLine)
		} else {
			result = append(result, indent+styledLine)
		}
	}
	return strings.Join(result, "\n")
}

func (r *TaskLineRenderer) wrapSelectedContent(title, prefixPart string, overhead int, titleStyle lipgloss.Style) string {
	available := safeWidth(r.width, overhead)
	wrapped := ansi.Wrap(title, available, "")
	wrapLines := strings.Split(wrapped, "\n")
	indent := strings.Repeat(" ", overhead)

	var result []string
	for i, line := range wrapLines {
		styledTitle := titleStyle.Render(line)
		if i == 0 {
			result = append(result, prefixPart+styledTitle)
		} else {
			styledIndent := ui.SelectedStyle.Render(indent)
			result = append(result, styledIndent+styledTitle)
		}
	}
	return strings.Join(result, "\n")
}

func (r *TaskLineRenderer) formatStatus(status string, selected bool) string {
	label := r.statusLabel(status)
	if selected {
		return label
	}
	return ui.StatusStyle(status).Render(label)
}

func (r *TaskLineRenderer) statusLabel(status string) string {
	if r.statusDisplay == "icons" {
		return statusIcon(status)
	}
	switch status {
	case domain.StatusInProgress:
		return " progress"
	case domain.StatusCompleted:
		return "completed"
	case domain.StatusCancelled:
		return "cancelled"
	default:
		return "  todo   "
	}
}

func statusIcon(status string) string {
	switch status {
	case domain.StatusInProgress:
		return "/"
	case domain.StatusCompleted:
		return "x"
	case domain.StatusCancelled:
		return "-"
	default:
		return " "
	}
}

func safeWidth(totalWidth, overhead int) int {
	available := totalWidth - overhead
	if available < 1 {
		return 1
	}
	return available
}
