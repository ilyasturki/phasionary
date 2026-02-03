package ui

import "github.com/charmbracelet/lipgloss"

var (
	HeaderStyle      = lipgloss.NewStyle().Bold(true)
	MutedStyle       = lipgloss.NewStyle().Faint(true)
	CategoryStyle    = lipgloss.NewStyle().Bold(true)
	SelectedStyle    = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("15")).Foreground(lipgloss.Color("0"))
	StatusLineStyle  = lipgloss.NewStyle().Faint(true)
	HelpDialogStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
	DialogTitleStyle = lipgloss.NewStyle().Bold(true)
	DialogHintStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	SuccessStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
)

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

func PriorityStyle(priority string) lipgloss.Style {
	switch priority {
	case "high":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	case "low":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	default:
		return lipgloss.NewStyle()
	}
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

func SelectedPriorityStyle(priority string) lipgloss.Style {
	base := lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("15"))
	switch priority {
	case "high":
		return base.Foreground(lipgloss.Color("1")) // red text
	case "low":
		return base.Foreground(lipgloss.Color("6")) // cyan text
	default:
		return base.Foreground(lipgloss.Color("0")) // black text for contrast
	}
}

func SelectedStatusStyle(status string) lipgloss.Style {
	base := lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("15"))
	switch status {
	case "in_progress":
		return base.Foreground(lipgloss.Color("4")) // blue text
	case "completed":
		return base.Foreground(lipgloss.Color("8")) // gray text
	case "cancelled":
		return base.Foreground(lipgloss.Color("1")) // red text
	default:
		return base.Foreground(lipgloss.Color("3")) // yellow text for todo
	}
}

// Unfocused selection style (underline, no background - works in light/dark modes)
var UnfocusedSelectedStyle = lipgloss.NewStyle().Bold(true).Underline(true)

// Unfocused cursor style for edit mode (underline only)
var UnfocusedCursorStyle = lipgloss.NewStyle().Underline(true)

func GetSelectedStyle(focused bool) lipgloss.Style {
	if focused {
		return SelectedStyle
	}
	return UnfocusedSelectedStyle
}

func GetSelectedStatusStyle(status string, focused bool) lipgloss.Style {
	if focused {
		return SelectedStatusStyle(status)
	}
	// Unfocused: use underline with status color
	switch status {
	case "in_progress":
		return lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("4"))
	case "completed":
		return lipgloss.NewStyle().Bold(true).Underline(true).Faint(true)
	case "cancelled":
		return lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("1"))
	default:
		return lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("3"))
	}
}

func GetSelectedPriorityStyle(priority string, focused bool) lipgloss.Style {
	if focused {
		return SelectedPriorityStyle(priority)
	}
	// Unfocused: use underline with priority color
	switch priority {
	case "high":
		return lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("1"))
	case "low":
		return lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("6"))
	default:
		return lipgloss.NewStyle().Bold(true).Underline(true)
	}
}

func GetCursorStyle(focused bool) lipgloss.Style {
	if focused {
		return SelectedStyle
	}
	return UnfocusedCursorStyle
}
