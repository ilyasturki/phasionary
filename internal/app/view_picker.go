package app

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"phasionary/internal/domain"
	"phasionary/internal/ui"
)

const (
	pickerWidthFraction   = 0.5
	pickerContentMinWidth = 50
	pickerContentMaxWidth = 70
)

var pickerDialogChromeWidth = ui.HelpDialogStyle.GetHorizontalPadding() + ui.HelpDialogStyle.GetHorizontalBorderSize()

const (
	// pickerChromeRows is the picker dialog's fixed vertical overhead, with no
	// project rows, scroll indicators, or hint footer: rounded border (2) +
	// vertical padding (2) + title (1) + blank under title (1) + pinned
	// New Project row (1) + its separator (1) + blank above the hints (1).
	pickerChromeRows = 9
	// pickerScrollReserve is the rows kept free below the project list for the
	// up/down scroll indicators, so the dialog never overflows (and gets
	// clipped) while scrolling. See pickerVisibleCount.
	pickerScrollReserve = 2
	// pickerMinVisible keeps the list usable on short terminals even when the
	// height budget would otherwise round down to nothing.
	pickerMinVisible = 3
	// pickerFallbackVisible is used before the first window-size message, when
	// the terminal height isn't known yet.
	pickerFallbackVisible = 10
)

// pickerNormalHints are the footer hints shown while browsing the picker. Plain
// navigation (j/k, paging, g/G) is intentionally omitted per dialog-footer
// convention; only picker-specific actions are listed.
func pickerNormalHints() []ui.Hint {
	return []ui.Hint{
		{Key: "⏎", Label: "select"},
		{Key: "J/K", Label: "reorder"},
		{Key: "d", Label: "delete"},
		{Key: "esc", Label: "cancel"},
	}
}

// pickerHintRows reports how many rows the footer hints occupy once wrapped to
// the dialog's content width (1 on wide terminals, more when they wrap).
func (m model) pickerHintRows() int {
	rendered := lipgloss.NewStyle().Width(m.pickerContentWidth()).Render(ui.RenderHints(pickerNormalHints()))
	return lipgloss.Height(rendered)
}

// pickerVisibleCount returns how many project slots the list shows at once,
// sized to the terminal so tall terminals scroll less and short ones don't
// clip the footer. It accounts for the dialog chrome, the (possibly wrapped)
// hint footer, and worst-case scroll indicators, and never exceeds the number
// of items.
func (m model) pickerVisibleCount() int {
	total := len(m.ui.Picker.projects)
	if m.ui.Screen.Height <= 0 {
		return min(total, pickerFallbackVisible)
	}
	budget := m.ui.Screen.Height - pickerChromeRows - m.pickerHintRows() - pickerScrollReserve
	if budget < pickerMinVisible {
		budget = pickerMinVisible
	}
	return min(budget, total)
}

func (m model) pickerContentWidth() int {
	total := m.ui.Screen.Width
	if total <= 0 {
		return pickerContentMinWidth
	}
	target := int(float64(total) * pickerWidthFraction)
	content := target - pickerDialogChromeWidth
	if content < pickerContentMinWidth {
		content = pickerContentMinWidth
	}
	if content > pickerContentMaxWidth {
		content = pickerContentMaxWidth
	}
	if maxContent := total - pickerDialogChromeWidth - 2; content > maxContent && maxContent > 10 {
		content = maxContent
	}
	if content < 10 {
		content = 10
	}
	return content
}

func padToWidth(line string, width int) string {
	w := lipgloss.Width(line)
	if w >= width {
		return line
	}
	return line + strings.Repeat(" ", width-w)
}

func (m model) projectPickerView() string {
	contentWidth := m.pickerContentWidth()

	title := fmt.Sprintf("Select Project (%d)", len(m.ui.Picker.projects))
	lines := []string{
		ui.DialogTitleStyle.Render(title),
		"",
		// "+ New Project" is pinned above the scrolling list so it's always
		// reachable (Tab), never scrolled away.
		m.renderNewProjectLine(m.ui.Picker.onNew, contentWidth),
		ui.DialogHintStyle.Render(strings.Repeat("─", contentWidth)),
	}

	n := len(m.ui.Picker.projects)
	start := m.ui.Picker.scrollOffset
	end := start + m.pickerVisibleCount()
	if end > n {
		end = n
	}

	if start > 0 {
		lines = append(lines, ui.DialogHintStyle.Render(scrollMoreAbove))
	}
	for i := start; i < end; i++ {
		isSelected := !m.ui.Picker.onNew && i == m.ui.Picker.selected
		lines = append(lines, m.renderPickerRow(i, isSelected, contentWidth))
	}
	if end < n {
		lines = append(lines, ui.DialogHintStyle.Render(scrollMoreBelow))
	}

	hints := pickerNormalHints()
	if m.ui.Picker.isAdding {
		hints = []ui.Hint{
			{Key: "⏎", Label: "create"},
			{Key: "esc", Label: "cancel"},
		}
	}
	lines = append(lines, "", ui.RenderHints(hints))

	// lipgloss.Width is the total block width (content + padding + border),
	// so add both back to land on the intended content area.
	return ui.HelpDialogStyle.Width(contentWidth + pickerDialogChromeWidth).Render(strings.Join(lines, "\n"))
}

