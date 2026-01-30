package app

type UIMode int

const (
	ModeNormal UIMode = iota
	ModeEdit
	ModeHelp
	ModeConfirmDelete
)

type EditState struct {
	value      string
	cursor     int
	isAdding   bool
	newItemID  string
	itemType   focusKind
}

func (e *EditState) reset() {
	e.value = ""
	e.cursor = 0
	e.isAdding = false
	e.newItemID = ""
	e.itemType = focusProject
}

func newEditState(value string, isAdding bool, newID string, kind focusKind) EditState {
	return EditState{
		value:     value,
		cursor:    len([]rune(value)),
		isAdding:  isAdding,
		newItemID: newID,
		itemType:  kind,
	}
}
