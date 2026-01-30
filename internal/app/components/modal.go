package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type Modal struct {
	width  int
	height int
}

func NewModal(width, height int) *Modal {
	return &Modal{
		width:  width,
		height: height,
	}
}

func (m *Modal) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *Modal) Render(background, overlay string) string {
	if m.width <= 0 || m.height <= 0 {
		return overlay
	}
	bgPlaced := lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, background)
	return placeOverlay(bgPlaced, overlay, m.width, m.height)
}

func placeOverlay(bg, fg string, width, height int) string {
	bgLines := strings.Split(bg, "\n")
	fgLines := strings.Split(fg, "\n")
	fgW := lipgloss.Width(fg)
	fgH := len(fgLines)
	startY := max(0, (height-fgH)/2)
	startX := max(0, (width-fgW)/2)
	for i, fgLine := range fgLines {
		y := startY + i
		if y >= len(bgLines) {
			break
		}
		left := ansi.Truncate(bgLines[y], startX, "")
		if w := ansi.StringWidth(left); w < startX {
			left += strings.Repeat(" ", startX-w)
		}
		right := ansi.TruncateLeft(bgLines[y], startX+fgW, "")
		bgLines[y] = left + fgLine + right
	}
	return strings.Join(bgLines, "\n")
}
