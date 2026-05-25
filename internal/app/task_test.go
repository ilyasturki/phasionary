package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"phasionary/internal/app/selection"
	"phasionary/internal/domain"
)

func TestToggleSelectedTask_CyclesStatus(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Selection.MoveTo(2) // t1 (Todo)
	m.toggleSelectedTask()
	assert.Equal(t, domain.StatusInProgress, m.project.Categories[0].Tasks[0].Status)
	m.toggleSelectedTask()
	assert.Equal(t, domain.StatusCompleted, m.project.Categories[0].Tasks[0].Status)
}

func TestToggleSelectedTask_NoopOnCategory(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Selection.MoveTo(1) // Cat A header
	// Should not panic / mutate any task
	m.toggleSelectedTask()
	assert.Equal(t, domain.StatusTodo, m.project.Categories[0].Tasks[0].Status)
}

func TestIncreasePriority_FromMediumToHigh(t *testing.T) {
	p := sampleProject()
	p.Categories[0].Tasks[0].Priority = domain.PriorityMedium
	m := newTestModel(t, p)
	m.ui.Selection.MoveTo(2)
	m.increasePriority()
	assert.Equal(t, domain.PriorityHigh, m.project.Categories[0].Tasks[0].Priority)
}

func TestDecreasePriority_FromMediumToLow(t *testing.T) {
	p := sampleProject()
	p.Categories[0].Tasks[0].Priority = domain.PriorityMedium
	m := newTestModel(t, p)
	m.ui.Selection.MoveTo(2)
	m.decreasePriority()
	assert.Equal(t, domain.PriorityLow, m.project.Categories[0].Tasks[0].Priority)
}

func TestIncreasePriority_NoopOnCategory(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Selection.MoveTo(1) // Cat A header
	originalPriority := m.project.Categories[0].Tasks[0].Priority
	m.increasePriority()
	assert.Equal(t, originalPriority, m.project.Categories[0].Tasks[0].Priority)
}

func TestMoveTaskDown_SwapsWithinCategory(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Selection.MoveTo(2) // t1
	m.moveTaskDown()
	assert.Equal(t, "t2", m.project.Categories[0].Tasks[0].ID)
	assert.Equal(t, "t1", m.project.Categories[0].Tasks[1].ID)
	// Selection follows the moved task
	pos, ok := m.selectedPosition()
	require.True(t, ok)
	assert.Equal(t, selection.FocusTask, pos.Kind)
	assert.Equal(t, 0, pos.CategoryIndex)
	assert.Equal(t, 1, pos.TaskIndex)
}

func TestMoveTaskDown_CrossesIntoNextCategory(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Selection.MoveTo(4) // t3 (last task in Cat A)
	m.moveTaskDown()
	// t3 should be inserted at the head of Cat B
	assert.Len(t, m.project.Categories[0].Tasks, 2)
	require.Len(t, m.project.Categories[1].Tasks, 3)
	assert.Equal(t, "t3", m.project.Categories[1].Tasks[0].ID)
	pos, ok := m.selectedPosition()
	require.True(t, ok)
	assert.Equal(t, 1, pos.CategoryIndex)
	assert.Equal(t, 0, pos.TaskIndex)
}

func TestMoveTaskDown_NoopAtVeryEnd(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Selection.MoveTo(7) // t5 (last task in last category)
	m.moveTaskDown()
	assert.Equal(t, "t5", m.project.Categories[1].Tasks[1].ID)
}

func TestMoveTaskUp_SwapsWithinCategory(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Selection.MoveTo(3) // t2
	m.moveTaskUp()
	assert.Equal(t, "t2", m.project.Categories[0].Tasks[0].ID)
	assert.Equal(t, "t1", m.project.Categories[0].Tasks[1].ID)
}

func TestMoveTaskUp_CrossesIntoPrevCategory(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Selection.MoveTo(6) // t4 (first task in Cat B)
	m.moveTaskUp()
	// t4 should be appended to Cat A
	require.Len(t, m.project.Categories[0].Tasks, 4)
	assert.Equal(t, "t4", m.project.Categories[0].Tasks[3].ID)
	assert.Len(t, m.project.Categories[1].Tasks, 1)
}

func TestMoveTaskUp_NoopAtVeryStart(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Selection.MoveTo(2) // t1 (first task in first category)
	m.moveTaskUp()
	assert.Equal(t, "t1", m.project.Categories[0].Tasks[0].ID)
}

func TestMoveCategoryDown_SwapsCategories(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Selection.MoveTo(1) // Cat A
	m.moveCategoryDown()
	assert.Equal(t, "c2", m.project.Categories[0].ID)
	assert.Equal(t, "c1", m.project.Categories[1].ID)
	// Selection follows the moved category
	pos, ok := m.selectedPosition()
	require.True(t, ok)
	assert.Equal(t, selection.FocusCategory, pos.Kind)
	assert.Equal(t, 1, pos.CategoryIndex)
}

func TestMoveCategoryDown_NoopAtEnd(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Selection.MoveTo(5) // Cat B (last)
	m.moveCategoryDown()
	assert.Equal(t, "c1", m.project.Categories[0].ID)
	assert.Equal(t, "c2", m.project.Categories[1].ID)
}

