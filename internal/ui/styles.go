package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

var (
	// Selection styles use Reverse so the terminal swaps fg/bg natively at
	// render time — this tracks live light/dark theme changes, which
	// lipgloss.AdaptiveColor cannot (background detection is cached once at
	// startup via sync.Once and never re-queried).
	HeaderStyle      = lipgloss.NewStyle().Bold(true)
	MutedStyle       = lipgloss.NewStyle().Faint(true)
	CategoryStyle    = lipgloss.NewStyle().Bold(true)
	SelectedStyle    = lipgloss.NewStyle().Bold(true).Reverse(true)
	StatusLineStyle  = lipgloss.NewStyle().Faint(true)
	HelpDialogStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
	DialogTitleStyle = lipgloss.NewStyle().Bold(true)
	DialogHintStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	SuccessStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	FilterTagStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)
	VisualTagStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("3")).Bold(true)

	// VisualSelectedStyle highlights rows that are part of an active visual
	// range. Distinct from the reverse-based cursor style so the user can
	// tell at a glance they're in visual mode.
	VisualSelectedStyle         = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")).Background(lipgloss.Color("3"))
	UnfocusedVisualSelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")).Background(lipgloss.Color("3")).Faint(true)

	// VisualCursorStyle marks the cursor row inside an active visual range.
	// Reverse swaps fg/bg natively so the cursor reads as a distinct block
	// against the yellow range band — same visual language as the normal-mode
	// cursor, which makes it obvious which end `j`/`k` extends and which end
	// `o` swaps to.
	VisualCursorStyle          = lipgloss.NewStyle().Bold(true).Reverse(true)
	UnfocusedVisualCursorStyle = lipgloss.NewStyle().Bold(true).Reverse(true).Faint(true)

	// CutBadgeStyle marks rows whose item is staged for cut/paste.
	CutBadgeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)
)

// CutMark is the badge appended to rows whose item is pending a cut/paste.
const CutMark = " ✂"

// ApplyCut decorates a style so cut rows read as "ghosted/in flight": faint
// plus italic so the differentiation survives even when the row is also
// selected (which inverts colors) or completed (which is already faint).
func ApplyCut(s lipgloss.Style) lipgloss.Style {
	return s.Faint(true).Italic(true)
}

func GetVisualSelectedStyle(focused bool) lipgloss.Style {
	if focused {
		return VisualSelectedStyle
	}
	return UnfocusedVisualSelectedStyle
}

func GetVisualCursorStyle(focused bool) lipgloss.Style {
	if focused {
		return VisualCursorStyle
	}
	return UnfocusedVisualCursorStyle
}

func StatusStyle(status string) lipgloss.Style {
	switch status {
	case "in_progress":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	case "completed":
		return lipgloss.NewStyle().Faint(true)
	case "cancelled":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	}
}

func priorityColor(priority string) (color.Color, bool) {
	switch priority {
	case "high":
		return lipgloss.Color("1"), true
	case "low":
		return lipgloss.Color("6"), true
	default:
		return nil, false
	}
}

// PriorityStyle returns the text style for a task title with the given
// priority. Color is applied only when colorMode is "full" (or empty, which
// is treated as the default "full").
func PriorityStyle(priority, colorMode string) lipgloss.Style {
	if colorMode == "" {
		colorMode = "full"
	}
	if colorMode != "full" {
		return lipgloss.NewStyle()
	}
	if c, ok := priorityColor(priority); ok {
		return lipgloss.NewStyle().Foreground(c)
	}
	return lipgloss.NewStyle()
}

// PriorityIconStyle returns the style for the priority icon. Color is applied
// when colorMode is "full" or "icon" (empty defaults to "full").
func PriorityIconStyle(priority, colorMode string) lipgloss.Style {
	if colorMode == "" {
		colorMode = "full"
	}
	if colorMode != "full" && colorMode != "icon" {
		return lipgloss.NewStyle()
	}
	if c, ok := priorityColor(priority); ok {
		return lipgloss.NewStyle().Foreground(c)
	}
	return lipgloss.NewStyle()
}

func TaskTitleStyle(priority, status, colorMode string) lipgloss.Style {
	base := PriorityStyle(priority, colorMode)
	if status == "completed" || status == "cancelled" {
		return base.Faint(true)
	}
	return base
}

func PriorityIcon(priority string) string {
	switch priority {
	case "high":
		return "▲"
	case "low":
		return "▼"
	default:
		return ""
	}
}

// Unfocused selection style (dimmed reverse — same visual language as focused,
// just quieter; lets the terminal swap fg/bg so it tracks light/dark themes).
var UnfocusedSelectedStyle = lipgloss.NewStyle().Bold(true).Reverse(true).Faint(true)

// Unfocused cursor style for edit mode (dimmed reverse)
var UnfocusedCursorStyle = lipgloss.NewStyle().Reverse(true).Faint(true)

func GetSelectedStyle(focused bool) lipgloss.Style {
	if focused {
		return SelectedStyle
	}
	return UnfocusedSelectedStyle
}

func GetSelectedStatusStyle(status string, focused bool) lipgloss.Style {
	return GetSelectedStyle(focused)
}

func GetSelectedPriorityStyle(priority string, focused bool) lipgloss.Style {
	return GetSelectedStyle(focused)
}

func GetCursorStyle(focused bool) lipgloss.Style {
	if focused {
		return SelectedStyle
	}
	return UnfocusedCursorStyle
}
