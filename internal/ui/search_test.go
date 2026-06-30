package ui

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"charm.land/lipgloss/v2"
)

func TestMatchIndices_CaseInsensitiveByDefault(t *testing.T) {
	got := MatchIndices("Add dark mode", "dark")
	assert.Equal(t, [][2]int{{4, 8}}, got)

	// Lowercase query matches mixed-case text.
	got = MatchIndices("Dark Mode", "dark")
	assert.Equal(t, [][2]int{{0, 4}}, got)
}

func TestMatchIndices_SmartcaseUppercaseIsSensitive(t *testing.T) {
	// An uppercase rune in the query makes the search case-sensitive.
	assert.Empty(t, MatchIndices("dark", "Dark"))
	assert.Equal(t, [][2]int{{0, 4}}, MatchIndices("Dark", "Dark"))
}

func TestMatchIndices_MultipleNonOverlapping(t *testing.T) {
	got := MatchIndices("aa bb aa", "aa")
	assert.Equal(t, [][2]int{{0, 2}, {6, 8}}, got)
}

func TestMatchIndices_Empty(t *testing.T) {
	assert.Nil(t, MatchIndices("anything", ""))
	assert.Nil(t, MatchIndices("", "x"))
	assert.Nil(t, MatchIndices("ab", "abc")) // needle longer than haystack
}

func TestMatchIndices_MultiByteOffsets(t *testing.T) {
	// "café" — the match after the multi-byte é must use byte offsets.
	text := "café bar"
	got := MatchIndices(text, "bar")
	assert.Len(t, got, 1)
	assert.Equal(t, "bar", text[got[0][0]:got[0][1]])
}

func TestContains_Smartcase(t *testing.T) {
	assert.True(t, Contains("Hello World", "world"))
	assert.False(t, Contains("Hello World", "World!"))
	assert.False(t, Contains("hello", "HELLO"))
}

func TestHighlightMatches_PreservesVisibleText(t *testing.T) {
	base := lipgloss.NewStyle()
	out := HighlightMatches("hello world", "wor", base, SearchMatchStyle)
	// Whatever the active color profile does to the styling, the visible text is
	// unchanged.
	assert.Equal(t, "hello world", ansi.Strip(out))
}

func TestHighlightMatches_NoMatchIsPlainRender(t *testing.T) {
	base := lipgloss.NewStyle().Bold(true)
	assert.Equal(t, base.Render("hello"), HighlightMatches("hello", "zzz", base, SearchMatchStyle))
	assert.Equal(t, base.Render("hello"), HighlightMatches("hello", "", base, SearchMatchStyle))
}
