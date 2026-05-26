package app

import (
	"charm.land/bubbles/v2/textarea"

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
}

type UIState struct {
	Selection         *selection.Manager
	Modes             *modes.Machine
	Screen            Screen
	Edit              EditState
	Picker            ProjectPickerState
	ConfirmDelete     ConfirmDeleteState
	Options           OptionsState
	Filter            FilterState
	Fold              FoldState
	ExternalEdit      ExternalEditState
	EstimatePicker    components.EstimatePickerState
	URLPicker         components.URLPickerState
	Help              HelpState
	Clipboard         ClipboardState
	Visual            VisualState
	DescriptionEdit   DescriptionEditState
	History           HistoryState
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
