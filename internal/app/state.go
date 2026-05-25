package app

import (
	"github.com/charmbracelet/bubbles/textinput"

	"phasionary/internal/app/selection"
	"phasionary/internal/domain"
)

var filterStatuses = []string{
	domain.StatusTodo,
	domain.StatusInProgress,
	domain.StatusCompleted,
	domain.StatusCancelled,
}

var filterPriorities = []string{
	domain.PriorityHigh,
	domain.PriorityMedium,
	domain.PriorityLow,
	"",
}

type FilterView int

const (
	FilterViewHub FilterView = iota
	FilterViewStatus
	FilterViewPriority
	FilterViewCategory
)

const (
	FilterHubStatus int = iota
	FilterHubPriority
	FilterHubCategory
	FilterHubClearAll
)

const filterHubCount = 4

type FilterState struct {
	view             FilterView
	hubSelected      int
	statusSelected   int
	prioritySelected int
	categorySelected int
	statuses         map[string]bool
	priorities       map[string]bool
	categories       map[string]bool
}

func NewFilterState() FilterState {
	return FilterState{
		view:       FilterViewHub,
		statuses:   make(map[string]bool),
		priorities: make(map[string]bool),
		categories: make(map[string]bool),
	}
}

func (f *FilterState) View() FilterView {
	return f.view
}

func (f *FilterState) SetView(v FilterView) {
	f.view = v
}

func (f *FilterState) ResetToHub() {
	f.view = FilterViewHub
}

func (f *FilterState) HubSelected() int {
	return f.hubSelected
}

func (f *FilterState) Selected() int {
	switch f.view {
	case FilterViewStatus:
		return f.statusSelected
	case FilterViewPriority:
		return f.prioritySelected
	case FilterViewCategory:
		return f.categorySelected
	default:
		return f.hubSelected
	}
}

func (f *FilterState) MoveUp() {
	switch f.view {
	case FilterViewHub:
		if f.hubSelected > 0 {
			f.hubSelected--
		}
	case FilterViewStatus:
		if f.statusSelected > 0 {
			f.statusSelected--
		}
	case FilterViewPriority:
		if f.prioritySelected > 0 {
			f.prioritySelected--
		}
	case FilterViewCategory:
		if f.categorySelected > 0 {
			f.categorySelected--
		}
	}
}

func (f *FilterState) MoveDown(catCount int) {
	switch f.view {
	case FilterViewHub:
		if f.hubSelected < filterHubCount-1 {
			f.hubSelected++
		}
	case FilterViewStatus:
		if f.statusSelected < len(filterStatuses)-1 {
			f.statusSelected++
		}
	case FilterViewPriority:
		if f.prioritySelected < len(filterPriorities)-1 {
			f.prioritySelected++
		}
	case FilterViewCategory:
		if catCount > 0 && f.categorySelected < catCount-1 {
			f.categorySelected++
		}
	}
}

func (f *FilterState) ToggleSelected(categories []domain.Category) {
	switch f.view {
	case FilterViewStatus:
		if f.statusSelected >= 0 && f.statusSelected < len(filterStatuses) {
			toggleSetMember(f.statuses, filterStatuses[f.statusSelected])
		}
	case FilterViewPriority:
		if f.prioritySelected >= 0 && f.prioritySelected < len(filterPriorities) {
			toggleSetMember(f.priorities, filterPriorities[f.prioritySelected])
		}
	case FilterViewCategory:
		if f.categorySelected >= 0 && f.categorySelected < len(categories) {
			toggleSetMember(f.categories, categories[f.categorySelected].ID)
		}
	}
}

func toggleSetMember(set map[string]bool, key string) {
	if set[key] {
		delete(set, key)
	} else {
		set[key] = true
	}
}

func (f *FilterState) IsStatusEnabled(status string) bool       { return f.statuses[status] }
func (f *FilterState) IsPriorityEnabled(priority string) bool   { return f.priorities[priority] }
func (f *FilterState) IsCategoryEnabled(categoryID string) bool { return f.categories[categoryID] }

func (f *FilterState) StatusCount() int   { return len(f.statuses) }
func (f *FilterState) PriorityCount() int { return len(f.priorities) }
func (f *FilterState) CategoryCount() int { return len(f.categories) }

func (f *FilterState) HasActiveFilter() bool {
	return len(f.statuses) > 0 || len(f.priorities) > 0 || len(f.categories) > 0
}

func (f *FilterState) TaskVisible(task domain.Task, categoryID string) bool {
	if !f.HasActiveFilter() {
		return true
	}
	if len(f.statuses) > 0 && !f.statuses[task.Status] {
		return false
	}
	if len(f.priorities) > 0 && !f.priorities[task.Priority] {
		return false
	}
	if len(f.categories) > 0 && !f.categories[categoryID] {
		return false
	}
	return true
}

func (f *FilterState) ClearAll() {
	f.statuses = make(map[string]bool)
	f.priorities = make(map[string]bool)
	f.categories = make(map[string]bool)
	f.statusSelected = 0
	f.prioritySelected = 0
	f.categorySelected = 0
}

type OptionsState struct {
	selectedOption int
}

type ProjectPickerState struct {
	projects     []domain.Project
	selected     int
	scrollOffset int
	isAdding     bool
	input        textinput.Model
}

func (p *ProjectPickerState) reset() {
	p.projects = nil
	p.selected = 0
	p.scrollOffset = 0
	p.isAdding = false
	p.input = textinput.Model{}
}

type ConfirmDeleteKind int

const (
	ConfirmDeleteSelection ConfirmDeleteKind = iota
	ConfirmDeleteProject
)

type ConfirmDeleteState struct {
	Kind      ConfirmDeleteKind
	ProjectID string
}

func (c *ConfirmDeleteState) reset() {
	c.Kind = ConfirmDeleteSelection
	c.ProjectID = ""
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

func NewFoldStateFrom(foldedIDs []string) FoldState {
	folded := make(map[string]bool, len(foldedIDs))
	for _, id := range foldedIDs {
		folded[id] = true
	}
	return FoldState{folded: folded}
}

func (f *FoldState) FoldedIDs() []string {
	ids := make([]string, 0, len(f.folded))
	for id := range f.folded {
		ids = append(ids, id)
	}
	return ids
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
	itemType  selection.FocusKind
}

func (e *EditState) reset() {
	e.input = textinput.New()
	e.isAdding = false
	e.newItemID = ""
	e.itemType = selection.FocusProject
}

func newEditState(value string, isAdding bool, newID string, kind selection.FocusKind) EditState {
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
