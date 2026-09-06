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
	// The picker's own rows around the project list: title, the blank under it,
	// the pinned New Project row, the blank under that, the blank above the
	// hints.
	pickerOwnRows = 5
	// Full-screen adds a leading blank row, standing in for the panel's top
	// padding.
	pickerFullOwnRows = pickerOwnRows + 1
	// pickerScrollReserve is the rows kept free below the project list for the
	// up/down scroll indicators, so the picker never overflows (and gets
	// clipped) while scrolling. See pickerVisibleCount.
	pickerScrollReserve = 2
	// pickerMinVisible keeps the list usable on short terminals even when the
	// height budget would otherwise round down to nothing.
	pickerMinVisible = 3
	// pickerFallbackVisible is used before the first window-size message, when
	// the terminal height isn't known yet.
	pickerFallbackVisible = 10
	// Left/right gutter of the full-screen frame, standing in for the panel's
	// horizontal padding.
	pickerFullMargin = 2
	// Floor on the full-screen content width, mirroring the dialog's own floor:
	// on a very narrow terminal, overflow beats collapsing to nothing.
	pickerFullMinWidth = 20
	// Room the metadata block needs at its widest ("9999 open", "99 ▸", "12mo",
	// plus gaps). It caps the full-screen frame rather than sizing anything, so
	// an implausibly large count costs alignment, not correctness.
	pickerFullMetaWidth = 24
	// Columns a name keeps before a metadata column is dropped to make room.
	pickerMinNameWidth = 16
	// Blank columns between two metadata columns.
	pickerColumnGap = 2
)

// pickerNormalHints are the footer hints shown while browsing the picker. Plain
// navigation (j/k, paging, g/G) is intentionally omitted per dialog-footer
// convention; only picker-specific actions are listed. The labels are clipped
// so the six of them fit one row at the maximum dialog width.
func (m model) pickerNormalHints() []ui.Hint {
	// With no project behind the picker there is nothing to go back to, and esc
	// exits the app instead.
	back := "back"
	if m.project.ID == "" {
		back = "quit"
	}
	return []ui.Hint{
		{Key: "⏎", Label: "select"},
		{Key: "a", Label: "new"},
		{Key: "/", Label: "search"},
		{Key: "J/K", Label: "move"},
		{Key: "d", Label: "delete"},
		{Key: "esc", Label: back},
	}
}

func (m model) pickerHints() []ui.Hint {
	switch {
	case m.ui.Picker.isAdding:
		return []ui.Hint{{Key: "⏎", Label: "create"}, {Key: "esc", Label: "cancel"}}
	case m.ui.Picker.filtering:
		return []ui.Hint{{Key: "⏎", Label: "select"}, {Key: "esc", Label: "clear"}}
	}
	return m.pickerNormalHints()
}

// pickerContentWidth is the width of the picker's content area: the width every
// dialog shares when it floats over the project, the terminal minus its gutters
// when it owns the screen.
func (m model) pickerContentWidth() int {
	if !m.ui.Picker.fullscreen {
		return m.dialogWidth()
	}
	// Wider than a dialog, but never the whole terminal: names keep the columns a
	// dialog gives them and the metadata stays beside them instead of stranded a
	// screen's width away.
	widest := ui.DialogContentWidth(0) + pickerFullMetaWidth
	if m.ui.Screen.Width <= 0 {
		return widest
	}
	return max(min(m.ui.Screen.Width-2*pickerFullMargin, widest), pickerFullMinWidth)
}

// pickerHintRows reports how many rows the footer hints occupy once wrapped to
// the picker's content width (1 on wide terminals, more when they wrap).
func (m model) pickerHintRows() int {
	rendered := lipgloss.NewStyle().Width(m.pickerContentWidth()).Render(ui.RenderHints(m.pickerNormalHints()))
	return lipgloss.Height(rendered)
}

// pickerVisibleCount returns how many project slots the list shows at once,
// sized to the terminal so tall terminals scroll less and short ones don't clip
// the footer. It accounts for the picker's own rows, the (possibly wrapped)
// hint footer, and worst-case scroll indicators, and never exceeds the number
// of items.
func (m model) pickerVisibleCount() int {
	total := len(m.ui.Picker.projects)
	if m.ui.Screen.Height <= 0 {
		return min(total, pickerFallbackVisible)
	}
	if m.ui.Picker.fullscreen {
		// No frame to pay for: the budget is the terminal itself.
		ownRows := pickerFullOwnRows + m.pickerHintRows() + pickerScrollReserve
		return min(max(m.ui.Screen.Height-ownRows, pickerMinVisible), total)
	}
	ownRows := pickerOwnRows + m.pickerHintRows() + pickerScrollReserve
	return min(ui.PanelBodyHeight(m.ui.Screen.Height, ownRows, pickerMinVisible), total)
}

