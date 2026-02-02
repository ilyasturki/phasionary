package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"phasionary/internal/app/components"
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
	split := splitAtCursor(m.ui.Edit.input.Value(), m.ui.Edit.input.Position())
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
	renderer := components.NewTaskLineRenderer(width, m.deps.CfgManager.Get().StatusDisplay)
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
	if m.ui.Edit.isAdding && m.ui.Edit.input.Value() == "" {
		cursorStyle := ui.SelectedStyle
		placeholder := ui.MutedStyle.Render("Enter category name...")
		styledText := cursorStyle.Render(" ") + placeholder
		if m.ui.Width > 0 {
			wrapped := wrapWithPrefix(styledText, m.ui.Width, prefixWidth, prefix)
			return strings.Join(wrapped.lines, "\n")
		}
		return prefix + styledText
	}
	return renderCursorLine(m.ui.Edit.input.Value(), m.ui.Edit.input.Position(), m.ui.Width, prefixWidth, prefix, ui.CategoryStyle, ui.SelectedStyle)
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
	if m.ui.Edit.isAdding && m.ui.Edit.input.Value() == "" {
		cursorStyle := ui.SelectedStyle
		placeholder := ui.MutedStyle.Render("Enter task title...")
		styledText := cursorStyle.Render(" ") + placeholder
		if m.ui.Width > 0 {
			available := safeWidth(m.ui.Width, overhead)
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
	if m.ui.Width <= 0 {
		split := splitAtCursor(edited, m.ui.Edit.input.Position())
		return prefixPart +
			titleStyle.Render(split.left) +
			ui.SelectedStyle.Render(split.cursorCh) +
			titleStyle.Render(split.right)
	}
	available := safeWidth(m.ui.Width, overhead)
	wrapped := ansi.Wrap(edited, available, "")
	wrapLines := strings.Split(wrapped, "\n")
	indent := strings.Repeat(" ", overhead)
	pos := 0
	var result []string
	cursor := m.ui.Edit.input.Position()
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
	filterIndicator := ""
	if m.ui.Filter.HasActiveFilter() {
		filterIndicator = " [filtered]"
	}
	if m.ui.StatusMsg != "" {
		return ui.StatusLineStyle.Render(m.ui.StatusMsg + filterIndicator)
	}
	position, ok := m.selectedPosition()
	if !ok {
		return ui.StatusLineStyle.Render("No items to display." + filterIndicator)
	}
	if position.Kind == focusProject {
		summary := fmt.Sprintf("Project: %s%s", m.project.Name, filterIndicator)
		return ui.StatusLineStyle.Render(summary)
	}
	category := m.project.Categories[position.CategoryIndex]
	if position.Kind == focusCategory {
		summary := fmt.Sprintf("Category: %s (%d tasks)%s", category.Name, len(category.Tasks), filterIndicator)
		return ui.StatusLineStyle.Render(summary)
	}
	task := category.Tasks[position.TaskIndex]
	summary := fmt.Sprintf("Selected: %s / %s (%s)%s", category.Name, task.Title, task.Status, filterIndicator)
	return ui.StatusLineStyle.Render(summary)
}

func (m model) shortcutsLine() string {
	if m.ui.Modes.IsEdit() {
		return ui.StatusLineStyle.Render("Shortcuts: enter save | esc cancel | arrows move cursor | ? help")
	}
	return ui.StatusLineStyle.Render("Shortcuts: j/k move | J/K reorder | s sort | f filter | a add task | A add category | enter edit | e external editor | space status | h/l priority | y copy | d delete | p projects | o options | ? help | q quit")
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
		ui.DialogHintStyle.Render("y/enter confirm | n/esc cancel"),
	}
	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}

