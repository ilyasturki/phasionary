package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"phasionary/internal/domain"
	"phasionary/internal/ui"
)

type TaskLineRenderer struct {
	width         int
	statusDisplay string
	priorityColor string
	focused       bool
	visualMode    bool
}

func NewTaskLineRenderer(width int, statusDisplay, priorityColor string, focused bool) *TaskLineRenderer {
	return &TaskLineRenderer{
		width:         width,
		statusDisplay: statusDisplay,
		priorityColor: priorityColor,
		focused:       focused,
	}
}

func (r *TaskLineRenderer) WithVisualMode(v bool) *TaskLineRenderer {
	r.visualMode = v
	return r
}

func (r *TaskLineRenderer) selectedStyle() lipgloss.Style {
	if r.visualMode {
		return ui.GetVisualSelectedStyle(r.focused)
	}
	return ui.GetSelectedStyle(r.focused)
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

// RenderDescription renders a task description as a standalone, focusable block
// indented to `indent` columns. When `selected`, the block uses the row-selection
// style; otherwise it renders muted to read as secondary text.
func (r *TaskLineRenderer) RenderDescription(description string, indent int, selected bool) string {
	if description == "" {
		return ""
	}
	indentStr := strings.Repeat(" ", indent)
	available := 0
	if r.width > 0 {
		available = r.width - indent
		if available < 1 {
			available = 0
		}
	}
	style := ui.MutedStyle
	if selected {
		style = r.selectedStyle()
	}

	// When selected, every visual line gets padded to the full row width so the
	// highlight rectangle is uniform — otherwise blank-line indents and short
	// paragraph tails produce a ragged selection band.
	shouldPad := selected && r.width > 0

	var out []string
	for _, paragraph := range strings.Split(description, "\n") {
		if paragraph == "" {
			out = append(out, style.Render(padDescriptionLine(indentStr, r.width, shouldPad)))
			continue
		}
		text := paragraph
		if available > 0 {
			text = ansi.Wrap(paragraph, available, "")
		}
		for _, l := range strings.Split(text, "\n") {
			out = append(out, style.Render(padDescriptionLine(indentStr+l, r.width, shouldPad)))
		}
	}
	return strings.Join(out, "\n")
}

func padDescriptionLine(s string, width int, pad bool) string {
	if !pad {
		return s
	}
	gap := width - ansi.StringWidth(s)
	if gap <= 0 {
		return s
	}
	return s + strings.Repeat(" ", gap)
}

func (r *TaskLineRenderer) renderUnselected(task domain.Task, prefix, priorityIcon string) string {
	status := r.formatStatus(task.Status, false)
	icon := ""
	if priorityIcon != "" {
		iconStyle := ui.PriorityIconStyle(task.Priority, r.priorityColor)
		if task.Status == domain.StatusCompleted || task.Status == domain.StatusCancelled {
			iconStyle = iconStyle.Faint(true)
		}
		icon = iconStyle.Render(priorityIcon) + " "
	}
	descMarker := r.formatDescriptionBadge(task.Description, false)
	estimate := r.formatEstimateBadge(task.EstimateMinutes, false)
	suffix := descMarker + estimate
	suffixText := r.descriptionBadgeText(task.Description) + r.estimateBadgeText(task.EstimateMinutes)
	titleStyle := ui.TaskTitleStyle(task.Priority, task.Status, r.priorityColor)
	prefixPart := fmt.Sprintf("%s[%s] %s", prefix, status, icon)

	if r.width <= 0 {
		return prefixPart + titleStyle.Render(task.Title) + suffix
	}

	return r.wrapTaskContentWithSuffix(task.Title, prefixPart, titleStyle, suffix, suffixText)
}

func (r *TaskLineRenderer) renderSelected(task domain.Task, prefix, priorityIcon string) string {
	statusText := r.statusLabel(task.Status)
	selectedStyle := r.selectedStyle()
	priorityStyle := selectedStyle
	statusStyle := selectedStyle

	icon := ""
	iconText := ""
	if priorityIcon != "" {
		icon = priorityStyle.Render(priorityIcon + " ")
		iconText = priorityIcon + " "
	}

	descMarker := r.formatDescriptionBadge(task.Description, true)
	estimate := r.formatEstimateBadge(task.EstimateMinutes, true)
	suffix := descMarker + estimate
	suffixText := r.descriptionBadgeText(task.Description) + r.estimateBadgeText(task.EstimateMinutes)

	prefixPart := selectedStyle.Render(prefix+"[") +
		statusStyle.Render(statusText) +
		selectedStyle.Render("] ") + icon

	if r.width <= 0 {
		return prefixPart + priorityStyle.Render(task.Title) + suffix
	}

	overhead := ansi.StringWidth(prefix + "[" + statusText + "] " + iconText)
	return r.wrapSelectedContentWithSuffix(task.Title, prefixPart, overhead, priorityStyle, suffix, suffixText)
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
	selectedStyle := r.selectedStyle()

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
		return r.selectedStyle().Render(text)
	}
	return ui.MutedStyle.Render(text)
}

func (r *TaskLineRenderer) estimateBadgeText(minutes int) string {
	if minutes == 0 {
		return ""
	}
	return " ~" + formatEstimateShort(minutes)
}

func (r *TaskLineRenderer) formatDescriptionBadge(description string, selected bool) string {
	text := r.descriptionBadgeText(description)
	if text == "" {
		return ""
	}
	if selected {
		return r.selectedStyle().Render(text)
	}
	return ui.MutedStyle.Render(text)
}

func (r *TaskLineRenderer) descriptionBadgeText(description string) string {
	if description == "" {
		return ""
	}
	return " ¶"
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
