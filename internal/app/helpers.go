package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const (
	prefixWidth     = 2
	footerHeight    = 3
	blankAfterProj  = 1
	blankBetweenCat = 1
	blankAfterCat   = 1
)

func safeWidth(totalWidth, overhead int) int {
	available := totalWidth - overhead
	if available < 1 {
		return 1
	}
	return available
}

type wrappedLines struct {
	lines  []string
	indent string
}

func wrapWithPrefix(text string, width, overhead int, prefix string) wrappedLines {
	if width <= 0 {
		return wrappedLines{lines: []string{prefix + text}, indent: strings.Repeat(" ", len(prefix))}
	}
	available := safeWidth(width, overhead)
	wrapped := ansi.Wrap(text, available, "")
	lines := strings.Split(wrapped, "\n")
	indent := strings.Repeat(" ", overhead)
	result := make([]string, len(lines))
	for i, line := range lines {
		if i == 0 {
			result[i] = prefix + line
		} else {
			result[i] = indent + line
		}
	}
	return wrappedLines{lines: result, indent: indent}
}

func countWrappedLines(text string, width, overhead int) int {
	if width <= 0 {
		return 1
	}
	available := safeWidth(width, overhead)
	wrapped := ansi.Wrap(text, available, "")
	return strings.Count(wrapped, "\n") + 1
}

type cursorSplit struct {
	left      string
	cursorCh  string
	right     string
	cursorPos int
}

func splitAtCursor(text string, cursor int) cursorSplit {
	if text == "" {
		text = " "
	}
	runes := []rune(text)
	pos := cursor
	if pos < 0 {
		pos = 0
	}
	if pos > len(runes) {
		pos = len(runes)
	}
	left := string(runes[:pos])
	right := string(runes[pos:])
	cursorCh := " "
	if pos < len(runes) {
		cursorCh = string(runes[pos])
		right = string(runes[pos+1:])
	}
	return cursorSplit{left: left, cursorCh: cursorCh, right: right, cursorPos: pos}
}

func renderCursorLine(text string, cursor int, width, overhead int, prefix string, textStyle, cursorStyle lipgloss.Style) string {
	if width <= 0 {
		split := splitAtCursor(text, cursor)
		return prefix + textStyle.Render(split.left) + cursorStyle.Render(split.cursorCh) + textStyle.Render(split.right)
	}
	if text == "" {
		text = " "
	}
	available := safeWidth(width, overhead)
	wrapped := ansi.Wrap(text, available, "")
	wrapLines := strings.Split(wrapped, "\n")
	indent := strings.Repeat(" ", overhead)
	pos := 0
	var result []string
	for i, line := range wrapLines {
		lineRunes := []rune(line)
		lineLen := len(lineRunes)
		var styledLine string
		if cursor >= pos && cursor < pos+lineLen {
			offset := cursor - pos
			l := string(lineRunes[:offset])
			c := string(lineRunes[offset])
			r := string(lineRunes[offset+1:])
			styledLine = textStyle.Render(l) + cursorStyle.Render(c) + textStyle.Render(r)
		} else if cursor == pos+lineLen {
			styledLine = textStyle.Render(line) + cursorStyle.Render(" ")
		} else {
			styledLine = textStyle.Render(line)
		}
		if i == 0 {
			result = append(result, prefix+styledLine)
		} else {
			result = append(result, indent+styledLine)
		}
		pos += lineLen + 1
	}
	return strings.Join(result, "\n")
}

func wrapAndStyleLines(text string, width, overhead int, prefix string, style lipgloss.Style) string {
	if width <= 0 {
		return style.Render(prefix + text)
	}
	available := safeWidth(width, overhead)
	wrapped := ansi.Wrap(text, available, "")
	lines := strings.Split(wrapped, "\n")
	indent := strings.Repeat(" ", overhead)
	var result []string
	for i, line := range lines {
		styledLine := style.Render(line)
		if i == 0 {
			result = append(result, prefix+styledLine)
		} else {
			result = append(result, indent+styledLine)
		}
	}
	return strings.Join(result, "\n")
}
