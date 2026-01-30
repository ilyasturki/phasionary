package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"phasionary/internal/config"
	"phasionary/internal/domain"
	"phasionary/internal/ui"
)

func renderProjectLine(name string, selected bool) string {
	prefix := "  "
	if selected {
		prefix = "> "
	}
	line := fmt.Sprintf("%s■ %s", prefix, name)
	if selected {
		return ui.SelectedStyle.Render(line)
	}
	return ui.HeaderStyle.Render(line)
}

func (m model) renderEditProjectLine() string {
	prefix := "> "
	icon := "■ "
	split := splitAtCursor(m.edit.input.Value(), m.edit.input.Position())
	cursorStyle := ui.SelectedStyle
	return fmt.Sprintf(
		"%s%s%s%s%s",
		prefix,
		ui.HeaderStyle.Render(icon),
		ui.HeaderStyle.Render(split.left),
		cursorStyle.Render(split.cursorCh),
		ui.HeaderStyle.Render(split.right),
	)
}

func renderCategoryLine(name string, selected bool, width int) string {
	prefix := "  "
	if selected {
		prefix = "> "
	}
	style := ui.CategoryStyle
	if selected {
		style = ui.SelectedStyle
	}
	if width <= 0 {
		return style.Render(prefix + name)
	}
	wrapped := wrapWithPrefix(name, width, prefixWidth, prefix)
	return style.Render(strings.Join(wrapped.lines, "\n"))
}

func (m model) renderTaskLine(task domain.Task, selected bool, width int) string {
	prefix := "  "
	if selected {
		prefix = "> "
	}
	priorityIcon := ui.PriorityIcon(task.Priority)
	if !selected {
		return m.renderUnselectedTask(task, prefix, priorityIcon, width)
	}
	return m.renderSelectedTask(task, prefix, priorityIcon, width)
}