func TestMoveCategoryUp_SwapsCategories(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Selection.MoveTo(5) // Cat B
	m.moveCategoryUp()
	assert.Equal(t, "c2", m.project.Categories[0].ID)
	assert.Equal(t, "c1", m.project.Categories[1].ID)
}

func TestMoveCategoryUp_NoopAtStart(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Selection.MoveTo(1) // Cat A
	m.moveCategoryUp()
	assert.Equal(t, "c1", m.project.Categories[0].ID)
}

func TestCutSelectedTask_PopulatesClipboardAndSetsStatus(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Selection.MoveTo(2) // t1
	m.cutSelectedTask()
	require.NotNil(t, m.ui.Clipboard.Task)
	assert.True(t, m.ui.Clipboard.IsCut)
	assert.Equal(t, "t1", m.ui.Clipboard.SourceID)
	assert.Contains(t, m.ui.Screen.StatusMsg, "Marked for cut")
}

func TestCutSelectedTask_RejectsCategory(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Selection.MoveTo(1) // Cat A header
	m.cutSelectedTask()
	assert.Nil(t, m.ui.Clipboard.Task)
	assert.Equal(t, "Can only cut tasks", m.ui.Screen.StatusMsg)
}

func TestPasteTask_AfterCut_MovesTaskToTarget(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Selection.MoveTo(2) // t1
	m.cutSelectedTask()
	// Move to t4 in Cat B
	m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
		return p.Kind == selection.FocusTask && p.CategoryIndex == 1 && p.TaskIndex == 0
	})
	m.pasteFromClipboard()

	// Cat A lost t1
	assert.Len(t, m.project.Categories[0].Tasks, 2)
	assert.Equal(t, "t2", m.project.Categories[0].Tasks[0].ID)
	// Cat B: cursor was on t4 (Delta); paste lands AFTER it.
	require.Len(t, m.project.Categories[1].Tasks, 3)
	assert.Equal(t, "Delta", m.project.Categories[1].Tasks[0].Title)
	assert.Equal(t, "Alpha", m.project.Categories[1].Tasks[1].Title)
	// Clipboard cleared, status updated
	assert.Nil(t, m.ui.Clipboard.Task)
	assert.Equal(t, "Moved!", m.ui.Screen.StatusMsg)
}

func TestPasteTask_AfterCopy_KeepsOriginal(t *testing.T) {
	m := newTestModel(t, sampleProject())
	// Simulate a copied (not cut) task via clipboard.
	src := m.project.Categories[0].Tasks[0]
	m.ui.Clipboard = ClipboardState{Task: &src, IsCut: false}
	// Paste onto t4 in Cat B
	m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
		return p.Kind == selection.FocusTask && p.CategoryIndex == 1 && p.TaskIndex == 0
	})
	m.pasteFromClipboard()

	// Original survived
	assert.Equal(t, "t1", m.project.Categories[0].Tasks[0].ID)
	// Cat B got a clone AFTER cursor (cursor was on t4 / Delta at index 0)
	require.Len(t, m.project.Categories[1].Tasks, 3)
	assert.Equal(t, "Delta", m.project.Categories[1].Tasks[0].Title)
	clone := m.project.Categories[1].Tasks[1]
	assert.NotEqual(t, "t1", clone.ID)
	assert.Equal(t, "Alpha", clone.Title)
	assert.Equal(t, "Pasted!", m.ui.Screen.StatusMsg)
}

func TestPasteTask_EmptyClipboardSetsStatus(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Selection.MoveTo(2)
	m.pasteFromClipboard()
	assert.Equal(t, "Nothing to paste", m.ui.Screen.StatusMsg)
}

func TestRemoveTaskByID(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.removeTaskByID("t2")
	assert.Len(t, m.project.Categories[0].Tasks, 2)
	for _, task := range m.project.Categories[0].Tasks {
		assert.NotEqual(t, "t2", task.ID)
	}
}

func TestRemoveTaskByID_NonexistentIsNoop(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.removeTaskByID("does-not-exist")
	assert.Len(t, m.project.Categories[0].Tasks, 3)
	assert.Len(t, m.project.Categories[1].Tasks, 2)
}

func TestDeleteTask_RemovesAndCopiesToClipboard(t *testing.T) {
	m := newTestModel(t, sampleProject())
	pos := selection.Position{Kind: selection.FocusTask, CategoryIndex: 0, TaskIndex: 0}
	m.ui.Selection.MoveTo(2)
	m.deleteTask(pos)
	assert.Len(t, m.project.Categories[0].Tasks, 2)
	require.NotNil(t, m.ui.Clipboard.Task)
	assert.Equal(t, "Alpha", m.ui.Clipboard.Task.Title)
	assert.False(t, m.ui.Clipboard.IsCut)
}

func TestDeleteCategory_RemovesEntireCategory(t *testing.T) {
	m := newTestModel(t, sampleProject())
	pos := selection.Position{Kind: selection.FocusCategory, CategoryIndex: 0}
	m.ui.Selection.MoveTo(1)
	m.deleteCategory(pos)
	require.Len(t, m.project.Categories, 1)
	assert.Equal(t, "c2", m.project.Categories[0].ID)
}
