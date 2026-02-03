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
	focused       bool
}

func NewTaskLineRenderer(width int, statusDisplay string, focused bool) *TaskLineRenderer {
	return &TaskLineRenderer{
		width:         width,
		statusDisplay: statusDisplay,
		focused:       focused,
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
		icon = ui.TaskTitleStyle(task.Priority, task.Status).Render(priorityIcon) + " "
	}
	estimate := r.formatEstimateBadge(task.EstimateMinutes, false)
	titleStyle := ui.TaskTitleStyle(task.Priority, task.Status)
	prefixPart := fmt.Sprintf("%s[%s] %s", prefix, status, icon)

	if r.width <= 0 {
		return prefixPart + titleStyle.Render(task.Title) + estimate
	}

	return r.wrapTaskContentWithSuffix(task.Title, prefixPart, titleStyle, estimate, r.estimateBadgeText(task.EstimateMinutes))
}

func (r *TaskLineRenderer) renderSelected(task domain.Task, prefix, priorityIcon string) string {
	statusText := r.statusLabel(task.Status)
	priorityStyle := ui.GetSelectedPriorityStyle(task.Priority, r.focused)
	statusStyle := ui.GetSelectedStatusStyle(task.Status, r.focused)
	selectedStyle := ui.GetSelectedStyle(r.focused)

	icon := ""
	iconText := ""
	if priorityIcon != "" {
		icon = priorityStyle.Render(priorityIcon + " ")
		iconText = priorityIcon + " "
	}

	estimate := r.formatEstimateBadge(task.EstimateMinutes, true)
	estimateText := r.estimateBadgeText(task.EstimateMinutes)

	prefixPart := selectedStyle.Render(prefix+"[") +
		statusStyle.Render(statusText) +
		selectedStyle.Render("] ") + icon

	if r.width <= 0 {
		return prefixPart + priorityStyle.Render(task.Title) + estimate
	}

	overhead := ansi.StringWidth(prefix + "[" + statusText + "] " + iconText)
	return r.wrapSelectedContentWithSuffix(task.Title, prefixPart, overhead, priorityStyle, estimate, estimateText)
}

func (r *TaskLineRenderer) wrapTaskContentWithSuffix(title, prefixPart string, titleStyle lipgloss.Style, suffix, suffixText string) string {
	overhead := ansi.StringWidth(prefixPart)
	suffixWidth := ansi.StringWidth(suffixText)
	available := safeWidth(r.width, overhead+suffixWidth)
	wrapped := ansi.Wrap(title, available, "")
	wrapLines := strings.Split(wrapped, "\n")
	indent := strings.Repeat(" ", overhead)

	var result []string
	for i, line := range wrapLines {
		styledLine := titleStyle.Render(line)
		if i == 0 {
			result = append(result, prefixPart+styledLine+suffix)
		} else {
			result = append(result, indent+styledLine)
		}
	}
	return strings.Join(result, "\n")
}

func (r *TaskLineRenderer) wrapSelectedContentWithSuffix(title, prefixPart string, overhead int, titleStyle lipgloss.Style, suffix, suffixText string) string {
	suffixWidth := ansi.StringWidth(suffixText)
	available := safeWidth(r.width, overhead+suffixWidth)
	wrapped := ansi.Wrap(title, available, "")
	wrapLines := strings.Split(wrapped, "\n")
	indent := strings.Repeat(" ", overhead)
	selectedStyle := ui.GetSelectedStyle(r.focused)

	var result []string
	for i, line := range wrapLines {
		styledTitle := titleStyle.Render(line)
		if i == 0 {
			result = append(result, prefixPart+styledTitle+suffix)
		} else {
			styledIndent := selectedStyle.Render(indent)
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

func (r *TaskLineRenderer) formatEstimateBadge(minutes int, selected bool) string {
	text := r.estimateBadgeText(minutes)
	if text == "" {
		return ""
	}
	if selected {
		return ui.GetSelectedStyle(r.focused).Render(text)
	}
	return ui.MutedStyle.Render(text)
}

func (r *TaskLineRenderer) estimateBadgeText(minutes int) string {
	if minutes == 0 {
		return ""
	}
	return " ~" + formatEstimateShort(minutes)
}

func formatEstimateShort(minutes int) string {
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	hours := minutes / 60
	if hours < 8 {
		return fmt.Sprintf("%dh", hours)
	}
	days := hours / 8
	return fmt.Sprintf("%dd", days)
}
