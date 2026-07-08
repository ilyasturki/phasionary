package app

import (
	"phasionary/internal/app/selection"
	"phasionary/internal/domain"
)

// storeTaskUpdate persists the in-memory project to disk under the project
// flock. The lock makes the WRITE atomic vs. concurrent `phasionary serve`
// activity, but the TUI's load-mutate-save cycle still wins blindly over any
// edits serve made in the meantime — use reloadProject (R) to pick up
// out-of-band changes before continuing to edit.
func (m *model) storeTaskUpdate() {
	if m.deps.Store == nil {
		return
	}
	if err := m.deps.Store.SaveProjectLocked(m.project); err != nil {
		m.ui.Screen.StatusMsg = "Save failed: " + err.Error()
	}
}

func (m *model) reloadProject() {
	if m.deps.Store == nil {
		return
	}

	var (
		prevKind   selection.FocusKind
		prevCatID  string
		prevTaskID string
		hadSel     bool
	)
	if pos, ok := m.selectedPosition(); ok {
		hadSel = true
		prevKind = pos.Kind
		if pos.Kind == selection.FocusCategory || pos.Kind == selection.FocusTask {
			if pos.CategoryIndex >= 0 && pos.CategoryIndex < len(m.project.Categories) {
				cat := m.project.Categories[pos.CategoryIndex]
				prevCatID = cat.ID
				if pos.Kind == selection.FocusTask && pos.TaskIndex >= 0 && pos.TaskIndex < len(cat.Tasks) {
					prevTaskID = cat.Tasks[pos.TaskIndex].ID
				}
			}
		}
	}

	project, err := m.deps.Store.LoadProject(m.project.ID)
	if err != nil {
		m.ui.Screen.StatusMsg = "Reload failed: " + err.Error()
		return
	}

	m.project = project
	m.ui.History.Reset()
	m.rebuildPositions()

	if hadSel {
		restored := m.restoreSelection(prevKind, prevCatID, prevTaskID)
		if !restored {
			idx := findFirstTaskIndex(m.ui.Selection.Positions())
			m.ui.Selection.SetSelected(idx)
		}
	}

	m.ensureVisible()
	m.ui.Screen.StatusMsg = "Reloaded from disk"
}

func (m *model) restoreSelection(kind selection.FocusKind, catID, taskID string) bool {
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
	case selection.FocusTask:
		if taskID != "" {
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

// taskRowVisible is the single source of truth for whether a row in a
// category's Tasks slice is shown. Separators are hidden whenever any filter is
// active (they carry no status/priority to match against); ordinary tasks defer
// to the filter's per-task rule.
func taskRowVisible(filter *FilterState, task domain.Task, categoryID string) bool {
	if task.IsSeparator() {
		return filter == nil || !filter.HasActiveFilter()
	}
	return filter == nil || filter.TaskVisible(task, categoryID)
}

func rebuildPositions(categories []domain.Category, filter *FilterState, fold *FoldState, expandDescriptions bool) []selection.Position {
	positions := make([]selection.Position, 0)
	positions = append(positions, selection.Position{
		Kind:          selection.FocusProject,
		CategoryIndex: -1,
		TaskIndex:     -1,
	})
	for cIndex, category := range categories {
		positions = append(positions, selection.Position{
			Kind:          selection.FocusCategory,
			CategoryIndex: cIndex,
			TaskIndex:     -1,
		})
		if fold != nil && fold.IsFolded(category.ID) {
			continue
		}
		for tIndex, task := range category.Tasks {
			if !taskRowVisible(filter, task, category.ID) {
				continue
			}
			if task.IsSeparator() {
				positions = append(positions, selection.Position{
					Kind:          selection.FocusSeparator,
					CategoryIndex: cIndex,
					TaskIndex:     tIndex,
				})
				continue
			}
			positions = append(positions, selection.Position{
				Kind:          selection.FocusTask,
				CategoryIndex: cIndex,
				TaskIndex:     tIndex,
			})
			if expandDescriptions && task.Description != "" {
				positions = append(positions, selection.Position{
					Kind:          selection.FocusDescription,
					CategoryIndex: cIndex,
					TaskIndex:     tIndex,
				})
			}
		}
	}
	return positions
}

func (m *model) rebuildPositions() {
	positions := rebuildPositions(m.project.Categories, &m.ui.Filter, &m.ui.Fold, m.ui.Screen.ExpandDescriptions)
	m.ui.Selection.SetPositions(positions)
}