func (m model) renderPickerRow(i int, isSelected bool, contentWidth int) string {
	p := m.ui.Picker.projects[i]
	prefix := "  "
	if isSelected {
		prefix = "> "
	}
	suffix := ""
	if p.ID == m.project.ID {
		suffix = " (current)"
	}
	badge, complete := projectProgressBadge(p)

	name, gap := pickerRowLayout(prefix, p.Name, suffix, badge, contentWidth)

	if isSelected {
		// The whole row is one reverse-video band, so every segment (including
		// the badge) renders inside it.
		row := prefix + name + suffix + strings.Repeat(" ", gap) + badge
		return ui.SelectedStyle.Render(padToWidth(row, contentWidth))
	}

	row := prefix + name
	if suffix != "" {
		row += ui.DialogHintStyle.Render(suffix)
	}
	row += strings.Repeat(" ", gap)
	if badge != "" {
		row += styleProgressBadge(badge, complete)
	}
	return row
}

func (m model) renderNewProjectLine(isSelected bool, contentWidth int) string {
	prefix := "  "
	if isSelected {
		prefix = "> "
	}

	if m.ui.Picker.isAdding && isSelected {
		split := splitAtCursor(m.ui.Picker.input.Value(), m.ui.Picker.input.Position())
		return fmt.Sprintf("%s%s %s%s%s",
			prefix,
			ui.SuccessStyle.Render("+"),
			split.left,
			ui.GetCursorStyle(m.ui.Screen.WindowFocused).Render(split.cursorCh),
			split.right,
		)
	}

	if isSelected {
		return ui.SelectedStyle.Render(padToWidth(prefix+"+ New Project", contentWidth))
	}
	// The green "+" reads as an affordance to act, not a disabled row.
	return prefix + ui.SuccessStyle.Render("+") + " New Project"
}

// pickerRowLayout truncates the project name so the row fits contentWidth with
// its " (current)" suffix and right-aligned progress badge intact, returning
// the (possibly truncated) name and the gap that right-aligns the badge.
func pickerRowLayout(prefix, name, suffix, badge string, contentWidth int) (string, int) {
	badgeW := lipgloss.Width(badge)
	reserve := badgeW
	if badgeW > 0 {
		reserve++ // at least one space between the name/suffix and the badge
	}
	maxNameW := contentWidth - lipgloss.Width(prefix) - lipgloss.Width(suffix) - reserve
	if maxNameW < 1 {
		maxNameW = 1
	}
	if lipgloss.Width(name) > maxNameW {
		name = ansi.Truncate(name, maxNameW, "…")
	}
	used := lipgloss.Width(prefix) + lipgloss.Width(name) + lipgloss.Width(suffix) + badgeW
	gap := contentWidth - used
	if gap < 0 {
		gap = 0
	}
	return name, gap
}

// projectProgress counts completed tasks against the active (non-cancelled)
// task total, so a project reads as done when every task that still matters is
// completed.
func projectProgress(p domain.Project) (done, total int) {
	for _, c := range p.Categories {
		for _, t := range c.Tasks {
			if t.Status == domain.StatusCancelled {
				continue
			}
			total++
			if t.Status == domain.StatusCompleted {
				done++
			}
		}
	}
	return done, total
}

// projectProgressBadge renders a project's "done/total" badge, with a ✓ once
// everything's complete. Empty projects get no badge.
func projectProgressBadge(p domain.Project) (badge string, complete bool) {
	done, total := projectProgress(p)
	if total == 0 {
		return "", false
	}
	if done == total {
		return fmt.Sprintf("%d/%d ✓", done, total), true
	}
	return fmt.Sprintf("%d/%d", done, total), false
}

func styleProgressBadge(badge string, complete bool) string {
	if complete {
		return ui.SuccessStyle.Render(badge)
	}
	return ui.DialogHintStyle.Render(badge)
}