func (m model) projectPickerView() string {
	if m.ui.Picker.fullscreen {
		return m.pickerFullscreenView()
	}
	contentWidth := m.pickerContentWidth()
	lines := append(m.pickerBody(contentWidth), "", ui.RenderHints(m.pickerHints()))
	// Width is what makes the frameless panel opaque: without it lipgloss leaves
	// short rows unpadded and the project list shows through them.
	return ui.PanelStyle.Width(contentWidth + ui.PanelChromeWidth).Render(strings.Join(lines, "\n"))
}

// pickerFullscreenView draws the picker as the whole terminal: no frame, no
// project behind it, and the hints resting on the bottom row.
func (m model) pickerFullscreenView() string {
	contentWidth := m.pickerContentWidth()
	gutter := strings.Repeat(" ", pickerFullMargin)

	lines := []string{""}
	for _, line := range m.pickerBody(contentWidth) {
		lines = append(lines, gutter+line)
	}

	// Nothing of the main view renders under a full-screen picker, so the status
	// line it normally shares would be lost. It takes the blank row above the
	// footer, which the height budget already reserves.
	status := ""
	if m.ui.Screen.StatusMsg != "" {
		status = gutter + ui.MutedStyle.Render(ansi.Truncate(m.ui.Screen.StatusMsg, contentWidth, "…"))
	}

	// Wrap the hints here: with no frame around them nothing else would, and a
	// footer that soft-wraps in the terminal occupies rows the padding below
	// never accounted for.
	footer := strings.Split(lipgloss.NewStyle().Width(contentWidth).Render(ui.RenderHints(m.pickerHints())), "\n")
	for len(lines)+1+len(footer) < m.ui.Screen.Height {
		lines = append(lines, "")
	}
	lines = append(lines, status)
	for _, line := range footer {
		lines = append(lines, gutter+line)
	}
	return strings.Join(lines, "\n")
}

// pickerBody is everything above the footer: the title, the pinned New Project
// (or filter) row, and the scrolled project list with its overflow markers.
func (m model) pickerBody(contentWidth int) []string {
	p := m.ui.Picker
	title := fmt.Sprintf("Projects (%d)", len(p.projects))
	// While filtering, the query prompt takes the pinned top row's place: New
	// Project isn't a filter target, so it's hidden until the filter is cleared.
	topLine := m.renderNewProjectLine(contentWidth)
	if p.filtering {
		topLine = m.renderPickerFilterLine(contentWidth)
	}
	lines := []string{ui.DialogTitleStyle.Render(title), "", topLine, ""}

	n := len(p.projects)
	start := p.scrollOffset
	end := min(start+m.pickerVisibleCount(), n)

	if p.filtering && n == 0 {
		lines = append(lines, ui.MutedStyle.Render("  no matches"))
	}
	if start > 0 {
		lines = append(lines, ui.DialogHintStyle.Render(scrollMoreAbove))
	}
	// Column widths come from the whole list, not the visible slice, so they
	// don't shuffle as it scrolls — and from the unfiltered set while filtering,
	// so they don't shuffle as the query is typed.
	cols := measurePickerColumns(p.columnSource(), contentWidth)
	for i := start; i < end; i++ {
		isSelected := !p.onNew && !p.isAdding && i == p.selected
		lines = append(lines, m.renderPickerRow(i, isSelected, cols, contentWidth))
	}
	if end < n {
		lines = append(lines, ui.DialogHintStyle.Render(scrollMoreBelow))
	}
	return lines
}

// pickerColumns holds the width of each metadata column, 0 for a column that
// isn't drawn — either because no project has anything to put in it, or because
// the frame was too narrow to keep it.
type pickerColumns struct {
	open       int
	inProgress int
	edited     int
}

func (c pickerColumns) width() int {
	total := 0
	for _, w := range []int{c.open, c.inProgress, c.edited} {
		if w > 0 {
			total += pickerColumnGap + w
		}
	}
	return total
}

// measurePickerColumns sizes each metadata column to its widest cell, then
// drops columns right to left until the name column is worth reading. The order
// is the reverse of their usefulness: the age goes first, the open count last.
func measurePickerColumns(projects []domain.Project, contentWidth int) pickerColumns {
	var cols pickerColumns
	for _, p := range projects {
		open, inProgress := projectStats(p)
		cols.open = max(cols.open, ansi.StringWidth(openCell(open)))
		if inProgress > 0 {
			cols.inProgress = max(cols.inProgress, ansi.StringWidth(inProgressCell(inProgress)))
		}
		cols.edited = max(cols.edited, ansi.StringWidth(FormatRelativeShort(p.UpdatedAt)))
	}
	for _, drop := range []*int{&cols.edited, &cols.inProgress, &cols.open} {
		if pickerRowPrefixWidth+pickerMinNameWidth+cols.width() <= contentWidth {
			break
		}
		*drop = 0
	}
	return cols
}

// pickerRowPrefixWidth is the "  " / "> " cursor gutter every row carries.
const pickerRowPrefixWidth = 2

