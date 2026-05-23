package app

import (
	"phasionary/internal/app/components"
)

func (m model) urlPickerView() string {
	return components.RenderListPicker(
		"Open URL",
		m.ui.URLPicker.URLs,
		m.ui.URLPicker.Selected,
		"j/k navigate | enter open | esc cancel",
	)
}
