package app

import (
	tea "charm.land/bubbletea/v2"

	"phasionary/internal/app/selection"
)

type bindingAction func(*model) tea.Cmd

type keyBinding struct {
	keys    []string
	prefix  rune
	display string
	desc    string
	section string
	action  bindingAction
}

const (
	sectionNavigation = "Navigation"
	sectionActions    = "Actions"
)

func void(fn func(*model)) bindingAction {
	return func(m *model) tea.Cmd {
		fn(m)
		return nil
	}
}

var normalBindings = []keyBinding{
	// Navigation
	{keys: []string{"up", "k"}, display: "j/k or ↑/↓", desc: "move selection", section: sectionNavigation,
		action: func(m *model) tea.Cmd { m.moveSelection(-1); return nil }},
	{keys: []string{"down", "j"}, section: sectionNavigation,
		action: func(m *model) tea.Cmd { m.moveSelection(1); return nil }},
	{keys: []string{"ctrl+d"}, display: "ctrl+d/u", desc: "half-page down/up", section: sectionNavigation,
		action: func(m *model) tea.Cmd { m.moveSelectionByPage(0.5); return nil }},
	{keys: []string{"ctrl+u"}, section: sectionNavigation,
		action: func(m *model) tea.Cmd { m.moveSelectionByPage(-0.5); return nil }},
	{keys: []string{"ctrl+f"}, display: "ctrl+f/b", desc: "full-page down/up", section: sectionNavigation,
		action: func(m *model) tea.Cmd { m.moveSelectionByPage(1.0); return nil }},
	{keys: []string{"ctrl+b"}, section: sectionNavigation,
		action: func(m *model) tea.Cmd { m.moveSelectionByPage(-1.0); return nil }},
	{keys: []string{"g"}, prefix: 'g', display: "gg", desc: "jump to first item", section: sectionNavigation,
		action: void((*model).jumpToFirst)},
	{keys: []string{"G"}, desc: "jump to last item", section: sectionNavigation,
		action: void((*model).jumpToLast)},
	{keys: []string{"}"}, display: "}/{", desc: "next/previous category", section: sectionNavigation,
		action: void((*model).jumpToNextCategory)},
	{keys: []string{"{"}, section: sectionNavigation,
		action: void((*model).jumpToPrevCategory)},
	{keys: []string{"z"}, prefix: 'z', display: "zz", desc: "center selection on screen", section: sectionNavigation,
		action: void((*model).centerOnSelected)},
	{keys: []string{"t"}, prefix: 'z', display: "zt", desc: "scroll selection to top", section: sectionNavigation,
		action: void((*model).topOnSelected)},
	{keys: []string{"b"}, prefix: 'z', display: "zb", desc: "scroll selection to bottom", section: sectionNavigation,
		action: void((*model).bottomOnSelected)},
	{keys: []string{"tab"}, display: "Tab/za", desc: "fold/unfold category", section: sectionNavigation,
		action: void((*model).toggleFold)},
	{keys: []string{"a"}, prefix: 'z', section: sectionNavigation,
		action: void((*model).toggleFold)},
	{keys: []string{"c"}, prefix: 'z', display: "zc", desc: "fold all categories", section: sectionNavigation,
		action: void((*model).foldAll)},
	{keys: []string{"o"}, prefix: 'z', display: "zo", desc: "unfold all categories", section: sectionNavigation,
		action: void((*model).unfoldAll)},
	{keys: []string{"d"}, prefix: 'z', display: "zd", desc: "toggle inline descriptions", section: sectionNavigation,
		action: void((*model).toggleExpandDescriptions)},
	{keys: []string{"ctrl+p"}, desc: "switch project", section: sectionNavigation,
		action: void((*model).openProjectPicker)},
	{keys: []string{"/"}, desc: "search text", section: sectionNavigation,
		action: func(m *model) tea.Cmd { return m.startSearch() }},
	{keys: []string{"n"}, display: "n/N", desc: "next/previous search match", section: sectionNavigation,
		action: void((*model).searchNext)},
	{keys: []string{"N"}, section: sectionNavigation,
		action: void((*model).searchPrev)},

	// Actions
	{keys: []string{"a"}, desc: "add new task", section: sectionActions,
		action: void((*model).startAddingTask)},
	{keys: []string{"A"}, desc: "add new category", section: sectionActions,
		action: void((*model).startAddingCategory)},
	{keys: []string{"-"}, desc: "insert separator below", section: sectionActions,
		action: void((*model).startAddingSeparator)},
	{keys: []string{"enter"}, desc: "edit selected item", section: sectionActions,
		action: func(m *model) tea.Cmd {
			if pos, ok := m.selectedPosition(); ok && pos.Kind == selection.FocusDescription {
				return m.startDescriptionInlineEdit(pos.CategoryIndex, pos.TaskIndex)
			}
			m.startEditing()
			return nil
		}},
	// Many terminals (kitty, ghostty, …) transmit Shift+Enter as the bytes
	// ESC+CR, which Bubble Tea decodes as "alt+enter"; others that speak the
	// Kitty protocol send a genuine "shift+enter". Bind both so the physical
	// Shift+Enter reaches this action regardless of terminal encoding.
	{keys: []string{"shift+enter", "alt+enter"}, display: "shift+enter", desc: "edit/jump to task description", section: sectionActions,
		action: func(m *model) tea.Cmd { return m.editOrFocusDescription() }},
	{keys: []string{"e"}, desc: "edit in external editor (whole task)", section: sectionActions,
		action: func(m *model) tea.Cmd { return m.startExternalEdit() }},
	{keys: []string{"space"}, display: "space/shift+space", desc: "cycle task status forward/back", section: sectionActions,
		action: void((*model).toggleSelectedTask)},
	{keys: []string{"shift+space"}, section: sectionActions,
		action: void((*model).toggleSelectedTaskReverse)},
	{keys: []string{"J"}, display: "J/K", desc: "reorder task/category up/down", section: sectionActions,
		action: func(m *model) tea.Cmd { m.moveSelectedRow(+1); return nil }},
	{keys: []string{"K"}, section: sectionActions,
		action: func(m *model) tea.Cmd { m.moveSelectedRow(-1); return nil }},
	{keys: []string{"S"}, desc: "reverse category order", section: sectionActions,
		action: void((*model).reverseCategories)},
	{keys: []string{"f"}, desc: "filter tasks", section: sectionActions,
		action: func(m *model) tea.Cmd { m.ui.Modes.ToFilter(); return nil }},
	{keys: []string{"h"}, display: "h/l", desc: "change priority", section: sectionActions,
		action: void((*model).decreasePriority)},
	{keys: []string{"l"}, section: sectionActions,
		action: void((*model).increasePriority)},
	{keys: []string{"ctrl+t"}, desc: "set time estimate", section: sectionActions,
		action: void((*model).openEstimatePicker)},
	{keys: []string{"t"}, display: "t/T", desc: "cycle tag color / edit tag (color + label)", section: sectionActions,
		action: void((*model).cycleTag)},
	{keys: []string{"T"}, section: sectionActions,
		action: func(m *model) tea.Cmd { return m.startTagEdit() }},
	{keys: []string{"y"}, desc: "copy selected text", section: sectionActions,
		action: func(m *model) tea.Cmd { return m.copySelected() }},
	{keys: []string{"Y"}, section: sectionActions,
		action: func(m *model) tea.Cmd { return m.copyCategoryContent() }},
	{keys: []string{"x"}, desc: "mark task for cut", section: sectionActions,
		action: void((*model).cutSelectedTask)},
	{keys: []string{"p"}, desc: "paste copied tag (if copied last) or cut task", section: sectionActions,
		action: void((*model).paste)},
	{keys: []string{"v"}, desc: "visual select (extend with j/k, then y/Y/x)", section: sectionActions,
		action: void((*model).enterVisualMode)},
	{keys: []string{"d"}, desc: "delete selected item", section: sectionActions,
		action: void((*model).deleteSelected)},
	{keys: []string{"i"}, prefix: 'g', display: "gi", desc: "show item info", section: sectionActions,
		action: func(m *model) tea.Cmd { m.ui.Modes.ToInfo(); return nil }},
	{keys: []string{"r"}, desc: "reload project from disk", section: sectionActions,
		action: void((*model).reloadProject)},
	{keys: []string{"u"}, desc: "undo last change", section: sectionActions,
		action: void((*model).undo)},
	{keys: []string{"ctrl+r"}, desc: "redo last undone change", section: sectionActions,
		action: void((*model).redo)},
	{keys: []string{"x"}, prefix: 'g', display: "gx", desc: "open URL in focused task", section: sectionActions,
		action: func(m *model) tea.Cmd { return m.openLinksForSelected() }},
	{keys: []string{"y"}, prefix: 'g', display: "gy", desc: "yank a part of focused item (uuid, url, title…)", section: sectionActions,
		action: func(m *model) tea.Cmd { return m.yankPartForSelected() }},
	{keys: []string{"t"}, prefix: 'g', display: "gt", desc: "copy tag from focused task (label to clipboard, paste with p)", section: sectionActions,
		action: func(m *model) tea.Cmd { return m.copyTagFromSelected() }},
	{keys: []string{","}, desc: "options", section: sectionActions,
		action: func(m *model) tea.Cmd {
			m.ui.Modes.ToOptions()
			m.ui.Options = OptionsState{selectedOption: 0}
			return nil
		}},
	{keys: []string{"?"}, desc: "toggle help", section: sectionActions,
		action: func(m *model) tea.Cmd {
			m.ui.Modes.ToggleHelp()
			if m.ui.Modes.IsHelp() {
				m.ui.Help = HelpState{}
			}
			return nil
		}},
	{keys: []string{"q", "ctrl+c"}, desc: "quit", section: sectionActions,
		action: func(m *model) tea.Cmd { return tea.Quit }},
}

