package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"phasionary/internal/app/modes"
	"phasionary/internal/app/selection"
	"phasionary/internal/config"
	"phasionary/internal/domain"
)

type stubConfigReader struct{ cfg config.Config }

func (s *stubConfigReader) Get() config.Config                 { return s.cfg }
func (s *stubConfigReader) Update(fn func(*config.Config)) error { fn(&s.cfg); return nil }

func newTestModel(t *testing.T, project domain.Project) *model {
	t.Helper()
	positions := rebuildPositions(project.Categories, nil, nil)
	sel := selection.NewManager(positions, 0)
	mode := modes.NewMachine(modes.ModeNormal)
	ui := NewUIState(sel, mode)
	ui.Screen.Width = 80
	ui.Screen.Height = 40
	m := &model{
		project: project,
		ui:      ui,
		deps:    &Dependencies{CfgManager: &stubConfigReader{cfg: config.DefaultConfig()}},
	}
	return m
}

func sampleProject() domain.Project {
	return domain.Project{
		ID:   "p1",
		Name: "P",
		Categories: []domain.Category{
			{
				ID:   "c1",
				Name: "Cat A",
				Tasks: []domain.Task{
					{ID: "t1", Title: "Alpha", Status: domain.StatusTodo},
					{ID: "t2", Title: "Beta", Status: domain.StatusInProgress},
					{ID: "t3", Title: "Gamma", Status: domain.StatusCompleted},
				},
			},
			{
				ID:   "c2",
				Name: "Cat B",
				Tasks: []domain.Task{
					{ID: "t4", Title: "Delta", Status: domain.StatusTodo},
					{ID: "t5", Title: "Epsilon", Status: domain.StatusTodo},
				},
			},
		},
	}
}

func TestNextVisualPosition_SkipsCategoryRows(t *testing.T) {
	positions := []selection.Position{
		{Kind: selection.FocusProject},
		{Kind: selection.FocusCategory, CategoryIndex: 0},
		{Kind: selection.FocusTask, CategoryIndex: 0, TaskIndex: 0},
		{Kind: selection.FocusTask, CategoryIndex: 0, TaskIndex: 1},
		{Kind: selection.FocusCategory, CategoryIndex: 1},
		{Kind: selection.FocusTask, CategoryIndex: 1, TaskIndex: 0},
	}

	// from a task, +1 with kind=Task should jump over the category row
	got := nextVisualPosition(positions, 3, 1, selection.FocusTask)
	assert.Equal(t, 5, got, "should skip category row at index 4")

	// from a category, +1 with kind=Category should jump over tasks
	got = nextVisualPosition(positions, 1, 1, selection.FocusCategory)
	assert.Equal(t, 4, got, "should skip task rows to next category")

	// At end, stays put
	got = nextVisualPosition(positions, 5, 1, selection.FocusTask)
	assert.Equal(t, 5, got)
}

func TestEnterVisualMode_FromTaskRow(t *testing.T) {
	m := newTestModel(t, sampleProject())
	// Position 0 is project, 1 is Cat A header, 2 is t1
	m.ui.Selection.MoveTo(2)
	m.enterVisualMode()
	assert.True(t, m.ui.Modes.IsVisual())
	assert.Equal(t, "c1", m.ui.Visual.AnchorCategoryID)
	assert.Equal(t, "t1", m.ui.Visual.AnchorTaskID)
	assert.Equal(t, selection.FocusTask, m.ui.Visual.Kind)
}

func TestEnterVisualMode_RejectedOnProjectRow(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Selection.MoveTo(0)
	m.enterVisualMode()
	assert.False(t, m.ui.Modes.IsVisual())
	assert.Contains(t, m.ui.Screen.StatusMsg, "Visual mode")
}

func TestVisualMoveCursor_ExtendsRangeAcrossCategories(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Selection.MoveTo(2) // t1
	m.enterVisualMode()
	m.visualMoveCursor(3) // should land on t4 (skipping the Cat B header)

	pos, ok := m.selectedPosition()
	require.True(t, ok)
	assert.Equal(t, selection.FocusTask, pos.Kind)
	assert.Equal(t, 1, pos.CategoryIndex)
	assert.Equal(t, 0, pos.TaskIndex)

	selPositions := m.visualSelectedPositions()
	assert.Len(t, selPositions, 4, "anchor + 3 more tasks")
}

func TestVisualCut_StoresMultipleTasksOnClipboard(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Selection.MoveTo(2) // t1
	m.enterVisualMode()
	m.visualMoveCursor(1) // include t2
	m.visualCut()

	assert.False(t, m.ui.Modes.IsVisual())
	assert.True(t, m.ui.Clipboard.IsCut)
	assert.Len(t, m.ui.Clipboard.Tasks, 2)
	assert.Equal(t, []string{"t1", "t2"}, m.ui.Clipboard.TaskIDs)
}

