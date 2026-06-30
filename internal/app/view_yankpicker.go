package app

import (
	"phasionary/internal/app/components"
	"phasionary/internal/ui"
)

func (m model) yankPickerView() string {
	return components.RenderListPicker(
		"Yank",
		m.ui.YankPicker.Labels(),
		m.ui.YankPicker.Selected,
		[]ui.Hint{{Key: "enter", Label: "copy"}, {Key: "esc", Label: "cancel"}},
	)
}
