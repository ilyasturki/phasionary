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
	cursorStyle := ui.GetCursorStyle(m.ui.WindowFocused)
	return fmt.Sprintf(
		"%s%s%s%s%s",
		prefix,
		ui.HeaderStyle.Render(icon),
		ui.HeaderStyle.Render(split.left),
		cursorStyle.Render(split.cursorCh),
		ui.HeaderStyle.Render(split.right),
	)
}

func renderCategoryLine(name string, estimateMinutes int, selected bool, folded bool, width int, focused bool) string {
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

	if width <= 0 {
		return style.Render(prefix+foldIndicator+name) + estimateBadge
	}

	suffixWidth := len(estimateBadgeText)
	foldWidth := 2
	available := safeWidth(width, prefixWidth+foldWidth+suffixWidth)
	wrapped := ansi.Wrap(name, available, "")
	lines := strings.Split(wrapped, "\n")
	indent := strings.Repeat(" ", prefixWidth+foldWidth)

	var result []string
	for i, line := range lines {
		styledLine := style.Render(line)
		if i == 0 {
			result = append(result, style.Render(prefix+foldIndicator)+styledLine+estimateBadge)
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
	cursorStyle := ui.GetCursorStyle(m.ui.WindowFocused)
	if m.ui.Edit.isAdding && m.ui.Edit.input.Value() == "" {
		placeholder := ui.MutedStyle.Render("Enter category name...")
		styledText := cursorStyle.Render(" ") + placeholder
		if m.ui.Width > 0 {
			wrapped := wrapWithPrefix(styledText, m.ui.Width, prefixWidth, prefix)
			return strings.Join(wrapped.lines, "\n")
		}
		return prefix + styledText
	}
	return renderCursorLine(m.ui.Edit.input.Value(), m.ui.Edit.input.Position(), m.ui.Width, prefixWidth, prefix, ui.CategoryStyle, cursorStyle)
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
	cursorStyle := ui.GetCursorStyle(m.ui.WindowFocused)
	if m.ui.Edit.isAdding && m.ui.Edit.input.Value() == "" {
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
			cursorStyle.Render(split.cursorCh) +
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
			styledLine = titleStyle.Render(l) + cursorStyle.Render(c) + titleStyle.Render(r)
		} else if cursor == pos+lineLen {
			styledLine = titleStyle.Render(line) + cursorStyle.Render(" ")
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
	return ui.StatusLineStyle.Render("Shortcuts: j/k move | J/K reorder | s sort | f filter | Tab fold | a add task | A add category | enter edit | e external editor | space status | h/l priority | t estimate | y copy | x cut | p paste | d delete | i info | P projects | o options | ? help | q quit")
}

func truncateText(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
}

func (m model) confirmDeleteView() string {
	if m.ui.Picker.pendingDeleteID != "" {
		return m.confirmDeleteProjectView()
	}

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

func (m model) confirmDeleteProjectView() string {
	var projectName string
	for _, p := range m.ui.Picker.projects {
		if p.ID == m.ui.Picker.pendingDeleteID {
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

func (m model) helpView() string {
	lines := []string{
		ui.DialogTitleStyle.Render("Navigation:"),
		"  j/k or ↑/↓    move selection",
		"  ctrl+d/u      half-page down/up",
		"  ctrl+f/b      full-page down/up",
		"  gg            jump to first item",
		"  G             jump to last item",
		"  zz            center selection on screen",
		"  Tab/za        fold/unfold category",
		"  zc            fold all categories",
		"  zo            unfold all categories",
		"  P             switch project",
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
		"  t             set time estimate",
		"  y             copy selected text",
		"  x             mark task for cut",
		"  p             paste cut task",
		"  d             delete selected item",
		"  i             show item info",
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

	total := m.ui.Picker.totalItems()
	visibleEnd := m.ui.Picker.scrollOffset + pickerVisibleItems
	if visibleEnd > total {
		visibleEnd = total
	}

	if m.ui.Picker.scrollOffset > 0 {
		lines = append(lines, ui.DialogHintStyle.Render("  ↑ more above"))
	}

	for i := m.ui.Picker.scrollOffset; i < visibleEnd; i++ {
		isSelected := i == m.ui.Picker.selected
		if i == len(m.ui.Picker.projects) {
			lines = append(lines, m.renderAddProjectLine(isSelected))
			continue
		}
		p := m.ui.Picker.projects[i]
		prefix := "  "
		if isSelected {
			prefix = "> "
		}
		name := p.Name
		if p.ID == m.project.ID {
			name += " (current)"
		}
		line := prefix + name
		if isSelected {
			line = ui.SelectedStyle.Render(line)
		} else if p.ID == m.project.ID {
			line = prefix + p.Name + ui.DialogHintStyle.Render(" (current)")
		}
		lines = append(lines, line)
	}

	if visibleEnd < total {
		lines = append(lines, ui.DialogHintStyle.Render("  ↓ more below"))
	}

	hintText := "j/k navigate | enter select | d delete | esc cancel"
	if m.ui.Picker.isAdding {
		hintText = "enter create | esc cancel"
	}
	lines = append(lines, "", ui.DialogHintStyle.Render(hintText))

	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}

func (m model) renderAddProjectLine(isSelected bool) string {
	prefix := "  "
	if isSelected {
		prefix = "> "
	}

	if m.ui.Picker.isAdding && isSelected {
		split := splitAtCursor(m.ui.Picker.input.Value(), m.ui.Picker.input.Position())
		return fmt.Sprintf("%s+ %s%s%s",
			prefix,
			split.left,
			ui.GetCursorStyle(m.ui.WindowFocused).Render(split.cursorCh),
			split.right,
		)
	}

	line := prefix + "+ New Project"
	if isSelected {
		return ui.SelectedStyle.Render(line)
	}
	return ui.MutedStyle.Render(line)
}

func (m model) infoView() string {
	pos, ok := m.selectedPosition()
	if !ok {
		return ""
	}

	var lines []string
	switch pos.Kind {
	case focusProject:
		lines = m.projectInfoLines()
	case focusCategory:
		lines = m.categoryInfoLines(pos.CategoryIndex)
	case focusTask:
		lines = m.taskInfoLines(pos.CategoryIndex, pos.TaskIndex)
	}

	lines = append(lines, "", ui.DialogHintStyle.Render("i/esc/q close"))
	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}

func (m model) taskInfoLines(catIdx, taskIdx int) []string {
	task := m.project.Categories[catIdx].Tasks[taskIdx]
	category := m.project.Categories[catIdx]

	statusDisplay := formatStatusLabel(task.Status)
	priorityDisplay := formatPriorityLabel(task.Priority)

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

	estimateDisplay := FormatEstimateLabel(task.EstimateMinutes)

	lines = append(lines,
		fmt.Sprintf("Status:   %s", statusDisplay),
		fmt.Sprintf("Priority: %s", priorityDisplay),
		fmt.Sprintf("Estimate: %s", estimateDisplay),
		fmt.Sprintf("Category: %s", category.Name),
		"",
		fmt.Sprintf("Created:  %s", FormatDateWithRelative(task.CreatedAt)),
		fmt.Sprintf("Updated:  %s", FormatDateWithRelative(task.UpdatedAt)),
	)

	if task.CompletionDate != "" {
		lines = append(lines, fmt.Sprintf("Completed: %s", FormatDateWithRelative(task.CompletionDate)))
	}

	return lines
}

func (m model) categoryInfoLines(catIdx int) []string {
	category := m.project.Categories[catIdx]

	todoCount := 0
	inProgressCount := 0
	completedCount := 0
	cancelledCount := 0

	for _, task := range category.Tasks {
		switch task.Status {
		case domain.StatusTodo:
			todoCount++
		case domain.StatusInProgress:
			inProgressCount++
		case domain.StatusCompleted:
			completedCount++
		case domain.StatusCancelled:
			cancelledCount++
		}
	}

	estimateDisplay := FormatEstimateLabel(category.EstimateMinutes)

	lines := []string{
		ui.DialogTitleStyle.Render("Category Info"),
		"",
		fmt.Sprintf("Name:     %s", category.Name),
		fmt.Sprintf("Estimate: %s", estimateDisplay),
		fmt.Sprintf("Created:  %s", FormatDateWithRelative(category.CreatedAt)),
	}

	if category.UpdatedAt != "" {
		lines = append(lines, fmt.Sprintf("Updated:  %s", FormatDateWithRelative(category.UpdatedAt)))
	}

	lines = append(lines,
		"",
		fmt.Sprintf("Total Tasks: %d", len(category.Tasks)),
		"",
		"Task Breakdown:",
		fmt.Sprintf("  Todo:        %d", todoCount),
		fmt.Sprintf("  In Progress: %d", inProgressCount),
		fmt.Sprintf("  Completed:   %d", completedCount),
		fmt.Sprintf("  Cancelled:   %d", cancelledCount),
	)

	return lines
}

func (m model) projectInfoLines() []string {
	totalTasks := 0
	todoCount := 0
	inProgressCount := 0
	completedCount := 0
	cancelledCount := 0

	for _, cat := range m.project.Categories {
		totalTasks += len(cat.Tasks)
		for _, task := range cat.Tasks {
			switch task.Status {
			case domain.StatusTodo:
				todoCount++
			case domain.StatusInProgress:
				inProgressCount++
			case domain.StatusCompleted:
				completedCount++
			case domain.StatusCancelled:
				cancelledCount++
			}
		}
	}

	lines := []string{
		ui.DialogTitleStyle.Render("Project Info"),
		"",
		fmt.Sprintf("Name:       %s", m.project.Name),
		fmt.Sprintf("Created:    %s", FormatDateWithRelative(m.project.CreatedAt)),
		fmt.Sprintf("Updated:    %s", FormatDateWithRelative(m.project.UpdatedAt)),
		"",
		fmt.Sprintf("Categories: %d", len(m.project.Categories)),
		fmt.Sprintf("Total Tasks: %d", totalTasks),
		"",
		"Task Breakdown:",
		fmt.Sprintf("  Todo:        %d", todoCount),
		fmt.Sprintf("  In Progress: %d", inProgressCount),
		fmt.Sprintf("  Completed:   %d", completedCount),
		fmt.Sprintf("  Cancelled:   %d", cancelledCount),
	}

	return lines
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

func (m model) estimatePickerView() string {
	lines := []string{ui.DialogTitleStyle.Render("Time Estimate"), ""}

	presetLabels := []string{
		"None",
		"15 minutes",
		"30 minutes",
		"1 hour",
		"2 hours",
		"4 hours",
		"1 day",
		"2 days",
		"3 days",
		"5 days",
	}

	for i, label := range presetLabels {
		prefix := "  "
		if i == m.ui.EstimatePicker.Selected {
			prefix = "> "
		}
		line := prefix + label
		if i == m.ui.EstimatePicker.Selected {
			line = ui.SelectedStyle.Render(line)
		}
		lines = append(lines, line)
	}

	lines = append(lines, "", ui.DialogHintStyle.Render("j/k navigate | enter select | esc cancel"))
	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}
