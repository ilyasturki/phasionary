package modes

type Mode int

const (
	ModeNormal Mode = iota
	ModeEdit
	ModeHelp
	ModeConfirmDelete
	ModeOptions
	ModeProjectPicker
	ModeFilter
	ModeExternalEdit
	ModeInfo
)

type Action int

const (
	ActionNavigate Action = iota
	ActionToggleTask
	ActionDeleteItem
	ActionEditItem
	ActionAddTask
	ActionAddCategory
	ActionChangePriority
	ActionMoveItem
	ActionSort
	ActionCopy
	ActionOpenPicker
	ActionOpenOptions
	ActionOpenHelp
)

type Machine struct {
	current Mode
}

func NewMachine(initial Mode) *Machine {
	return &Machine{current: initial}
}

func (m *Machine) Current() Mode {
	return m.current
}

func (m *Machine) IsNormal() bool {
	return m.current == ModeNormal
}

func (m *Machine) IsEdit() bool {
	return m.current == ModeEdit
}

func (m *Machine) IsHelp() bool {
	return m.current == ModeHelp
}

func (m *Machine) IsConfirmDelete() bool {
	return m.current == ModeConfirmDelete
}

func (m *Machine) IsOptions() bool {
	return m.current == ModeOptions
}

func (m *Machine) IsProjectPicker() bool {
	return m.current == ModeProjectPicker
}

func (m *Machine) IsFilter() bool {
	return m.current == ModeFilter
}

func (m *Machine) IsExternalEdit() bool {
	return m.current == ModeExternalEdit
}

func (m *Machine) IsInfo() bool {
	return m.current == ModeInfo
}

func (m *Machine) TransitionTo(mode Mode) bool {
	if !m.canTransition(mode) {
		return false
	}
	m.current = mode
	return true
}

func (m *Machine) canTransition(target Mode) bool {
	switch m.current {
	case ModeNormal:
		return true
	case ModeEdit:
		return target == ModeNormal
	case ModeHelp:
		return target == ModeNormal
	case ModeConfirmDelete:
		return target == ModeNormal
	case ModeOptions:
		return target == ModeNormal
	case ModeProjectPicker:
		return target == ModeNormal
	case ModeFilter:
		return target == ModeNormal
	case ModeExternalEdit:
		return target == ModeNormal
	case ModeInfo:
		return target == ModeNormal
	}
	return false
}

func (m *Machine) CanPerformAction(action Action) bool {
	switch m.current {
	case ModeNormal:
		return true
	case ModeEdit:
		return false
	case ModeHelp:
		return action == ActionOpenHelp
	case ModeConfirmDelete:
		return false
	case ModeOptions:
		return false
	case ModeProjectPicker:
		return false
	case ModeFilter:
		return false
	}
	return false
}

func (m *Machine) ToNormal() {
	m.current = ModeNormal
}

func (m *Machine) ToEdit() bool {
	return m.TransitionTo(ModeEdit)
}

func (m *Machine) ToHelp() bool {
	return m.TransitionTo(ModeHelp)
}

func (m *Machine) ToConfirmDelete() bool {
	return m.TransitionTo(ModeConfirmDelete)
}

func (m *Machine) ToOptions() bool {
	return m.TransitionTo(ModeOptions)
}

func (m *Machine) ToProjectPicker() bool {
	return m.TransitionTo(ModeProjectPicker)
}

func (m *Machine) ToFilter() bool {
	return m.TransitionTo(ModeFilter)
}

func (m *Machine) ToExternalEdit() bool {
	return m.TransitionTo(ModeExternalEdit)
}

func (m *Machine) ToggleHelp() {
	if m.current == ModeHelp {
		m.current = ModeNormal
	} else if m.current == ModeNormal {
		m.current = ModeHelp
	}
}

func (m *Machine) ToInfo() bool {
	return m.TransitionTo(ModeInfo)
}

func (m *Machine) ToggleInfo() {
	if m.current == ModeInfo {
		m.current = ModeNormal
	} else if m.current == ModeNormal {
		m.current = ModeInfo
	}
}
