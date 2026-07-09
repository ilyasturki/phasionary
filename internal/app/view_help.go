package app

import (
	"fmt"
	"strings"

	"phasionary/internal/ui"
)

type helpRow struct {
	text         string
	focusable    bool
	disabled     bool
	bindingIndex int
	// header marks a section title row (e.g. "Actions:"). It groups the content
	// rows beneath it so the `/` filter can drop empty sections.
	header bool
	// filter is the plain "key label" text a row matches against under the `/`
	// filter, in natural case (smartcase applies). Empty for headers and spacers,
	// which are never matched directly.
	filter string
}

var helpHint = []ui.Hint{
	{Key: "ctrl+d/u", Label: "page"},
	{Key: "/", Label: "filter"},
	{Key: "enter", Label: "run"},
	{Key: "?/esc", Label: "close"},
}

var helpFilterHint = []ui.Hint{
	{Key: "enter", Label: "run"},
	{Key: "esc", Label: "clear"},
}

func helpDisabled(b keyBinding) bool {
	for _, k := range b.keys {
		if k == "?" {
			return true
		}
	}
	return false
}

func helpHeaderRow(title string) helpRow {
	return helpRow{text: ui.DialogTitleStyle.Render(title + ":"), header: true}
}

// helpTextRow builds a non-runnable reference row (Editing / Visual mode). It is
// searchable by the `/` filter but never focusable, so Enter can't run it.
func helpTextRow(key, desc string) helpRow {
	return helpRow{
		text:   fmt.Sprintf("  %-14s%s", key, desc),
		filter: key + " " + desc,
	}
}

var helpRowsAll, helpFocusablesAll = computeHelpRows()

func computeHelpRows() ([]helpRow, []int) {
	var rows []helpRow
	var focusables []int

	for _, section := range []string{sectionNavigation, sectionActions} {
		rows = append(rows, helpHeaderRow(section))
		for i, b := range normalBindings {
			if b.section != section || b.desc == "" {
				continue
			}
			display := b.display
			if display == "" {
				display = strings.Join(b.keys, "/")
			}
			line := fmt.Sprintf("  %-14s%s", display, b.desc)
			row := helpRow{
				text:         line,
				focusable:    true,
				bindingIndex: i,
				disabled:     helpDisabled(b),
				filter:       display + " " + b.desc,
			}
			if row.disabled {
				row.text = ui.MutedStyle.Render(line)
			}
			focusables = append(focusables, len(rows))
			rows = append(rows, row)
		}
		rows = append(rows, helpRow{})
	}

	rows = append(rows,
		helpHeaderRow("Editing"),
		helpTextRow("enter", "save changes"),
		helpTextRow("esc", "cancel editing"),
		helpTextRow("←/→", "move cursor"),
		helpTextRow("ctrl+a/e", "start/end of line"),
		helpTextRow("ctrl+w", "delete word backward"),
		helpTextRow("ctrl+k/u", "delete to end/start"),
		helpTextRow("ctrl+←/→", "word navigation"),
		helpRow{},
		helpHeaderRow("Visual mode"),
		helpTextRow("v", "enter visual mode (anchor at cursor)"),
		helpTextRow("j/k", "extend range (skips category rows)"),
		helpTextRow("J/K", "shift range down/up"),
		helpTextRow("o", "swap anchor and cursor"),
		helpTextRow("y", "copy titles (newline-joined)"),
		helpTextRow("Y", "copy as markdown checklist"),
		helpTextRow("x", "cut range (then p to paste)"),
		helpTextRow("d", "delete range (with confirmation)"),
		helpTextRow("esc", "exit visual mode"),
	)
	return rows, focusables
}

// filteredHelpRows narrows the shortcut list to rows whose key or label matches
// query, keeping a section header only when the section has a match. An empty
// query returns the full list. The returned focusables index into the returned
// rows.
func filteredHelpRows(query string) ([]helpRow, []int) {
	if strings.TrimSpace(query) == "" {
		return helpRowsAll, helpFocusablesAll
	}
	var rows []helpRow
	var focusables []int
	for i := 0; i < len(helpRowsAll); {
		if !helpRowsAll[i].header {
			i++
			continue
		}
		header := helpRowsAll[i]
		j := i + 1
		var matched []helpRow
		for j < len(helpRowsAll) && !helpRowsAll[j].header {
			r := helpRowsAll[j]
			if r.filter != "" && ui.Contains(r.filter, query) {
				matched = append(matched, r)
			}
			j++
		}
		if len(matched) > 0 {
			rows = append(rows, header)
			for _, r := range matched {
				if r.focusable {
					focusables = append(focusables, len(rows))
				}
				rows = append(rows, r)
			}
			rows = append(rows, helpRow{})
		}
		i = j
	}
	return rows, focusables
}

