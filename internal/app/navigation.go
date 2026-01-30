package app

import "phasionary/internal/app/modes"

func (m *model) moveSelection(delta int) {
	if !m.ui.Modes.CanPerformAction(modes.ActionNavigate) || m.ui.Selection.IsEmpty() {
		return
	}
	m.ui.Selection.MoveBy(delta)
	m.ensureVisible()
}

func (m *model) moveSelectionByPage(factor float64) {
	if !m.ui.Modes.CanPerformAction(modes.ActionNavigate) || m.ui.Selection.IsEmpty() {
		return
	}
	pageSize := int(float64(m.availableHeight()) * factor)
	if pageSize == 0 {
		if factor > 0 {
			pageSize = 1
		} else {
			pageSize = -1
		}
	}
	m.moveSelection(pageSize)
}

func (m *model) jumpToFirst() {
	if !m.ui.Modes.CanPerformAction(modes.ActionNavigate) || m.ui.Selection.IsEmpty() {
		return
	}
	m.ui.Selection.JumpToFirst()
	m.ensureVisible()
}

func (m *model) jumpToLast() {
	if !m.ui.Modes.CanPerformAction(modes.ActionNavigate) || m.ui.Selection.IsEmpty() {
		return
	}
	m.ui.Selection.JumpToLast()
	m.ensureVisible()
}