var chordPrefixes = func() map[rune]struct{} {
	set := make(map[rune]struct{})
	for _, b := range normalBindings {
		if b.prefix != 0 {
			set[b.prefix] = struct{}{}
		}
	}
	return set
}()

func matchBinding(prefix rune, key string) *keyBinding {
	for i := range normalBindings {
		b := &normalBindings[i]
		if b.prefix != prefix {
			continue
		}
		for _, k := range b.keys {
			if k == key {
				return b
			}
		}
	}
	return nil
}

func (m *model) dispatchNormalKey(key string) tea.Cmd {
	// Esc cancels a pending chord and clears an active search highlight (the
	// vim `:nohlsearch` gesture). It otherwise stays a no-op in normal mode.
	if key == "esc" {
		m.ui.Screen.PendingKey = 0
		if m.ui.Search.query != "" {
			m.clearSearch()
		}
		return nil
	}
	if m.ui.Screen.PendingKey != 0 {
		if b := matchBinding(m.ui.Screen.PendingKey, key); b != nil {
			m.ui.Screen.PendingKey = 0
			return b.action(m)
		}
		m.ui.Screen.PendingKey = 0
	}

	runes := []rune(key)
	if len(runes) == 1 {
		if _, isChord := chordPrefixes[runes[0]]; isChord {
			if matchBinding(0, key) == nil {
				m.ui.Screen.PendingKey = runes[0]
				return nil
			}
		}
	}

	if b := matchBinding(0, key); b != nil {
		return b.action(m)
	}
	return nil
}