func TestPasteMultiTasks_FlattensIntoCursorCategory(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Selection.MoveTo(2) // t1
	m.enterVisualMode()
	m.visualMoveCursor(1) // include t2
	m.visualCut()

	// Move cursor into Cat B (position 5 = Cat B header, 6 = t4)
	// After visual cut, positions still valid. Land on t4.
	m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
		return p.Kind == selection.FocusTask && p.CategoryIndex == 1 && p.TaskIndex == 0
	})

	m.pasteFromClipboard()

	// Cat A should now have just t3
	require.Len(t, m.project.Categories[0].Tasks, 1)
	assert.Equal(t, "t3", m.project.Categories[0].Tasks[0].ID)

	// Cat B should now have t4, then pasted t1, t2 (after cursor), then t5
	require.Len(t, m.project.Categories[1].Tasks, 4)
	assert.Equal(t, "Delta", m.project.Categories[1].Tasks[0].Title)
	assert.Equal(t, "Alpha", m.project.Categories[1].Tasks[1].Title)
	assert.Equal(t, "Beta", m.project.Categories[1].Tasks[2].Title)
	assert.Equal(t, "Epsilon", m.project.Categories[1].Tasks[3].Title)

	// Clipboard cleared
	assert.Empty(t, m.ui.Clipboard.Tasks)
}

func TestVisualYank_PopulatesInternalClipboardForPaste(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Selection.MoveTo(2) // t1
	m.enterVisualMode()
	m.visualMoveCursor(1) // include t2

	_ = m.visualCopyTitles() // y

	// After yank: clipboard should hold the two tasks, NOT marked as cut.
	require.Len(t, m.ui.Clipboard.Tasks, 2)
	assert.False(t, m.ui.Clipboard.IsCut)

	// Move cursor into Cat B and paste — sources should remain in Cat A.
	m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
		return p.Kind == selection.FocusTask && p.CategoryIndex == 1 && p.TaskIndex == 0
	})
	m.pasteFromClipboard()

	// Cat A unchanged (still 3 tasks t1, t2, t3)
	require.Len(t, m.project.Categories[0].Tasks, 3)
	assert.Equal(t, "t1", m.project.Categories[0].Tasks[0].ID)
	assert.Equal(t, "t2", m.project.Categories[0].Tasks[1].ID)

	// Cat B: cursor was on t4 (Delta); paste inserts AFTER cursor.
	require.Len(t, m.project.Categories[1].Tasks, 4)
	assert.Equal(t, "Delta", m.project.Categories[1].Tasks[0].Title)
	assert.Equal(t, "Alpha", m.project.Categories[1].Tasks[1].Title)
	assert.Equal(t, "Beta", m.project.Categories[1].Tasks[2].Title)
	assert.Equal(t, "Epsilon", m.project.Categories[1].Tasks[3].Title)
}

func TestVisualDelete_OpensConfirmThenRemovesAll(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Selection.MoveTo(2) // t1
	m.enterVisualMode()
	m.visualMoveCursor(2) // t1..t3
	m.visualDelete()

	// Confirmation opened with all 3 IDs captured.
	assert.True(t, m.ui.Modes.IsConfirmDelete())
	assert.Equal(t, ConfirmDeleteVisualRange, m.ui.ConfirmDelete.Kind)
	assert.Equal(t, []string{"t1", "t2", "t3"}, m.ui.ConfirmDelete.TaskIDs)

	m.confirmDeleteVisualRange()

	// Cat A is now empty; Cat B untouched.
	assert.Empty(t, m.project.Categories[0].Tasks)
	require.Len(t, m.project.Categories[1].Tasks, 2)
	assert.True(t, m.ui.Modes.IsNormal())
	assert.Contains(t, m.ui.Screen.StatusMsg, "Deleted 3")
}

func TestVisualDelete_CategoryRangeRemovesCategoriesAndTasks(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Selection.MoveTo(1) // Cat A
	m.enterVisualMode()
	m.visualMoveCursor(1) // Cat A + Cat B
	m.visualDelete()

	assert.Equal(t, []string{"c1", "c2"}, m.ui.ConfirmDelete.CategoryIDs)
	m.confirmDeleteVisualRange()

	assert.Empty(t, m.project.Categories)
	assert.Contains(t, m.ui.Screen.StatusMsg, "Deleted 2 categories")
}

func TestVisualMode_TaskAnchorDoesNotEscapeToCategories(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Selection.MoveTo(2) // t1
	m.enterVisualMode()
	// Try to go backwards into a category row — anchor on task, kind=Task
	m.visualMoveCursor(-1)
	// Should still be on t1 since the previous row is a category header (no earlier task)
	assert.Equal(t, 2, m.ui.Selection.Selected())
}

func TestVisualMode_CategoryAnchorSelectsOnlyCategories(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Selection.MoveTo(1) // Cat A header
	m.enterVisualMode()
	assert.Equal(t, selection.FocusCategory, m.ui.Visual.Kind)

	m.visualMoveCursor(1) // should jump to Cat B header
	pos, ok := m.selectedPosition()
	require.True(t, ok)
	assert.Equal(t, selection.FocusCategory, pos.Kind)
	assert.Equal(t, 1, pos.CategoryIndex)

	selPositions := m.visualSelectedPositions()
	assert.Len(t, selPositions, 2)
	for _, p := range selPositions {
		assert.Equal(t, selection.FocusCategory, p.Kind)
	}
}
