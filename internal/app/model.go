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
	StatusMsg    string
	ScrollOffset int
	PendingKey   rune
	Width        int
	Height       int
}

type Dependencies struct {
	Store      data.ProjectRepository
	CfgManager *config.Manager
}

func NewUIState(sel *selection.Manager, modeMachine *modes.Machine) *UIState {
	return &UIState{
		Selection: sel,
		Modes:     modeMachine,
	}
}

func NewDependencies(store data.ProjectRepository, cfgManager *config.Manager) *Dependencies {
	return &Dependencies{
		Store:      store,
		CfgManager: cfgManager,
	}
}
