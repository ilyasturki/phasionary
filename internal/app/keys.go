package app

import (
	tea "github.com/charmbracelet/bubbletea"

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
	{keys: []string{"tab"}, display: "Tab/za", desc: "fold/unfold category", section: sectionNavigation,
		action: void((*model).toggleFold)},
	{keys: []string{"a"}, prefix: 'z', section: sectionNavigation,
		action: void((*model).toggleFold)},
	{keys: []string{"c"}, prefix: 'z', display: "zc", desc: "fold all categories", section: sectionNavigation,
		action: void((*model).foldAll)},
	{keys: []string{"o"}, prefix: 'z', display: "zo", desc: "unfold all categories", section: sectionNavigation,
		action: void((*model).unfoldAll)},
	{keys: []string{"P"}, desc: "switch project", section: sectionNavigation,
		action: void((*model).openProjectPicker)},

	// Actions
	{keys: []string{"a"}, desc: "add new task", section: sectionActions,
		action: void((*model).startAddingTask)},
	{keys: []string{"A"}, desc: "add new category", section: sectionActions,
		action: void((*model).startAddingCategory)},
	{keys: []string{"enter"}, desc: "edit selected item", section: sectionActions,
		action: void((*model).startEditing)},
	{keys: []string{"e"}, desc: "edit in external editor", section: sectionActions,
		action: func(m *model) tea.Cmd { return m.startExternalEdit() }},
	{keys: []string{" "}, display: "space", desc: "toggle task status", section: sectionActions,
		action: void((*model).toggleSelectedTask)},
	{keys: []string{"J"}, display: "J/K", desc: "reorder task/category up/down", section: sectionActions,
		action: void(moveDown)},
	{keys: []string{"K"}, section: sectionActions,
		action: void(moveUp)},
	{keys: []string{"s"}, display: "s/S", desc: "sort tasks by status", section: sectionActions,
		action: void((*model).sortTasksByStatus)},
	{keys: []string{"S"}, section: sectionActions,
		action: void((*model).sortTasksByStatusReverse)},
	{keys: []string{"f"}, desc: "filter tasks by status", section: sectionActions,
		action: func(m *model) tea.Cmd { m.ui.Modes.ToFilter(); return nil }},
	{keys: []string{"h"}, display: "h/l", desc: "change priority", section: sectionActions,
		action: void((*model).decreasePriority)},
	{keys: []string{"l"}, section: sectionActions,
		action: void((*model).increasePriority)},
	{keys: []string{"t"}, desc: "set time estimate", section: sectionActions,
		action: void((*model).openEstimatePicker)},
	{keys: []string{"y"}, desc: "copy selected text", section: sectionActions,
		action: func(m *model) tea.Cmd { return m.copySelected() }},
	{keys: []string{"Y"}, section: sectionActions,
		action: func(m *model) tea.Cmd { return m.copyCategoryContent() }},
	{keys: []string{"x"}, desc: "mark task for cut", section: sectionActions,
		action: void((*model).cutSelectedTask)},
	{keys: []string{"p"}, desc: "paste cut task", section: sectionActions,
		action: void((*model).pasteTask)},
	{keys: []string{"d"}, desc: "delete selected item", section: sectionActions,
		action: void((*model).deleteSelected)},
	{keys: []string{"i"}, desc: "show item info", section: sectionActions,
		action: func(m *model) tea.Cmd { m.ui.Modes.ToInfo(); return nil }},
	{keys: []string{"x"}, prefix: 'g', display: "gx", desc: "open URL in focused task", section: sectionActions,
		action: func(m *model) tea.Cmd { return m.openLinksForSelected() }},
	{keys: []string{"o"}, desc: "options", section: sectionActions,
		action: func(m *model) tea.Cmd {
			m.ui.Modes.ToOptions()
			m.ui.Options = OptionsState{selectedOption: 0}
			return nil
		}},
	{keys: []string{"?"}, desc: "toggle help", section: sectionActions,
		action: func(m *model) tea.Cmd { m.ui.Modes.ToggleHelp(); return nil }},
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

func moveDown(m *model) {
	if pos, ok := m.selectedPosition(); ok && pos.Kind == selection.FocusCategory {
		m.moveCategoryDown()
	} else {
		m.moveTaskDown()
	}
}

func moveUp(m *model) {
	if pos, ok := m.selectedPosition(); ok && pos.Kind == selection.FocusCategory {
		m.moveCategoryUp()
	} else {
		m.moveTaskUp()
	}
}

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
