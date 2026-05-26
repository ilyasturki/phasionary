package app

import (
	"phasionary/internal/app/selection"
	"phasionary/internal/domain"
)

const historyLimit = 100

type historyEntry struct {
	project      domain.Project
	hadSelection bool
	selKind      selection.FocusKind
	selCatID     string
	selTaskID    string
}

type HistoryState struct {
	undo []historyEntry
	redo []historyEntry
}

func NewHistoryState() HistoryState {
	return HistoryState{}
}

func (h *HistoryState) CanUndo() bool { return len(h.undo) > 0 }
func (h *HistoryState) CanRedo() bool { return len(h.redo) > 0 }

func (h *HistoryState) UndoDepth() int { return len(h.undo) }
func (h *HistoryState) RedoDepth() int { return len(h.redo) }

func (h *HistoryState) Reset() {
	h.undo = nil
	h.redo = nil
}

func cloneProject(p domain.Project) domain.Project {
	out := p
	out.Categories = make([]domain.Category, len(p.Categories))
	for i, c := range p.Categories {
		nc := c
		nc.Tasks = make([]domain.Task, len(c.Tasks))
		copy(nc.Tasks, c.Tasks)
		out.Categories[i] = nc
	}
	return out
}

func (m *model) snapshotEntry() historyEntry {
	entry := historyEntry{project: cloneProject(m.project)}
	if pos, ok := m.selectedPosition(); ok {
		entry.hadSelection = true
		entry.selKind = pos.Kind
		if pos.CategoryIndex >= 0 && pos.CategoryIndex < len(m.project.Categories) {
			cat := m.project.Categories[pos.CategoryIndex]
			entry.selCatID = cat.ID
			if (pos.Kind == selection.FocusTask || pos.Kind == selection.FocusDescription) &&
				pos.TaskIndex >= 0 && pos.TaskIndex < len(cat.Tasks) {
				entry.selTaskID = cat.Tasks[pos.TaskIndex].ID
			}
		}
	}
	return entry
}

// recordHistory captures the current state on the undo stack and clears redo.
// Call BEFORE any project mutation.
func (m *model) recordHistory() {
	entry := m.snapshotEntry()
	m.ui.History.undo = append(m.ui.History.undo, entry)
	if len(m.ui.History.undo) > historyLimit {
		m.ui.History.undo = m.ui.History.undo[len(m.ui.History.undo)-historyLimit:]
	}
	m.ui.History.redo = nil
}

// discardLastHistory removes the most-recently recorded undo entry. Used when a
// recorded mutation was cancelled or turned out to be a no-op.
func (m *model) discardLastHistory() {
	if n := len(m.ui.History.undo); n > 0 {
		m.ui.History.undo = m.ui.History.undo[:n-1]
	}
}

func (m *model) undo() {
	if !m.ui.History.CanUndo() {
		m.ui.Screen.StatusMsg = "Nothing to undo"
		return
	}
	current := m.snapshotEntry()
	n := len(m.ui.History.undo) - 1
	entry := m.ui.History.undo[n]
	m.ui.History.undo = m.ui.History.undo[:n]
	m.ui.History.redo = append(m.ui.History.redo, current)
	m.applyHistoryEntry(entry)
	m.ui.Screen.StatusMsg = "Undid change"
}

func (m *model) redo() {
	if !m.ui.History.CanRedo() {
		m.ui.Screen.StatusMsg = "Nothing to redo"
		return
	}
	current := m.snapshotEntry()
	n := len(m.ui.History.redo) - 1
	entry := m.ui.History.redo[n]
	m.ui.History.redo = m.ui.History.redo[:n]
	m.ui.History.undo = append(m.ui.History.undo, current)
	m.applyHistoryEntry(entry)
	m.ui.Screen.StatusMsg = "Redid change"
}

func (m *model) applyHistoryEntry(e historyEntry) {
	m.project = cloneProject(e.project)
	m.rebuildPositions()
	if e.hadSelection {
		if !m.historyRestoreSelection(e.selKind, e.selCatID, e.selTaskID) {
			idx := findFirstTaskIndex(m.ui.Selection.Positions())
			m.ui.Selection.SetSelected(idx)
		}
	}
	m.ensureVisible()
	m.storeTaskUpdate()
}

func (m *model) historyRestoreSelection(kind selection.FocusKind, catID, taskID string) bool {
	categories := m.project.Categories
	switch kind {
	case selection.FocusProject:
		return m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
			return p.Kind == selection.FocusProject
		})
	case selection.FocusCategory:
		if catID == "" {
			return false
		}
		return m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
			return p.Kind == selection.FocusCategory &&
				p.CategoryIndex >= 0 && p.CategoryIndex < len(categories) &&
				categories[p.CategoryIndex].ID == catID
		})
	case selection.FocusTask, selection.FocusDescription:
		if taskID != "" {
			if m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
				if p.Kind != kind {
					return false
				}
				if p.CategoryIndex < 0 || p.CategoryIndex >= len(categories) {
					return false
				}
				cat := categories[p.CategoryIndex]
				return p.TaskIndex >= 0 && p.TaskIndex < len(cat.Tasks) && cat.Tasks[p.TaskIndex].ID == taskID
			}) {
				return true
			}
			// Fallback: if the description row is gone (e.g. desc cleared), try the task.
			if kind == selection.FocusDescription {
				if m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
					if p.Kind != selection.FocusTask {
						return false
					}
					if p.CategoryIndex < 0 || p.CategoryIndex >= len(categories) {
						return false
					}
					cat := categories[p.CategoryIndex]
					return p.TaskIndex >= 0 && p.TaskIndex < len(cat.Tasks) && cat.Tasks[p.TaskIndex].ID == taskID
				}) {
					return true
				}
			}
		}
		if catID != "" {
			return m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
				return p.Kind == selection.FocusCategory &&
					p.CategoryIndex >= 0 && p.CategoryIndex < len(categories) &&
					categories[p.CategoryIndex].ID == catID
			})
		}
	}
	return false
}
