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
	if query == "" {
		return nil
	}
	caseSensitive := hasUpper(query)
	hay := []rune(text)
	needle := []rune(query)
	if len(needle) == 0 || len(needle) > len(hay) {
		return nil
	}

	// Byte offset of every rune boundary in text, so the returned ranges slice
	// text correctly even when it contains multi-byte runes.
	offsets := make([]int, len(hay)+1)
	b := 0
	for i, r := range hay {
		offsets[i] = b
		b += utf8.RuneLen(r)
	}
	offsets[len(hay)] = b

	eq := func(a, c rune) bool {
		if caseSensitive {
			return a == c
		}
		return unicode.ToLower(a) == unicode.ToLower(c)
	}

	var out [][2]int
	for i := 0; i+len(needle) <= len(hay); {
		matched := true
		for j := 0; j < len(needle); j++ {
			if !eq(hay[i+j], needle[j]) {
				matched = false
				break
			}
		}
		if matched {
			out = append(out, [2]int{offsets[i], offsets[i+len(needle)]})
			i += len(needle)
		} else {
			i++
		}
	}
	return out
}

// Contains reports whether text contains query under smartcase semantics.
func Contains(text, query string) bool {
	return len(MatchIndices(text, query)) > 0
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
