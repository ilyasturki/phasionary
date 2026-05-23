package app

import (
	"strings"

	"phasionary/internal/ui"
)

func (m model) estimatePickerView() string {
	lines := []string{ui.DialogTitleStyle.Render("Time Estimate"), ""}

	presetLabels := []string{
		"None",
		"15 minutes",
		"30 minutes",
		"1 hour",
		"2 hours",
		"4 hours",
		"1 day",
		"2 days",
		"3 days",
		"5 days",
	}

	for i, label := range presetLabels {
		prefix := "  "
		if i == m.ui.EstimatePicker.Selected {
			prefix = "> "
		}
		line := prefix + label
		if i == m.ui.EstimatePicker.Selected {
			line = ui.SelectedStyle.Render(line)
		}
		lines = append(lines, line)
	}

	lines = append(lines, "", ui.DialogHintStyle.Render("j/k navigate | enter select | esc cancel"))
	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}
