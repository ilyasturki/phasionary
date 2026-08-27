package ui

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
)

// MatchIndices returns the byte ranges within text that match query, using
// smartcase semantics: the search is case-insensitive unless query contains an
// uppercase rune, in which case it is case-sensitive. Ranges are
// non-overlapping and left-to-right. Returns nil when query is empty or absent.
func MatchIndices(text, query string) [][2]int {
	needle := []rune(query)
	if len(needle) == 0 {
		return nil
	}
	caseSensitive := hasUpper(query)

	var out [][2]int
	for at := 0; at < len(text); {
		if end := matchAt(text, at, needle, caseSensitive); end >= 0 {
			out = append(out, [2]int{at, end})
			at = end
			continue
		}
		_, size := utf8.DecodeRuneInString(text[at:])
		at += size
	}
	return out
}

// FirstMatchIndex returns the byte offset where the first smartcase match of
// query begins in text, or -1 when there is none. It is used to aim a clamped
// display window at a match instead of at the start of the string (see
// WrapClamped), on fields that carry no length limit.
func FirstMatchIndex(text, query string) int {
	needle := []rune(query)
	if len(needle) == 0 {
		return -1
	}
	caseSensitive := hasUpper(query)
	for at := range text {
		if matchAt(text, at, needle, caseSensitive) >= 0 {
			return at
		}
	}
	return -1
}

// matchAt returns the byte offset just past a match of needle beginning at at,
// or -1 when text does not match there.
func matchAt(text string, at int, needle []rune, caseSensitive bool) int {
	for _, n := range needle {
		if at >= len(text) {
			return -1
		}
		r, size := utf8.DecodeRuneInString(text[at:])
		if r != n && (caseSensitive || unicode.ToLower(r) != unicode.ToLower(n)) {
			return -1
		}
		at += size
	}
	return at
}

// Contains reports whether text contains query under smartcase semantics.
func Contains(text, query string) bool {
	return FirstMatchIndex(text, query) >= 0
}

// HighlightMatches renders text with base, except every smartcase match of
// query is rendered with match instead. With an empty or absent query it is
// equivalent to base.Render(text). The wrapped text is expected to be plain
// (already wrapped by the caller); a match split across a wrap boundary is not
// highlighted.
func HighlightMatches(text, query string, base, match lipgloss.Style) string {
	idxs := MatchIndices(text, query)
	if len(idxs) == 0 {
		return base.Render(text)
	}
	var b strings.Builder
	prev := 0
	for _, r := range idxs {
		if r[0] > prev {
			b.WriteString(base.Render(text[prev:r[0]]))
		}
		b.WriteString(match.Render(text[r[0]:r[1]]))
		prev = r[1]
	}
	if prev < len(text) {
		b.WriteString(base.Render(text[prev:]))
	}
	return b.String()
}

func hasUpper(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}
