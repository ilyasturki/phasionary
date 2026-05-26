package app

import (
	"phasionary/internal/app/components"
	"phasionary/internal/ui"
)

func (m model) urlPickerView() string {
	return components.RenderListPicker(
		"Open URL",
		m.ui.URLPicker.URLs,
		m.ui.URLPicker.Selected,
		[]ui.Hint{{Key: "enter", Label: "open"}, {Key: "esc", Label: "cancel"}},
	)
}
