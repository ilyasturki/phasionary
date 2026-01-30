package components

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewModal(t *testing.T) {
	modal := NewModal(80, 24)
	assert.NotNil(t, modal)
}

func TestModal_SetSize(t *testing.T) {
	modal := NewModal(80, 24)
	modal.SetSize(100, 30)
	assert.Equal(t, 100, modal.width)
	assert.Equal(t, 30, modal.height)
}

func TestModal_Render(t *testing.T) {
	t.Run("returns overlay when dimensions are zero", func(t *testing.T) {
		modal := NewModal(0, 0)
		overlay := "test overlay"
		result := modal.Render("background", overlay)
		assert.Equal(t, overlay, result)
	})

	t.Run("centers overlay on background", func(t *testing.T) {
		modal := NewModal(10, 5)
		bg := strings.Repeat(".\n", 5)
		overlay := "X"
		result := modal.Render(bg, overlay)
		assert.Contains(t, result, "X")
	})

	t.Run("handles multi-line overlay", func(t *testing.T) {
		modal := NewModal(20, 10)
		bg := strings.Repeat(strings.Repeat(".", 20)+"\n", 10)
		overlay := "Line1\nLine2\nLine3"
		result := modal.Render(bg, overlay)
		assert.Contains(t, result, "Line1")
		assert.Contains(t, result, "Line2")
		assert.Contains(t, result, "Line3")
	})
}

func TestPlaceOverlay(t *testing.T) {
	t.Run("places overlay in center", func(t *testing.T) {
		bg := ".....\n.....\n.....\n.....\n....."
		fg := "X"
		result := placeOverlay(bg, fg, 5, 5)
		lines := strings.Split(result, "\n")
		assert.Equal(t, 5, len(lines))
		assert.Contains(t, lines[2], "X")
	})

	t.Run("handles overlay wider than single character", func(t *testing.T) {
		bg := "..........\n..........\n..........\n..........\n.........."
		fg := "ABC"
		result := placeOverlay(bg, fg, 10, 5)
		lines := strings.Split(result, "\n")
		assert.Contains(t, lines[2], "ABC")
	})
}
