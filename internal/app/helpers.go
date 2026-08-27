package app

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"phasionary/internal/app/selection"
	"phasionary/internal/ui"
)

func (m model) positions() []selection.Position {
	return m.ui.Selection.Positions()
}

func (m model) selected() int {
	return m.ui.Selection.Selected()
}

func (m model) selectedPosition() (selection.Position, bool) {
	return m.ui.Selection.SelectedPosition()
}

func (m model) isTaskCut(taskID string) bool {
	if !m.ui.Clipboard.IsCut || taskID == "" {
		return false
	}
	if m.ui.Clipboard.SourceID == taskID {
		return true
	}
	for _, id := range m.ui.Clipboard.TaskIDs {
		if id == taskID {
			return true
		}
	}
	return false
}

func (m model) isCategoryCut(categoryID string) bool {
	if !m.ui.Clipboard.IsCut || categoryID == "" {
		return false
	}
	for _, id := range m.ui.Clipboard.CategoryIDs {
		if id == categoryID {
			return true
		}
	}
	return false
}

const (
	prefixWidth     = 2
	footerHeight    = 0 // No bottom status area — status text appears on the project line.
	blankAfterProj  = 1
	blankBetweenCat = 1
	blankAfterCat   = 1
)

func (m model) footerHeight() int {
	return footerHeight + m.bottomBarHeight()
}

func (m model) layoutConfig() LayoutConfig {
	cfg := DefaultLayoutConfig()
	cfg.FooterHeight += m.bottomBarHeight()
	return cfg
}

func sanitizeInput(input *textinput.Model) {
	val := input.Value()
	cleaned := strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == '\t' {
			return ' '
		}
		return r
	}, val)
	if cleaned != val {
		pos := input.Position()
		input.SetValue(cleaned)
		input.SetCursor(pos)
	}
}

func safeWidth(totalWidth, overhead int) int {
	available := totalWidth - overhead
	if available < 1 {
		return 1
	}
	return available
}

// prefixLine renders already-styled text on a single row behind prefix,
// truncating rather than wrapping. Edit rows are sized by the layout as one row
// (see renderCursorLine), so their placeholders have to stay on one too.
func prefixLine(text string, width, overhead int, prefix string) string {
	if width <= 0 {
		return prefix + text
	}
	return prefix + ansi.Truncate(text, safeWidth(width, overhead), ui.Ellipsis)
}

// countWrappedLines sizes a single-line field's row under the same clamp the
// renderers apply, so the height reserved here always matches the height drawn.
func countWrappedLines(text string, width, overhead int) int {
	if width <= 0 {
		return 1
	}
	return ui.CountClamped(text, safeWidth(width, overhead), ui.MaxLineRows)
}

type cursorSplit struct {
	left     string
	cursorCh string
	right    string
}

func splitAtCursor(text string, cursor int) cursorSplit {
	if text == "" {
		text = " "
	}
	runes := []rune(text)
	pos := min(max(cursor, 0), len(runes))
	left := string(runes[:pos])
	right := string(runes[pos:])
	cursorCh := " "
	if pos < len(runes) {
		cursorCh = string(runes[pos])
		right = string(runes[pos+1:])
	}
	return cursorSplit{left: left, cursorCh: cursorCh, right: right}
}

// renderCursorLine draws an inline editor as exactly one row: a horizontal
// window onto the buffer, positioned to keep the cursor visible, with an
// ellipsis at either end marking text scrolled out of view.
//
// It stays one row on purpose. The layout sizes an edited row from the *stored*
// value, so a growing buffer that wrapped would push the rest of the screen —
// and the bottom bar — off the terminal as the user typed. A window can't: it is
// always exactly as tall as the row the layout reserved, whatever gets typed.
func renderCursorLine(text string, cursor int, width, overhead int, prefix string, textStyle, cursorStyle lipgloss.Style) string {
	if width <= 0 {
		split := splitAtCursor(text, cursor)
		return prefix + textStyle.Render(split.left) + cursorStyle.Render(split.cursorCh) + textStyle.Render(split.right)
	}
	runes := []rune(text)
	pos := min(max(cursor, 0), len(runes))

	// The cursor needs a cell of its own past the last rune when it sits at the
	// end of the buffer, so the window is measured over that virtual cell too.
	span := len(runes)
	if pos == span {
		span++
	}
	available := safeWidth(width, overhead)
	// An ellipsis costs a cell at whichever end it appears. Reserving both ends
	// whenever the buffer overflows leaves at most one column unused, which a
	// left-aligned row doesn't show, and spares the render a fixpoint.
	content := available
	if span > available {
		content = max(available-2, 1)
	}
	start, end := cursorWindow(span, pos, content)
	end = min(end, len(runes))

	line := prefix
	if start > 0 {
		line += textStyle.Render(ui.Ellipsis)
	}
	cursorCh, right := " ", ""
	if pos < end {
		cursorCh = string(runes[pos])
		right = string(runes[pos+1 : end])
	}
	line += textStyle.Render(string(runes[start:pos])) + cursorStyle.Render(cursorCh) + textStyle.Render(right)
	if end < len(runes) {
		line += textStyle.Render(ui.Ellipsis)
	}
	return line
}

// cursorWindow returns the [start, end) rune window of the given width that
// contains pos, sliding only as far as it must to keep the cursor in view.
func cursorWindow(span, pos, width int) (int, int) {
	if span <= width {
		return 0, span
	}
	start := min(max(pos-width+1, 0), span-width)
	return start, start + width
}
