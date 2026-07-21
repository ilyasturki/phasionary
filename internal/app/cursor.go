package app

import (
	"phasionary/internal/app/selection"
	"phasionary/internal/data"
)

// cursorKindNames maps focus kinds to the strings stored in state.json. The
// names are part of the on-disk format: renaming one silently drops every saved
// cursor of that kind, which degrades to "start on the first task".
var cursorKindNames = map[selection.FocusKind]string{
	selection.FocusProject:     "project",
	selection.FocusCategory:    "category",
	selection.FocusTask:        "task",
	selection.FocusSeparator:   "separator",
	selection.FocusDescription: "description",
}

func cursorKindFromName(name string) (selection.FocusKind, bool) {
	for kind, n := range cursorKindNames {
		if n == name {
			return kind, true
		}
	}
	return 0, false
}

// currentCursor describes the focused row by stable IDs so it can be resolved
// again against a project that changed in the meantime. Reports false when
// nothing is selected, which is the case for an empty project.
func (m model) currentCursor() (data.Cursor, bool) {
	pos, ok := m.selectedPosition()
	if !ok {
		return data.Cursor{}, false
	}
	name, ok := cursorKindNames[pos.Kind]
	if !ok {
		return data.Cursor{}, false
	}

	cursor := data.Cursor{Kind: name}
	if pos.Kind == selection.FocusProject {
		return cursor, true
	}
	if pos.CategoryIndex < 0 || pos.CategoryIndex >= len(m.project.Categories) {
		return data.Cursor{}, false
	}

	cat := m.project.Categories[pos.CategoryIndex]
	cursor.CategoryID = cat.ID
	// Separators and descriptions share the task addressing: both hang off an
	// entry in the category's Tasks slice. An out-of-range index leaves TaskID
	// empty, which resolves back to the category header rather than nothing.
	if pos.Kind != selection.FocusCategory && pos.TaskIndex >= 0 && pos.TaskIndex < len(cat.Tasks) {
		cursor.TaskID = cat.Tasks[pos.TaskIndex].ID
	}
	return cursor, true
}

// saveCursorState persists the focused row for the current project. Called when
// the project changes and once on exit — not on every move, which would rewrite
// state.json on every keystroke (the file is shared with `phasionary serve`, and
// each write reloads and rewrites it whole).
func (m *model) saveCursorState() {
	if m.deps.StateManager == nil || m.project.ID == "" {
		return
	}
	// A project with nothing selected stores the zero cursor, clearing any entry
	// left over from when it still had rows.
	cursor, _ := m.currentCursor()
	_ = m.deps.StateManager.SetCursor(m.project.ID, cursor)
}

// restoreCursor moves the selection onto a previously saved row, reporting
// whether it landed.
func (m *model) restoreCursor(cursor data.Cursor) bool {
	if cursor.IsZero() {
		return false
	}
	kind, ok := cursorKindFromName(cursor.Kind)
	if !ok {
		return false
	}
	return m.restoreSelection(kind, cursor.CategoryID, cursor.TaskID)
}

// applyStoredCursor puts the cursor where this project was last left, falling
// back to the first task when nothing is stored or the remembered row is
// unreachable — deleted elsewhere, or hidden behind a fold or filter. Call after
// rebuildPositions, since it selects over the current position list.
func (m *model) applyStoredCursor() {
	if m.deps.StateManager != nil && m.restoreCursor(m.deps.StateManager.GetCursor(m.project.ID)) {
		return
	}
	m.ui.Selection.SetSelected(findFirstTaskIndex(m.ui.Selection.Positions()))
}