// projectStats counts a project's open (todo or in-progress) and in-progress
// tasks. Separators are dividers rather than work, and a cancelled task is work
// that no longer has to happen, so neither counts.
func projectStats(p domain.Project) (open, inProgress int) {
	for _, c := range p.Categories {
		for _, t := range c.Tasks {
			if t.IsSeparator() {
				continue
			}
			switch t.Status {
			case domain.StatusCompleted, domain.StatusCancelled:
			case domain.StatusInProgress:
				open++
				inProgress++
			default:
				open++
			}
		}
	}
	return open, inProgress
}

func openCell(open int) string {
	return fmt.Sprintf("%d open", open)
}

func inProgressCell(inProgress int) string {
	return fmt.Sprintf("%d ▸", inProgress)
}

// pickerMetaCells renders a project's metadata columns, each right-aligned in
// its own width. base carries the selected row's reverse band, so on that row
// every cell renders inside it rather than as a color block pasted onto it.
func pickerMetaCells(p domain.Project, cols pickerColumns, base lipgloss.Style, focused bool) string {
	open, inProgress := projectStats(p)

	type cell struct {
		text  string
		width int
		style lipgloss.Style
	}
	cells := []cell{
		{openCell(open), cols.open, ui.DialogHintStyle},
		{"", cols.inProgress, ui.StatusStyle(domain.StatusInProgress)},
		{FormatRelativeShort(p.UpdatedAt), cols.edited, ui.MutedStyle},
	}
	if inProgress > 0 {
		cells[1].text = inProgressCell(inProgress)
	}

	var out strings.Builder
	for _, c := range cells {
		if c.width == 0 {
			continue
		}
		style := base.Inherit(c.style)
		if focused {
			style = base
		}
		pad := c.width - ansi.StringWidth(c.text)
		out.WriteString(base.Render(strings.Repeat(" ", pickerColumnGap+max(pad, 0))))
		out.WriteString(style.Render(c.text))
	}
	return out.String()
}

func (m model) renderPickerRow(i int, isSelected bool, cols pickerColumns, contentWidth int) string {
	p := m.ui.Picker.projects[i]

	base := lipgloss.NewStyle()
	if isSelected {
		base = ui.GetSelectedStyle(m.ui.Screen.WindowFocused)
	}
	prefix := "  "
	if isSelected {
		prefix = "> "
	}
	suffix := ""
	if p.ID == m.project.ID {
		suffix = " (current)"
	}

	meta := pickerMetaCells(p, cols, base, isSelected)
	name, gap := pickerRowLayout(prefix, p.Name, suffix, lipgloss.Width(meta), contentWidth)

	displayName := base.Render(name)
	if !isSelected && m.ui.Picker.filtering && m.ui.Picker.query != "" {
		displayName = ui.HighlightMatches(name, m.ui.Picker.query, lipgloss.NewStyle(), ui.SearchMatchStyle)
	}
	suffixStyle := base.Inherit(ui.DialogHintStyle)
	if isSelected {
		suffixStyle = base
	}
	return base.Render(prefix) + displayName + suffixStyle.Render(suffix) +
		base.Render(strings.Repeat(" ", gap)) + meta
}

// renderNewProjectLine draws the pinned "+ New Project" row, or the inline name
// input once adding has started. Adding is keyed off isAdding alone, not off
// the cursor: `a` opens the input from wherever the cursor sits, and leaves it
// there so cancelling returns to it.
func (m model) renderNewProjectLine(contentWidth int) string {
	p := m.ui.Picker
	if p.isAdding {
		split := splitAtCursor(p.input.Value(), p.input.Position())
		return fmt.Sprintf("  %s %s%s%s",
			ui.SuccessStyle.Render("+"),
			split.left,
			ui.GetCursorStyle(m.ui.Screen.WindowFocused).Render(split.cursorCh),
			split.right,
		)
	}
	if p.onNew {
		return ui.GetSelectedStyle(m.ui.Screen.WindowFocused).Render(ui.PadTo("> + New Project", contentWidth))
	}
	// The green "+" reads as an affordance to act, not a disabled row.
	return "  " + ui.SuccessStyle.Render("+") + " New Project"
}

// renderPickerFilterLine draws the type-to-filter prompt that replaces the New
// Project row while filtering.
func (m model) renderPickerFilterLine(contentWidth int) string {
	return m.filterPromptRow(m.ui.Picker.filter, m.ui.Picker.query, len(m.ui.Picker.projects), contentWidth)
}

// pickerRowLayout truncates the project name so the row fits contentWidth with
// its " (current)" suffix and metadata columns intact, returning the (possibly
// truncated) name and the gap that pushes the columns flush right.
func pickerRowLayout(prefix, name, suffix string, metaWidth, contentWidth int) (string, int) {
	maxNameW := contentWidth - lipgloss.Width(prefix) - lipgloss.Width(suffix) - metaWidth
	if maxNameW < 1 {
		maxNameW = 1
	}
	if lipgloss.Width(name) > maxNameW {
		name = ansi.Truncate(name, maxNameW, "…")
	}
	used := lipgloss.Width(prefix) + lipgloss.Width(name) + lipgloss.Width(suffix) + metaWidth
	return name, max(contentWidth-used, 0)
}
