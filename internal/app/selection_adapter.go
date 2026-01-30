package app

import "phasionary/internal/app/selection"

func toSelectionPositions(positions []focusPosition) []selection.Position {
	result := make([]selection.Position, len(positions))
	for i, p := range positions {
		result[i] = selection.Position{
			Kind:          toSelectionKind(p.Kind),
			CategoryIndex: p.CategoryIndex,
			TaskIndex:     p.TaskIndex,
		}
	}
	return result
}

func toSelectionKind(k focusKind) selection.FocusKind {
	switch k {
	case focusCategory:
		return selection.FocusCategory
	case focusTask:
		return selection.FocusTask
	default:
		return selection.FocusProject
	}
}

func fromSelectionPosition(p selection.Position) focusPosition {
	return focusPosition{
		Kind:          fromSelectionKind(p.Kind),
		CategoryIndex: p.CategoryIndex,
		TaskIndex:     p.TaskIndex,
	}
}

func fromSelectionKind(k selection.FocusKind) focusKind {
	switch k {
	case selection.FocusCategory:
		return focusCategory
	case selection.FocusTask:
		return focusTask
	default:
		return focusProject
	}
}

func (m model) positions() []focusPosition {
	selPositions := m.ui.Selection.Positions()
	result := make([]focusPosition, len(selPositions))
	for i, p := range selPositions {
		result[i] = fromSelectionPosition(p)
	}
	return result
}

func (m model) selected() int {
	return m.ui.Selection.Selected()
}

func (m model) selectedPosition() (focusPosition, bool) {
	pos, ok := m.ui.Selection.SelectedPosition()
	if !ok {
		return focusPosition{}, false
	}
	return fromSelectionPosition(pos), true
}
