package app

import (
	"phasionary/internal/app/components"
	"phasionary/internal/app/modes"
	"phasionary/internal/app/selection"
	"phasionary/internal/config"
	"phasionary/internal/data"
	"phasionary/internal/domain"
)

type ClipboardState struct {
	Task     *domain.Task
	IsCut    bool
	SourceID string
}

type Screen struct {
	Width         int
	Height        int
	ScrollOffset  int
	StatusMsg     string
	PendingKey    rune
	WindowFocused bool
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
	Clipboard         ClipboardState
	LastSortAscending *bool
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
	}
}

func NewDependencies(store data.ProjectRepository, cfgManager config.Reader, stateManager data.StateRepository) *Dependencies {
	return &Dependencies{
		Store:        store,
		CfgManager:   cfgManager,
		StateManager: stateManager,
	}
}
