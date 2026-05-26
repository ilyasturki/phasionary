package app

import (
	"phasionary/internal/app/modes"
	"phasionary/internal/ui"
)

// Bottom "lazygit-style" hint footer. Only keeps shortcuts that aren't
// obvious from on-screen affordances or vim/arrow conventions (j/k, enter,
// arrows). Bindings already surfaced inside dialogs stay out.
var normalShortcuts = []ui.Hint{
	{Key: "?", Label: "help"},
	{Key: "a", Label: "add"},
	{Key: "A", Label: "category"},
	{Key: "f", Label: "filter"},
	{Key: "v", Label: "visual"},
	{Key: "y/x/p", Label: "copy/cut/paste"},
	{Key: "gx", Label: "open url"},
	{Key: "u/^r", Label: "undo/redo"},
	{Key: ",", Label: "options"},
	{Key: "q", Label: "quit"},
}

var visualShortcuts = []ui.Hint{
	{Key: "y", Label: "copy"},
	{Key: "Y", Label: "markdown"},
	{Key: "x", Label: "cut"},
	{Key: "d", Label: "delete"},
	{Key: "esc", Label: "exit"},
}

func (m model) shortcutBarVisible() bool {
	if !m.deps.CfgManager.Get().ShowShortcutBar {
		return false
	}
	switch m.ui.Modes.Current() {
	case modes.ModeNormal, modes.ModeVisual:
		return true
	default:
		return false
	}
}

func (m model) shortcutBarHeight() int {
	if m.shortcutBarVisible() {
		return 1
	}
	return 0
}

func (m model) renderShortcutBar() string {
	if !m.shortcutBarVisible() {
		return ""
	}
	hints := normalShortcuts
	if m.ui.Modes.IsVisual() {
		hints = visualShortcuts
	}
	return ui.RenderHintsToWidth(hints, m.ui.Screen.Width)
}
