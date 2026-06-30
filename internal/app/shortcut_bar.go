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
	{Key: "/", Label: "search"},
	{Key: "v", Label: "visual"},
	{Key: "y/x/p", Label: "copy/cut/paste"},
	{Key: "gx", Label: "open url"},
	{Key: "gy", Label: "yank part"},
	{Key: "u/^r", Label: "undo/redo"},
	{Key: ",", Label: "options"},
	{Key: "q", Label: "quit"},
}

var visualShortcuts = []ui.Hint{
	{Key: "J/K", Label: "move"},
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

// bottomBarHeight is the rows reserved at the bottom of the screen: the search
// prompt while searching, otherwise the shortcut bar (when enabled).
func (m model) bottomBarHeight() int {
	if m.ui.Modes.IsSearch() {
		return 1
	}
	return m.shortcutBarHeight()
}

// renderBottomBar draws the search prompt while searching, otherwise the
// shortcut bar. The two never show at once (search isn't normal/visual mode).
func (m model) renderBottomBar() string {
	if m.ui.Modes.IsSearch() {
		return m.renderSearchBar()
	}
	return m.renderShortcutBar()
}
