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

func renderProjectLine(name string, selected, focused, filtered, visualMode bool, statusText string, width int) string {
	prefix := "  "
	if selected {
		prefix = "> "
	}
	line := fmt.Sprintf("%s■ %s", prefix, name)
	if selected {
		if visualMode {
			line = ui.GetVisualSelectedStyle(focused).Render(line)
		} else {
			line = ui.GetSelectedStyle(focused).Render(line)
		}
	} else {
		line = ui.HeaderStyle.Render(line)
	}
	if filtered {
		line += " " + ui.FilterTagStyle.Render("[filtered]")
	}
	if visualMode {
		line += " " + ui.VisualTagStyle.Render("[visual]")
	}
	if statusText != "" {
		line += " " + ui.StatusLineStyle.Render(statusText)
	}
	if width > 0 {
		line = ansi.Truncate(line, width, "")
	}
	return line
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

func renderCategoryLine(name string, estimateMinutes int, aggregateStatus string, selected bool, folded bool, width int, focused bool, visualMode bool) string {
	prefix := "  "
	if selected {
		prefix = "> "
	}
	style := ui.CategoryStyle
	if selected {
		if visualMode {
			style = ui.GetVisualSelectedStyle(focused)
		} else {
			style = ui.GetSelectedStyle(focused)
		}
	}

	foldIndicator := "▼ "
	if folded {
		foldIndicator = "▶ "
	}

	statusBadge := ""
	statusBadgeText := ""
	if aggregateStatus != "" {
		statusBadgeText = " [" + statusIcon(aggregateStatus) + "]"
		switch {
		case selected && visualMode:
			statusBadge = ui.GetVisualSelectedStyle(focused).Render(statusBadgeText)
		case selected:
			statusBadge = ui.GetSelectedStatusStyle(aggregateStatus, focused).Render(statusBadgeText)
		default:
			statusBadge = ui.StatusStyle(aggregateStatus).Render(statusBadgeText)
		}
	}

	estimateBadge := ""
	estimateBadgeText := ""
	if estimateMinutes > 0 {
		estimateBadgeText = " ~" + FormatEstimate(estimateMinutes)
		switch {
		case selected && visualMode:
			estimateBadge = ui.GetVisualSelectedStyle(focused).Render(estimateBadgeText)
		case selected:
			estimateBadge = ui.GetSelectedStyle(focused).Render(estimateBadgeText)
		default:
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

func (m model) renderTaskLine(task domain.Task, selected bool, width int, focused bool, visualMode bool) string {
	cfg := m.deps.CfgManager.Get()
	renderer := components.NewTaskLineRenderer(width, cfg.StatusDisplay, cfg.PriorityColor, focused).WithVisualMode(visualMode)
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
	cfg := m.deps.CfgManager.Get()
	statusText := formatStatus(task.Status, cfg.StatusDisplay)
	titleStyle := ui.PriorityStyle(task.Priority, cfg.PriorityColor)
	iconStyle := ui.PriorityIconStyle(task.Priority, cfg.PriorityColor)
	icon := ui.PriorityIcon(task.Priority)
	iconPrefix := ""
	iconPlain := ""
	if icon != "" {
		iconPrefix = iconStyle.Render(icon) + " "
		iconPlain = icon + " "
	}
	prefixPart := fmt.Sprintf("%s[%s] %s", prefix, statusText, iconPrefix)
	overhead := ansi.StringWidth(prefix + "[" + statusLabel(task.Status, cfg.StatusDisplay) + "] " + iconPlain)
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

func (m model) statusText() string {
	if m.ui.Screen.StatusMsg != "" {
		return m.ui.Screen.StatusMsg
	}
	if m.ui.Modes.IsVisual() {
		count := len(m.visualSelectedPositions())
		noun := plural(count, "task", "tasks")
		if m.ui.Visual.Kind == selection.FocusCategory {
			noun = plural(count, "category", "categories")
		}
		return fmt.Sprintf("-- VISUAL -- %d %s selected", count, noun)
	}
	return ""
}

func truncateText(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
}
