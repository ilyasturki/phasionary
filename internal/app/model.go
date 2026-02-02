package app

import (
	"phasionary/internal/app/modes"
	"phasionary/internal/app/selection"
	"phasionary/internal/config"
	"phasionary/internal/data"
)

type UIState struct {
	Selection    *selection.Manager
	Modes        *modes.Machine
	Edit         EditState
	Picker       ProjectPickerState
	Options      OptionsState
	Filter       FilterState
	ExternalEdit ExternalEditState
	StatusMsg    string
	ScrollOffset int
	PendingKey   rune
	Width        int
	Height       int
}

type Dependencies struct {
	Store        data.ProjectRepository
	CfgManager   *config.Manager
	StateManager *data.StateManager
}

func NewUIState(sel *selection.Manager, modeMachine *modes.Machine) *UIState {
	return &UIState{
		Selection: sel,
		Modes:     modeMachine,
		Filter:    NewFilterState(),
	}
}

func NewDependencies(store data.ProjectRepository, cfgManager *config.Manager, stateManager *data.StateManager) *Dependencies {
	return &Dependencies{
		Store:        store,
		CfgManager:   cfgManager,
		StateManager: stateManager,
	}
}
