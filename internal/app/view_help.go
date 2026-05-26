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
}

const helpHint = "ctrl+d/u page  ·  enter run  ·  ?/esc close"

func helpDisabled(b keyBinding) bool {
	for _, k := range b.keys {
		if k == "?" {
			return true
		}
	}
	return false
}

var helpRows, helpFocusables = computeHelpRows()

func computeHelpRows() ([]helpRow, []int) {
	var rows []helpRow
	var focusables []int

	for _, section := range []string{sectionNavigation, sectionActions} {
		rows = append(rows, helpRow{text: ui.DialogTitleStyle.Render(section + ":")})
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
		helpRow{text: ui.DialogTitleStyle.Render("Editing:")},
		helpRow{text: "  enter         save changes"},
		helpRow{text: "  esc           cancel editing"},
		helpRow{text: "  ←/→           move cursor"},
		helpRow{text: "  ctrl+a/e      start/end of line"},
		helpRow{text: "  ctrl+w        delete word backward"},
		helpRow{text: "  ctrl+k/u      delete to end/start"},
		helpRow{text: "  ctrl+←/→      word navigation"},
		helpRow{},
		helpRow{text: ui.DialogTitleStyle.Render("Visual mode:")},
		helpRow{text: "  v             enter visual mode (anchor at cursor)"},
		helpRow{text: "  j/k           extend range (skips category rows)"},
		helpRow{text: "  o             swap anchor and cursor"},
		helpRow{text: "  y             copy titles (newline-joined)"},
		helpRow{text: "  Y             copy as markdown checklist"},
		helpRow{text: "  x             cut range (then p to paste)"},
		helpRow{text: "  d             delete range (with confirmation)"},
		helpRow{text: "  esc           exit visual mode"},
	)
	return rows, focusables
}

func (m model) helpViewportHeight() int {
	const chrome = 8
	h := m.ui.Screen.Height - chrome
	if h < 5 {
		return 5
	}
	return h
}

func (m *model) ensureHelpVisible() {
	if len(helpFocusables) == 0 {
		return
	}
	if m.ui.Help.Focused < 0 {
		m.ui.Help.Focused = 0
	}
	if m.ui.Help.Focused >= len(helpFocusables) {
		m.ui.Help.Focused = len(helpFocusables) - 1
	}
	rowIdx := helpFocusables[m.ui.Help.Focused]
	height := m.helpViewportHeight()

	if rowIdx < m.ui.Help.ScrollOffset {
		m.ui.Help.ScrollOffset = rowIdx
	}
	if rowIdx >= m.ui.Help.ScrollOffset+height {
		m.ui.Help.ScrollOffset = rowIdx - height + 1
	}
	maxOffset := len(helpRows) - height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.ui.Help.ScrollOffset > maxOffset {
		m.ui.Help.ScrollOffset = maxOffset
	}
	if m.ui.Help.ScrollOffset < 0 {
		m.ui.Help.ScrollOffset = 0
	}
}

func (m *model) moveHelpFocus(delta int) {
	if len(helpFocusables) == 0 {
		return
	}
	target := m.ui.Help.Focused + delta
	if target < 0 {
		target = 0
	}
	if target >= len(helpFocusables) {
		target = len(helpFocusables) - 1
	}
	m.ui.Help.Focused = target
	m.ensureHelpVisible()
}

func (m model) helpView() string {
	focusedRowIdx := -1
	if len(helpFocusables) > 0 {
		idx := m.ui.Help.Focused
		if idx < 0 {
			idx = 0
		}
		if idx >= len(helpFocusables) {
			idx = len(helpFocusables) - 1
		}
		focusedRowIdx = helpFocusables[idx]
	}

	height := m.helpViewportHeight()
	start := m.ui.Help.ScrollOffset
	if start < 0 {
		start = 0
	}
	if start > len(helpRows) {
		start = len(helpRows)
	}
	end := start + height
	if end > len(helpRows) {
		end = len(helpRows)
	}

	var lines []string
	if start > 0 {
		lines = append(lines, ui.MutedStyle.Render(scrollMoreAbove))
	}
	for i := start; i < end; i++ {
		text := helpRows[i].text
		if i == focusedRowIdx {
			text = ui.SelectedStyle.Render(text)
		}
		lines = append(lines, text)
	}
	if end < len(helpRows) {
		lines = append(lines, ui.MutedStyle.Render(scrollMoreBelow))
	}

	lines = append(lines, "", ui.DialogHintStyle.Render(helpHint))
	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}
