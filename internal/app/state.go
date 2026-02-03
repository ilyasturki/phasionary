package app

import (
	"github.com/charmbracelet/bubbles/textinput"

	"phasionary/internal/domain"
)

var filterStatuses = []string{
	domain.StatusTodo,
	domain.StatusInProgress,
	domain.StatusCompleted,
	domain.StatusCancelled,
}

type FilterState struct {
	selected int
	enabled  map[string]bool
}

func NewFilterState() FilterState {
	return FilterState{
		selected: 0,
		enabled:  make(map[string]bool),
	}
}

func (f *FilterState) IsStatusVisible(status string) bool {
	if len(f.enabled) == 0 {
		return true
	}
	return f.enabled[status]
}

func (f *FilterState) Toggle(status string) {
	if f.enabled[status] {
		delete(f.enabled, status)
	} else {
		f.enabled[status] = true
	}
}

func (f *FilterState) HasActiveFilter() bool {
	return len(f.enabled) > 0
}

func (f *FilterState) MoveUp() {
	if f.selected > 0 {
		f.selected--
	}
}

func (f *FilterState) MoveDown() {
	if f.selected < len(filterStatuses)-1 {
		f.selected++
	}
}

func (f *FilterState) ToggleSelected() {
	if f.selected >= 0 && f.selected < len(filterStatuses) {
		f.Toggle(filterStatuses[f.selected])
	}
}

func (f *FilterState) Selected() int {
	return f.selected
}

func (f *FilterState) IsEnabled(status string) bool {
	return f.enabled[status]
}

type OptionsState struct {
	selectedOption int
}

type ProjectPickerState struct {
	projects        []domain.Project
	selected        int
	scrollOffset    int
	isAdding        bool
	input           textinput.Model
	pendingDeleteID string
}

func (p *ProjectPickerState) reset() {
	p.projects = nil
	p.selected = 0
	p.scrollOffset = 0
	p.isAdding = false
	p.input = textinput.Model{}
	p.pendingDeleteID = ""
}

func (p *ProjectPickerState) totalItems() int {
	return len(p.projects) + 1
}

func (p *ProjectPickerState) isOnAddButton() bool {
	return p.selected == len(p.projects)
}

func (p *ProjectPickerState) startAdding() {
	p.isAdding = true
	p.input = textinput.New()
	p.input.Focus()
}

func (p *ProjectPickerState) cancelAdding() {
	p.isAdding = false
	p.input = textinput.Model{}
}

type FoldState struct {
	folded map[string]bool
}

func NewFoldState() FoldState {
	return FoldState{
		folded: make(map[string]bool),
	}
}

func (f *FoldState) IsFolded(categoryID string) bool {
	return f.folded[categoryID]
}

func (f *FoldState) Toggle(categoryID string) {
	if f.folded[categoryID] {
		delete(f.folded, categoryID)
	} else {
		f.folded[categoryID] = true
	}
}

func (f *FoldState) FoldAll(categoryIDs []string) {
	for _, id := range categoryIDs {
		f.folded[id] = true
	}
}

func (f *FoldState) UnfoldAll() {
	f.folded = make(map[string]bool)
}

func (f *FoldState) HasFolded() bool {
	return len(f.folded) > 0
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
