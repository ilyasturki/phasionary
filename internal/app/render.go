package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"phasionary/internal/app/components"
	"phasionary/internal/app/selection"
	"phasionary/internal/config"
	"phasionary/internal/domain"
	"phasionary/internal/ui"
)

func renderProjectLine(name string, selected bool, focused bool) string {
	prefix := "  "
	if selected {
		prefix = "> "
	}
	line := fmt.Sprintf("%s■ %s", prefix, name)
	if selected {
		return ui.GetSelectedStyle(focused).Render(line)
	}
	return ui.HeaderStyle.Render(line)
}

func (m model) renderEditProjectLine() string {
	prefix := "> "
	icon := "■ "
	split := splitAtCursor(m.ui.Edit.input.Value(), m.ui.Edit.input.Position())
	cursorStyle := ui.GetCursorStyle(m.ui.Screen.WindowFocused)
	return fmt.Sprintf(
		"%s%s%s%s%s",
		prefix,
		ui.HeaderStyle.Render(icon),
		ui.HeaderStyle.Render(split.left),
		cursorStyle.Render(split.cursorCh),
		ui.HeaderStyle.Render(split.right),
	)
}

func renderCategoryLine(name string, estimateMinutes int, aggregateStatus string, selected bool, folded bool, width int, focused bool) string {
	prefix := "  "
	if selected {
		prefix = "> "
	}
	style := ui.CategoryStyle
	if selected {
		style = ui.GetSelectedStyle(focused)
	}

	foldIndicator := "▼ "
	if folded {
		foldIndicator = "▶ "
	}

	statusBadge := ""
	statusBadgeText := ""
	if aggregateStatus != "" {
		statusBadgeText = " [" + statusIcon(aggregateStatus) + "]"
		if selected {
			statusBadge = ui.GetSelectedStatusStyle(aggregateStatus, focused).Render(statusBadgeText)
		} else {
			statusBadge = ui.StatusStyle(aggregateStatus).Render(statusBadgeText)
		}
	}

	estimateBadge := ""
	estimateBadgeText := ""
	if estimateMinutes > 0 {
		estimateBadgeText = " ~" + FormatEstimate(estimateMinutes)
		if selected {
			estimateBadge = ui.GetSelectedStyle(focused).Render(estimateBadgeText)
		} else {
			estimateBadge = ui.MutedStyle.Render(estimateBadgeText)
		}
	}

	suffix := statusBadge + estimateBadge

	if width <= 0 {
		return style.Render(prefix+foldIndicator+name) + suffix
	}

	suffixWidth := len(statusBadgeText) + len(estimateBadgeText)
	foldWidth := 2
	available := safeWidth(width, prefixWidth+foldWidth+suffixWidth)
	wrapped := ansi.Wrap(name, available, "")
	lines := strings.Split(wrapped, "\n")
	indent := strings.Repeat(" ", prefixWidth+foldWidth)

	var result []string
	for i, line := range lines {
		styledLine := style.Render(line)
		if i == 0 {
			result = append(result, style.Render(prefix+foldIndicator)+styledLine+suffix)
		} else {
			result = append(result, style.Render(indent)+styledLine)
		}
	}
	return strings.Join(result, "\n")
}

func (m model) renderTaskLine(task domain.Task, selected bool, width int, focused bool) string {
	renderer := components.NewTaskLineRenderer(width, m.deps.CfgManager.Get().StatusDisplay, focused)
	return renderer.Render(task, selected)
}

func statusLabel(status, displayMode string) string {
	if displayMode == config.StatusDisplayIcons {
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

func formatStatus(status, displayMode string) string {
	return ui.StatusStyle(status).Render(statusLabel(status, displayMode))
}

func (m model) renderEditCategoryLine() string {
	prefix := "> "
	cursorStyle := ui.GetCursorStyle(m.ui.Screen.WindowFocused)
	if m.ui.Edit.isAdding && m.ui.Edit.input.Value() == "" {
		placeholder := ui.MutedStyle.Render("Enter category name...")
		styledText := cursorStyle.Render(" ") + placeholder
		if m.ui.Screen.Width > 0 {
			wrapped := wrapWithPrefix(styledText, m.ui.Screen.Width, prefixWidth, prefix)
			return strings.Join(wrapped.lines, "\n")
		}
		return prefix + styledText
	}
	return renderCursorLine(m.ui.Edit.input.Value(), m.ui.Edit.input.Position(), m.ui.Screen.Width, prefixWidth, prefix, ui.CategoryStyle, cursorStyle)
}

func (m model) renderEditTaskLine(task domain.Task) string {
	prefix := "> "
	statusText := formatStatus(task.Status, m.deps.CfgManager.Get().StatusDisplay)
	titleStyle := ui.PriorityStyle(task.Priority)
	icon := ui.PriorityIcon(task.Priority)
	iconPrefix := ""
	iconPlain := ""
	if icon != "" {
		iconPrefix = titleStyle.Render(icon) + " "
		iconPlain = icon + " "
	}
	prefixPart := fmt.Sprintf("%s[%s] %s", prefix, statusText, iconPrefix)
	overhead := ansi.StringWidth(prefix + "[" + statusLabel(task.Status, m.deps.CfgManager.Get().StatusDisplay) + "] " + iconPlain)
	cursorStyle := ui.GetCursorStyle(m.ui.Screen.WindowFocused)
	if m.ui.Edit.isAdding && m.ui.Edit.input.Value() == "" {
		placeholder := ui.MutedStyle.Render("Enter task title...")
		styledText := cursorStyle.Render(" ") + placeholder
		if m.ui.Screen.Width > 0 {
			available := safeWidth(m.ui.Screen.Width, overhead)
			wrapped := ansi.Wrap(styledText, available, "")
			lines := strings.Split(wrapped, "\n")
			indent := strings.Repeat(" ", overhead)
			for i, line := range lines {
				if i == 0 {
					lines[i] = prefixPart + line
				} else {
					lines[i] = indent + line
				}
			}
			return strings.Join(lines, "\n")
		}
		return prefixPart + styledText
	}
	edited := m.ui.Edit.input.Value()
	if edited == "" {
		edited = " "
	}
	return renderCursorLine(edited, m.ui.Edit.input.Position(), m.ui.Screen.Width, overhead, prefixPart, titleStyle, cursorStyle)
}

func (m model) statusLine() string {
	filterIndicator := ""
	if m.ui.Filter.HasActiveFilter() {
		filterIndicator = " [filtered]"
	}
	if m.ui.Screen.StatusMsg != "" {
		return ui.StatusLineStyle.Render(m.ui.Screen.StatusMsg + filterIndicator)
	}
	position, ok := m.selectedPosition()
	if !ok {
		return ui.StatusLineStyle.Render("No items to display." + filterIndicator)
	}
	if position.Kind == selection.FocusProject {
		summary := fmt.Sprintf("Project: %s%s", m.project.Name, filterIndicator)
		return ui.StatusLineStyle.Render(summary)
	}
	category := m.project.Categories[position.CategoryIndex]
	if position.Kind == selection.FocusCategory {
		summary := fmt.Sprintf("Category: %s (%d tasks)%s", category.Name, len(category.Tasks), filterIndicator)
		return ui.StatusLineStyle.Render(summary)
	}
	task := category.Tasks[position.TaskIndex]
	summary := fmt.Sprintf("Selected: %s / %s (%s)%s", category.Name, task.Title, task.Status, filterIndicator)
	return ui.StatusLineStyle.Render(summary)
}

func truncateText(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
}