func (m model) renderUnselectedTask(task domain.Task, prefix, priorityIcon string, width int) string {
	status := formatStatus(task.Status, m.cfgManager.Get().StatusDisplay)
	icon := ""
	if priorityIcon != "" {
		icon = ui.PriorityStyle(task.Priority).Render(priorityIcon) + " "
	}
	titleStyle := ui.PriorityStyle(task.Priority)
	prefixPart := fmt.Sprintf("%s[%s] %s", prefix, status, icon)
	if width <= 0 {
		return prefixPart + titleStyle.Render(task.Title)
	}
	overhead := ansi.StringWidth(prefixPart)
	available := safeWidth(width, overhead)
	wrapped := ansi.Wrap(task.Title, available, "")
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

func (m model) renderSelectedTask(task domain.Task, prefix, priorityIcon string, width int) string {
	statusText := statusLabel(task.Status, m.cfgManager.Get().StatusDisplay)
	priorityStyle := ui.SelectedPriorityStyle(task.Priority)
	statusStyle := ui.SelectedStatusStyle(task.Status)
	icon := ""
	iconText := ""
	if priorityIcon != "" {
		icon = priorityStyle.Render(priorityIcon + " ")
		iconText = priorityIcon + " "
	}
	if width <= 0 {
		title := priorityStyle.Render(task.Title)
		return ui.SelectedStyle.Render(prefix+"[") +
			statusStyle.Render(statusText) +
			ui.SelectedStyle.Render("] ") +
			icon + title
	}
	overhead := ansi.StringWidth(prefix + "[" + statusText + "] " + iconText)
	available := safeWidth(width, overhead)
	wrapped := ansi.Wrap(task.Title, available, "")
	wrapLines := strings.Split(wrapped, "\n")
	indent := strings.Repeat(" ", overhead)
	var result []string
	for i, line := range wrapLines {
		styledTitle := priorityStyle.Render(line)
		if i == 0 {
			firstLine := ui.SelectedStyle.Render(prefix+"[") +
				statusStyle.Render(statusText) +
				ui.SelectedStyle.Render("] ") +
				icon + styledTitle
			result = append(result, firstLine)
		} else {
			styledIndent := ui.SelectedStyle.Render(indent)
			result = append(result, styledIndent+styledTitle)
		}
	}
	return strings.Join(result, "\n")
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
	if m.edit.isAdding && m.edit.input.Value() == "" {
		cursorStyle := ui.SelectedStyle
		placeholder := ui.MutedStyle.Render("Enter category name...")
		styledText := cursorStyle.Render(" ") + placeholder
		if m.width > 0 {
			wrapped := wrapWithPrefix(styledText, m.width, prefixWidth, prefix)
			return strings.Join(wrapped.lines, "\n")
		}
		return prefix + styledText
	}
	return renderCursorLine(m.edit.input.Value(), m.edit.input.Position(), m.width, prefixWidth, prefix, ui.CategoryStyle, ui.SelectedStyle)
}

func (m model) renderEditTaskLine(task domain.Task) string {
	prefix := "> "
	statusText := formatStatus(task.Status, m.cfgManager.Get().StatusDisplay)
	titleStyle := ui.PriorityStyle(task.Priority)
	icon := ui.PriorityIcon(task.Priority)
	iconPrefix := ""
	iconPlain := ""
	if icon != "" {
		iconPrefix = titleStyle.Render(icon) + " "
		iconPlain = icon + " "
	}
	prefixPart := fmt.Sprintf("%s[%s] %s", prefix, statusText, iconPrefix)
	overhead := ansi.StringWidth(prefix + "[" + statusLabel(task.Status, m.cfgManager.Get().StatusDisplay) + "] " + iconPlain)
	if m.edit.isAdding && m.edit.input.Value() == "" {
		cursorStyle := ui.SelectedStyle
		placeholder := ui.MutedStyle.Render("Enter task title...")
		styledText := cursorStyle.Render(" ") + placeholder
		if m.width > 0 {
			available := safeWidth(m.width, overhead)
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
	edited := m.edit.input.Value()
	if edited == "" {
		edited = " "
	}
	if m.width <= 0 {
		split := splitAtCursor(edited, m.edit.input.Position())
		return prefixPart +
			titleStyle.Render(split.left) +
			ui.SelectedStyle.Render(split.cursorCh) +
			titleStyle.Render(split.right)
	}
	available := safeWidth(m.width, overhead)
	wrapped := ansi.Wrap(edited, available, "")
	wrapLines := strings.Split(wrapped, "\n")
	indent := strings.Repeat(" ", overhead)
	pos := 0
	var result []string
	cursor := m.edit.input.Position()
	for i, line := range wrapLines {
		lineRunes := []rune(line)
		lineLen := len(lineRunes)
		var styledLine string
		if cursor >= pos && cursor < pos+lineLen {
			offset := cursor - pos
			l := string(lineRunes[:offset])
			c := string(lineRunes[offset])
			r := string(lineRunes[offset+1:])
			styledLine = titleStyle.Render(l) + ui.SelectedStyle.Render(c) + titleStyle.Render(r)
		} else if cursor == pos+lineLen {
			styledLine = titleStyle.Render(line) + ui.SelectedStyle.Render(" ")
		} else {
			styledLine = titleStyle.Render(line)
		}
		if i == 0 {
			result = append(result, prefixPart+styledLine)
		} else {
			result = append(result, indent+styledLine)
		}
		pos += lineLen + 1
	}
	return strings.Join(result, "\n")
}

func (m model) statusLine() string {
	if m.statusMsg != "" {
		return ui.StatusLineStyle.Render(m.statusMsg)
	}
	position, ok := m.selectedPosition()
	if !ok {
		return ui.StatusLineStyle.Render("No items to display.")
	}
	if position.Kind == focusProject {
		summary := fmt.Sprintf("Project: %s", m.project.Name)
		return ui.StatusLineStyle.Render(summary)
	}
	category := m.project.Categories[position.CategoryIndex]
	if position.Kind == focusCategory {
		summary := fmt.Sprintf("Category: %s (%d tasks)", category.Name, len(category.Tasks))
		return ui.StatusLineStyle.Render(summary)
	}
	task := category.Tasks[position.TaskIndex]
	summary := fmt.Sprintf("Selected: %s / %s (%s)", category.Name, task.Title, task.Status)
	return ui.StatusLineStyle.Render(summary)
}

func (m model) shortcutsLine() string {
	if m.mode == ModeEdit {
		return ui.StatusLineStyle.Render("Shortcuts: enter save | esc cancel | arrows move cursor | ? help")
	}
	return ui.StatusLineStyle.Render("Shortcuts: j/k move | J/K reorder | a add task | A add category | enter edit | space status | h/l priority | y copy | d delete | o options | ? help | q quit")
}

func placeOverlay(bg, fg string, width, height int) string {
	bgLines := strings.Split(bg, "\n")
	fgLines := strings.Split(fg, "\n")
	fgW := lipgloss.Width(fg)
	fgH := len(fgLines)
	startY := max(0, (height-fgH)/2)
	startX := max(0, (width-fgW)/2)
	for i, fgLine := range fgLines {
		y := startY + i
		if y >= len(bgLines) {
			break
		}
		left := ansi.Truncate(bgLines[y], startX, "")
		if w := ansi.StringWidth(left); w < startX {
			left += strings.Repeat(" ", startX-w)
		}
		right := ansi.TruncateLeft(bgLines[y], startX+fgW, "")
		bgLines[y] = left + fgLine + right
	}
	return strings.Join(bgLines, "\n")
}

func truncateText(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
}

func (m model) confirmDeleteView() string {
	position, ok := m.selectedPosition()
	if !ok || position.Kind == focusProject {
		return ""
	}
	var message string
	if position.Kind == focusTask {
		task := m.project.Categories[position.CategoryIndex].Tasks[position.TaskIndex]
		message = fmt.Sprintf("Delete task %q?", truncateText(task.Title, 30))
	} else {
		cat := m.project.Categories[position.CategoryIndex]
		message = fmt.Sprintf("Delete category %q and %d tasks?", truncateText(cat.Name, 30), len(cat.Tasks))
	}
	lines := []string{
		message,
		"",
		"y/enter confirm | n/esc cancel",
	}
	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}

func (m model) helpView() string {
	lines := []string{
		"Navigation:",
		"  j/k or ↑/↓    move selection",
		"  ctrl+d/u      half-page down/up",
		"  ctrl+f/b      full-page down/up",
		"  gg            jump to first item",
		"  G             jump to last item",
		"  zz            center selection on screen",
		"",
		"Actions:",
		"  a             add new task",
		"  A             add new category",
		"  enter         edit selected item",
		"  space         toggle task status",
		"  J/K           reorder task/category up/down",
		"  h/l           change priority",
		"  y             copy selected text",
		"  d             delete selected item",
		"  o             options",
		"  ?             toggle help",
		"  q or ctrl+c   quit",
		"",
		"Editing:",
		"  enter         save changes",
		"  esc           cancel editing",
		"  ←/→           move cursor",
		"  ctrl+a/e      start/end of line",
		"  ctrl+w        delete word backward",
		"  ctrl+k/u      delete to end/start",
		"  ctrl+←/→      word navigation",
	}
	return ui.HelpDialogStyle.Render(ui.MutedStyle.Render(strings.Join(lines, "\n")))
}

func (m model) optionsView() string {
	statusValue := "Text Labels"
	if m.cfgManager.Get().StatusDisplay == config.StatusDisplayIcons {
		statusValue = "Icons"
	}
	lines := []string{
		"Options",
		"",
		fmt.Sprintf("> Status Display: [%s]", statusValue),
		"",
		"space/tab toggle | q/esc/enter close",
	}
	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}