// currentHelpRows returns the rows and focusables in effect: the active filter's
// narrowed set while filtering, otherwise the full list.
func (m model) currentHelpRows() ([]helpRow, []int) {
	if m.ui.Help.Filtering {
		return filteredHelpRows(m.ui.Help.Filter.Value())
	}
	return helpRowsAll, helpFocusablesAll
}

func (m model) helpViewportHeight() int {
	chrome := 8
	if m.ui.Help.Filtering {
		chrome += 2 // the filter prompt line and its blank separator
	}
	h := m.ui.Screen.Height - chrome
	if h < 5 {
		return 5
	}
	return h
}

func (m *model) ensureHelpVisible() {
	rows, focusables := m.currentHelpRows()
	height := m.helpViewportHeight()
	maxOffset := len(rows) - height
	if maxOffset < 0 {
		maxOffset = 0
	}
	clamp := func() {
		if m.ui.Help.ScrollOffset > maxOffset {
			m.ui.Help.ScrollOffset = maxOffset
		}
		if m.ui.Help.ScrollOffset < 0 {
			m.ui.Help.ScrollOffset = 0
		}
	}
	if len(focusables) == 0 {
		clamp()
		return
	}
	if m.ui.Help.Focused < 0 {
		m.ui.Help.Focused = 0
	}
	if m.ui.Help.Focused >= len(focusables) {
		m.ui.Help.Focused = len(focusables) - 1
	}
	rowIdx := focusables[m.ui.Help.Focused]
	if rowIdx < m.ui.Help.ScrollOffset {
		m.ui.Help.ScrollOffset = rowIdx
	}
	if rowIdx >= m.ui.Help.ScrollOffset+height {
		m.ui.Help.ScrollOffset = rowIdx - height + 1
	}
	clamp()
}

func (m *model) moveHelpFocus(delta int) {
	_, focusables := m.currentHelpRows()
	if len(focusables) == 0 {
		return
	}
	target := m.ui.Help.Focused + delta
	if target < 0 {
		target = 0
	}
	if target >= len(focusables) {
		target = len(focusables) - 1
	}
	m.ui.Help.Focused = target
	m.ensureHelpVisible()
}

// helpFilterPrompt renders the "/query" input line shown at the top of the
// dialog while filtering, with a block cursor at the edit position.
func (m model) helpFilterPrompt() string {
	value := m.ui.Help.Filter.Value()
	cursorStyle := ui.GetCursorStyle(m.ui.Screen.WindowFocused)
	split := splitAtCursor(value, m.ui.Help.Filter.Position())
	return "/" + split.left + cursorStyle.Render(split.cursorCh) + split.right
}

func (m model) helpView() string {
	rows, focusables := m.currentHelpRows()

	focusedRowIdx := -1
	if len(focusables) > 0 {
		idx := m.ui.Help.Focused
		if idx < 0 {
			idx = 0
		}
		if idx >= len(focusables) {
			idx = len(focusables) - 1
		}
		focusedRowIdx = focusables[idx]
	}

	height := m.helpViewportHeight()
	start := m.ui.Help.ScrollOffset
	if start < 0 {
		start = 0
	}
	if start > len(rows) {
		start = len(rows)
	}
	end := start + height
	if end > len(rows) {
		end = len(rows)
	}

	var lines []string
	if m.ui.Help.Filtering {
		lines = append(lines, m.helpFilterPrompt(), "")
	}
	if start > 0 {
		lines = append(lines, ui.MutedStyle.Render(scrollMoreAbove))
	}
	if len(rows) == 0 {
		lines = append(lines, ui.MutedStyle.Render("  no matching shortcuts"))
	}
	for i := start; i < end; i++ {
		text := rows[i].text
		if i == focusedRowIdx {
			text = ui.SelectedStyle.Render(text)
		}
		lines = append(lines, text)
	}
	if end < len(rows) {
		lines = append(lines, ui.MutedStyle.Render(scrollMoreBelow))
	}

	hint := helpHint
	if m.ui.Help.Filtering {
		hint = helpFilterHint
	}
	lines = append(lines, "", ui.RenderHints(hint))
	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}
