package app

import (
	"github.com/charmbracelet/bubbles/textinput"

	"phasionary/internal/domain"
)

type UIMode int

const (
	ModeNormal UIMode = iota
	ModeEdit
	ModeHelp
	ModeConfirmDelete
	ModeOptions
	ModeProjectPicker
)

type OptionsState struct {
	selectedOption int
}

type ProjectPickerState struct {
	projects     []domain.Project
	selected     int
	scrollOffset int
}

func (p *ProjectPickerState) reset() {
	p.projects = nil
	p.selected = 0
	p.scrollOffset = 0
}

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
