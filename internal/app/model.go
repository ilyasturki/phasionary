package app

import (
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"

	"phasionary/internal/app/components"
	"phasionary/internal/app/modes"
	"phasionary/internal/app/selection"
	"phasionary/internal/config"
	"phasionary/internal/data"
	"phasionary/internal/domain"
)

type DescriptionEditState struct {
	textarea      textarea.Model
	categoryIndex int
	taskIndex     int
	original      string
	creating      bool // true when the task had no description before this edit
}

type ClipboardState struct {
	Task     *domain.Task
	IsCut    bool
	SourceID string

	Tasks       []domain.Task
	TaskIDs     []string
	Categories  []domain.Category
	CategoryIDs []string
}

type VisualState struct {
	Active           bool
	Kind             selection.FocusKind
	AnchorCategoryID string
	AnchorTaskID     string
}

// TagClipboard holds a copied tag for the "tag painter": gt grabs the focused
// task's tag, p (or p in visual mode) applies it elsewhere. Set distinguishes
// "nothing copied yet" from a copied empty tag (which pastes as a clear).
type TagClipboard struct {
	Set   bool
	Color string
	Label string
}

type Screen struct {
	Width              int
	Height             int
	ScrollOffset       int
	StatusMsg          string
	PendingKey         rune
	WindowFocused      bool
	ExpandDescriptions bool
}

type HelpState struct {
	Focused      int
	ScrollOffset int
	// Filtering is true while the `/` incremental filter is active; Filter holds
	// its query. The filter narrows the shortcut list, matching each row against
	// both its key and its label.
	Filtering bool
	Filter    textinput.Model
}

type UIState struct {
	Selection      *selection.Manager
	Modes          *modes.Machine
	Screen         Screen
	Edit           EditState
	Picker         ProjectPickerState
	ConfirmDelete  ConfirmDeleteState
	Options        OptionsState
	Filter         FilterState
	Fold           FoldState
	ExternalEdit   ExternalEditState
	EstimatePicker components.EstimatePickerState
	URLPicker      components.URLPickerState
	YankPicker     components.YankPickerState
	Help           HelpState
	Clipboard      ClipboardState
	TagClip        TagClipboard
	// TagCopiedLast records whether the most recent copy gesture was a tag copy
	// (gt) rather than a task/category cut or copy. It routes the normal-mode `p`:
	// true → paint the copied tag, false → paste the clipboard task/category.
	TagCopiedLast   bool
	Visual          VisualState
	DescriptionEdit DescriptionEditState
	TagEdit         TagEditState
	History         HistoryState
	Search          SearchState
}

type Dependencies struct {
	Store        data.ProjectRepository
	CfgManager   config.Reader
	StateManager data.StateRepository
}

func NewUIState(sel *selection.Manager, modeMachine *modes.Machine) *UIState {
	return &UIState{
		Selection: sel,
		Modes:     modeMachine,
		Screen:    Screen{WindowFocused: true},
		Filter:    NewFilterState(),
		Fold:      NewFoldState(),
		History:   NewHistoryState(),
	}
}

func NewDependencies(store data.ProjectRepository, cfgManager config.Reader, stateManager data.StateRepository) *Dependencies {
	return &Dependencies{
		Store:        store,
		CfgManager:   cfgManager,
		StateManager: stateManager,
	}
}
