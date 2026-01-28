package ui

import "github.com/charmbracelet/lipgloss"

var (
	HeaderStyle     = lipgloss.NewStyle().Bold(true)
	MutedStyle      = lipgloss.NewStyle().Faint(true)
	CategoryStyle   = lipgloss.NewStyle().Bold(true)
	SelectedStyle   = lipgloss.NewStyle().Bold(true).Reverse(true)
	StatusLineStyle = lipgloss.NewStyle().Faint(true)
	HelpDialogStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
)

func StatusStyle(status string) lipgloss.Style {
	switch status {
	case "in_progress":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	case "completed":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	case "cancelled":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	default:
		return lipgloss.NewStyle()
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
