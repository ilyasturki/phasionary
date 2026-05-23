package app

import (
	"phasionary/internal/app/components"
)

var estimatePresetLabels = []string{
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

func (m model) estimatePickerView() string {
	return components.RenderListPicker(
		"Time Estimate",
		estimatePresetLabels,
		m.ui.EstimatePicker.Selected,
		"j/k navigate | enter select | esc cancel",
	)
}
