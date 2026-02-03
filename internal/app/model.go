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

type UIState struct {
	Selection          *selection.Manager
	Modes              *modes.Machine
	Edit               EditState
	Picker             ProjectPickerState
	Options            OptionsState
	Filter             FilterState
	ExternalEdit       ExternalEditState
	EstimatePicker     components.EstimatePickerState
	Clipboard          ClipboardState
	StatusMsg          string
	ScrollOffset       int
	PendingKey         rune
	Width              int
	Height             int
	LastSortAscending  *bool
	WindowFocused      bool
}

type Dependencies struct {
	Store        data.ProjectRepository
	CfgManager   *config.Manager
	StateManager *data.StateManager
}

func NewUIState(sel *selection.Manager, modeMachine *modes.Machine) *UIState {
	return &UIState{
		Selection:     sel,
		Modes:         modeMachine,
		Filter:        NewFilterState(),
		WindowFocused: true,
	}
}

func NewDependencies(store data.ProjectRepository, cfgManager *config.Manager, stateManager *data.StateManager) *Dependencies {
	return &Dependencies{
		Store:        store,
		CfgManager:   cfgManager,
		StateManager: stateManager,
	}
}
