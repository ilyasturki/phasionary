package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// words builds a space-separated filler string of n runes ending in a sentinel,
// so a test can tell which part of a long string a window landed on.
func words(n int) string {
	var b strings.Builder
	for i := 0; b.Len() < n; i++ {
		b.WriteString("word ")
	}
	return b.String()[:n-4] + "ZEND"
}

func TestWrapClamped(t *testing.T) {
	t.Run("leaves text that already fits alone", func(t *testing.T) {
		assert.Equal(t, []string{"a short title"}, WrapClamped("a short title", 20, 3, -1))
	})

	t.Run("wraps up to maxRows without marking anything", func(t *testing.T) {
		lines := WrapClamped("one two three four", 10, 3, -1)
		assert.Equal(t, []string{"one two", "three four"}, lines)
	})

	t.Run("clamps to maxRows and marks the cut", func(t *testing.T) {
		lines := WrapClamped(words(4000), 20, 3, -1)
		require.Len(t, lines, 3)
		assert.True(t, strings.HasSuffix(lines[2], Ellipsis), "last row should be marked elided: %q", lines[2])
		for _, l := range lines {
			assert.LessOrEqual(t, len([]rune(l)), 20)
		}
	})

	t.Run("marks the cut even when the last row fills the width exactly", func(t *testing.T) {
		// "abcd abcd abcd" wraps to rows of exactly 4; without forcing the mark,
		// a full-width row is not longer than the limit and goes unmarked.
		lines := WrapClamped("abcd abcd abcd abcd abcd", 4, 3, -1)
		require.Len(t, lines, 3)
		assert.True(t, strings.HasSuffix(lines[2], Ellipsis), "got %q", lines[2])
	})

	t.Run("aims the window at focus and marks both ends", func(t *testing.T) {
		text := words(4000)
		focus := FirstMatchIndex(text, "ZEND")
		require.Positive(t, focus)

		lines := WrapClamped(text, 20, 3, focus)
		require.Len(t, lines, 3)
		assert.Contains(t, strings.Join(lines, ""), "ZEND", "the focused match should be inside the window")
		assert.True(t, strings.HasPrefix(lines[0], Ellipsis), "text before the window should be marked: %q", lines[0])
	})

	t.Run("keeps the row count independent of focus", func(t *testing.T) {
		// The layout sizes a row without knowing where a search match sits inside
		// it, so a shifted window must not change how many rows the text renders
		// into — otherwise the reservation and the render disagree.
		text := words(4000)
		want := CountClamped(text, 20, 3)
		for _, focus := range []int{-1, 0, 1, 500, 2000, len([]rune(text)) - 1} {
			assert.Len(t, WrapClamped(text, 20, 3, focus), want, "focus=%d", focus)
		}
	})

	t.Run("survives degenerate widths and empty text", func(t *testing.T) {
		assert.Len(t, WrapClamped(words(400), 0, 3, -1), 3)
		assert.Len(t, WrapClamped(words(400), 1, 1, -1), 1)
		assert.Equal(t, []string{""}, WrapClamped("", 20, 3, -1))
	})

	t.Run("costs the same whatever the length", func(t *testing.T) {
		// Not a timing assertion — the guarantee is structural: the window is
		// sliced out before wrapping, so the work is bounded by the screen.
		short := CountClamped(words(400), 20, 3)
		long := CountClamped(strings.Repeat(words(4000), 500), 20, 3)
		assert.Equal(t, short, long)
	})
}

func TestFirstMatchIndex(t *testing.T) {
	t.Run("returns a byte offset, so it can slice the text directly", func(t *testing.T) {
		assert.Equal(t, 6, FirstMatchIndex("ééét", "t"))
	})

	t.Run("is smartcase, like the rest of search", func(t *testing.T) {
		assert.Equal(t, 6, FirstMatchIndex("hello World", "world"))
		assert.Equal(t, -1, FirstMatchIndex("hello world", "World"))
	})

	t.Run("reports the first of several matches", func(t *testing.T) {
		assert.Equal(t, 2, FirstMatchIndex("aabaab", "b"))
	})

	t.Run("agrees with MatchIndices", func(t *testing.T) {
		text := "un café, deux cafés"
		idxs := MatchIndices(text, "café")
		require.NotEmpty(t, idxs)
		assert.Equal(t, idxs[0][0], FirstMatchIndex(text, "café"))
	})

	t.Run("reports -1 when absent or empty", func(t *testing.T) {
		assert.Equal(t, -1, FirstMatchIndex("abc", "z"))
		assert.Equal(t, -1, FirstMatchIndex("abc", ""))
		assert.Equal(t, -1, FirstMatchIndex("ab", "abc"))
	})
}
