package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"phasionary/internal/app/modes"
	"phasionary/internal/app/selection"
	"phasionary/internal/domain"
)

func separatorProject() domain.Project {
	return domain.Project{
		ID:   "p1",
		Name: "P",
		Categories: []domain.Category{
			{
				ID:   "c1",
				Name: "Cat A",
				Tasks: []domain.Task{
					{ID: "t1", Title: "Alpha", Status: domain.StatusTodo},
					{ID: "s1", Kind: domain.KindSeparator, Title: "mid"},
					{ID: "t2", Title: "Beta", Status: domain.StatusTodo},
				},
			},
		},
	}
}

func (m *model) selectSeparator(catIdx, taskIdx int) bool {
	return m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
		return p.Kind == selection.FocusSeparator && p.CategoryIndex == catIdx && p.TaskIndex == taskIdx
	})
}

func TestRebuildPositionsEmitsSeparator(t *testing.T) {
	m := newTestModel(t, separatorProject())
	var found bool
	for _, p := range m.ui.Selection.Positions() {
		if p.Kind == selection.FocusSeparator && p.CategoryIndex == 0 && p.TaskIndex == 1 {
			found = true
		}
	}
	assert.True(t, found, "expected a FocusSeparator position for the separator row")
}

func TestSeparatorHiddenWhileFiltered(t *testing.T) {
	m := newTestModel(t, separatorProject())
	m.ui.Filter = NewFilterState()
	m.ui.Filter.statuses = map[string]bool{domain.StatusTodo: true}
	m.rebuildPositions()

	for _, p := range m.ui.Selection.Positions() {
		if p.Kind == selection.FocusSeparator {
			t.Fatal("separator should be hidden while a filter is active")
		}
	}
}

func TestStartAddingSeparatorInsertsBelow(t *testing.T) {
	m := newTestModel(t, separatorProject())
	// Select the first task (Alpha) then insert a separator below it.
	require.True(t, m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
		return p.Kind == selection.FocusTask && p.TaskIndex == 0
	}))
	m.startAddingSeparator()

	tasks := m.project.Categories[0].Tasks
	require.Len(t, tasks, 4)
	assert.True(t, tasks[1].IsSeparator(), "new separator should sit right below Alpha")
	assert.Equal(t, "", tasks[1].Title, "new separator should be bare")

	// Mode stays normal (no inline editor opened) and selection lands on it.
	assert.True(t, m.ui.Modes.IsNormal())
	pos, ok := m.selectedPosition()
	require.True(t, ok)
	assert.Equal(t, selection.FocusSeparator, pos.Kind)
	assert.Equal(t, 1, pos.TaskIndex)
}

func TestStartAddingSeparatorBlockedWhileFiltered(t *testing.T) {
	m := newTestModel(t, separatorProject())
	m.ui.Filter = NewFilterState()
	m.ui.Filter.statuses = map[string]bool{domain.StatusTodo: true}
	m.rebuildPositions()
	require.True(t, m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
		return p.Kind == selection.FocusTask
	}))

	before := len(m.project.Categories[0].Tasks)
	m.startAddingSeparator()
	assert.Equal(t, before, len(m.project.Categories[0].Tasks), "insert must be blocked while filtered")
}

func TestDeleteSeparatorInstant(t *testing.T) {
	m := newTestModel(t, separatorProject())
	require.True(t, m.selectSeparator(0, 1))

	m.deleteSelected()

	// No confirm dialog: mode stays normal and the separator is already gone.
	assert.True(t, m.ui.Modes.IsNormal(), "separator delete must not open the confirm dialog")
	for _, tk := range m.project.Categories[0].Tasks {
		assert.False(t, tk.IsSeparator(), "separator should be deleted")
	}
	// Undo restores it.
	m.undo()
	require.Len(t, m.project.Categories[0].Tasks, 3)
	assert.True(t, m.project.Categories[0].Tasks[1].IsSeparator())
}

func TestMoveSeparatorDown(t *testing.T) {
	m := newTestModel(t, separatorProject())
	require.True(t, m.selectSeparator(0, 1))

	m.moveTaskDown()

	tasks := m.project.Categories[0].Tasks
	assert.Equal(t, "t1", tasks[0].ID)
	assert.Equal(t, "t2", tasks[1].ID)
	assert.Equal(t, "s1", tasks[2].ID, "separator should have moved down past Beta")

	// Selection follows the separator to its new index.
	pos, ok := m.selectedPosition()
	require.True(t, ok)
	assert.Equal(t, selection.FocusSeparator, pos.Kind)
	assert.Equal(t, 2, pos.TaskIndex)
}

func TestSeparatorLabelEditAndClear(t *testing.T) {
	m := newTestModel(t, separatorProject())
	require.True(t, m.selectSeparator(0, 1))

	// Edit the label.
	m.startEditing()
	require.True(t, m.ui.Modes.IsEdit())
	m.ui.Edit.input.SetValue("Milestone")
	m.finishEditing()
	assert.Equal(t, "Milestone", m.project.Categories[0].Tasks[1].Title)

	// Clearing the label back to bare is allowed (unlike task titles).
	require.True(t, m.selectSeparator(0, 1))
	m.startEditing()
	m.ui.Edit.input.SetValue("")
	m.finishEditing()
	assert.Equal(t, "", m.project.Categories[0].Tasks[1].Title)
	assert.True(t, m.project.Categories[0].Tasks[1].IsSeparator())
}

func TestSeparatorTaskActionsAreNoOps(t *testing.T) {
	m := newTestModel(t, separatorProject())
	require.True(t, m.selectSeparator(0, 1))

	// Status cycle: no change.
	m.toggleSelectedTask()
	assert.Equal(t, "", m.project.Categories[0].Tasks[1].Status)

	// Estimate picker: does not open.
	m.openEstimatePicker()
	assert.False(t, m.ui.Modes.IsEstimatePicker())
	assert.True(t, m.ui.Modes.Current() == modes.ModeNormal)

	// Visual mode: refuses to start on a separator.
	m.enterVisualMode()
	assert.False(t, m.ui.Modes.IsVisual())
}