func (m model) helpView() string {
	lines := []string{
		ui.DialogTitleStyle.Render("Navigation:"),
		"  j/k or ↑/↓    move selection",
		"  ctrl+d/u      half-page down/up",
		"  ctrl+f/b      full-page down/up",
		"  gg            jump to first item",
		"  G             jump to last item",
		"  zz            center selection on screen",
		"  p             switch project",
		"",
		ui.DialogTitleStyle.Render("Actions:"),
		"  a             add new task",
		"  A             add new category",
		"  enter         edit selected item",
		"  e             edit in external editor",
		"  space         toggle task status",
		"  J/K           reorder task/category up/down",
		"  s/S           sort tasks by status",
		"  f             filter tasks by status",
		"  h/l           change priority",
		"  y             copy selected text",
		"  d             delete selected item",
		"  o             options",
		"  ?             toggle help",
		"  q or ctrl+c   quit",
		"",
		ui.DialogTitleStyle.Render("Editing:"),
		"  enter         save changes",
		"  esc           cancel editing",
		"  ←/→           move cursor",
		"  ctrl+a/e      start/end of line",
		"  ctrl+w        delete word backward",
		"  ctrl+k/u      delete to end/start",
		"  ctrl+←/→      word navigation",
	}
	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}

func (m model) optionsView() string {
	statusValue := "Text Labels"
	if m.deps.CfgManager.Get().StatusDisplay == config.StatusDisplayIcons {
		statusValue = "Icons"
	}
	lines := []string{
		ui.DialogTitleStyle.Render("Options"),
		"",
		fmt.Sprintf("> Status Display: [%s]", statusValue),
		"",
		ui.DialogHintStyle.Render("space/tab toggle | q/esc/enter close"),
	}
	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}

func (m model) filterView() string {
	statusLabels := map[string]string{
		domain.StatusTodo:       "Todo",
		domain.StatusInProgress: "In Progress",
		domain.StatusCompleted:  "Completed",
		domain.StatusCancelled:  "Cancelled",
	}
	lines := []string{ui.DialogTitleStyle.Render("Filter by Status:"), ""}
	for i, status := range filterStatuses {
		prefix := "  "
		if i == m.ui.Filter.Selected() {
			prefix = "> "
		}
		checkbox := "[ ]"
		if m.ui.Filter.IsEnabled(status) {
			checkbox = "[x]"
		}
		line := fmt.Sprintf("%s%s %s", prefix, checkbox, statusLabels[status])
		if i == m.ui.Filter.Selected() {
			line = ui.SelectedStyle.Render(line)
		}
		lines = append(lines, line)
	}
	lines = append(lines, "", ui.DialogHintStyle.Render("j/k navigate | space toggle | q/esc/f close"))
	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}

func (m model) projectPickerView() string {
	lines := []string{ui.DialogTitleStyle.Render("Select Project:"), ""}

	visibleEnd := m.ui.Picker.scrollOffset + pickerVisibleItems
	if visibleEnd > len(m.ui.Picker.projects) {
		visibleEnd = len(m.ui.Picker.projects)
	}

	if m.ui.Picker.scrollOffset > 0 {
		lines = append(lines, ui.DialogHintStyle.Render("  ↑ more above"))
	}

	for i := m.ui.Picker.scrollOffset; i < visibleEnd; i++ {
		p := m.ui.Picker.projects[i]
		prefix := "  "
		if i == m.ui.Picker.selected {
			prefix = "> "
		}
		name := p.Name
		if p.ID == m.project.ID {
			name += " (current)"
		}
		line := prefix + name
		if i == m.ui.Picker.selected {
			line = ui.SelectedStyle.Render(line)
		} else if p.ID == m.project.ID {
			line = prefix + p.Name + ui.DialogHintStyle.Render(" (current)")
		}
		lines = append(lines, line)
	}

	if visibleEnd < len(m.ui.Picker.projects) {
		lines = append(lines, ui.DialogHintStyle.Render("  ↓ more below"))
	}

	lines = append(lines, "", ui.DialogHintStyle.Render("j/k navigate | enter select | esc cancel"))

	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}
