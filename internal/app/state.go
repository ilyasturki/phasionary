package app

import "github.com/charmbracelet/bubbles/textinput"

type UIMode int

const (
	ModeNormal UIMode = iota
	ModeEdit
	ModeHelp
	ModeConfirmDelete
)

type EditState struct {
	input     textinput.Model
	isAdding  bool
	newItemID string
	itemType  focusKind
}

func (e *EditState) reset() {
	e.input = textinput.New()
	e.isAdding = false
	e.newItemID = ""
	e.itemType = focusProject
}

func newEditState(value string, isAdding bool, newID string, kind focusKind) EditState {
	ti := textinput.New()
	ti.SetValue(value)
	ti.SetCursor(len([]rune(value)))
	ti.Focus()
	return EditState{
		input:     ti,
		isAdding:  isAdding,
		newItemID: newID,
		itemType:  kind,
	}
}
