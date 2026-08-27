package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"phasionary/internal/domain"
)

func TestParsePastedTaskLines_StripsMarkersAndReadsCheckboxes(t *testing.T) {
	got := parsePastedTaskLines("- foo\r\n* bar\n1. baz\n2) qux\n- [x] done\n- [ ] open\n\n  plain  \n• dot")
	want := []pastedTaskLine{
		{title: "foo", status: domain.StatusTodo},
		{title: "bar", status: domain.StatusTodo},
		{title: "baz", status: domain.StatusTodo},
		{title: "qux", status: domain.StatusTodo},
		{title: "done", status: domain.StatusCompleted},
		{title: "open", status: domain.StatusTodo},
		{title: "plain", status: domain.StatusTodo},
		{title: "dot", status: domain.StatusTodo},
	}
	assert.Equal(t, want, got)
}

// startAddTaskAt lands the cursor on (catIdx, taskIdx) and opens the add flow,
// so the pending blank task sits at taskIdx+1.
func startAddTaskAt(t *testing.T, m *model, catIdx, taskIdx int) {
	t.Helper()
	require.True(t, m.selectTaskIdx(catIdx, taskIdx))
	m.startAddingTask()
	require.True(t, m.ui.Modes.IsEdit())
}

func TestPasteLinesWhileAdding_CreatesOneTaskPerLine(t *testing.T) {
	m := newTestModel(t, sampleProject())
	startAddTaskAt(t, m, 0, 0) // pending task at index 1

	handled := m.pasteLinesWhileAdding("- [x] First\nSecond\n")

	require.True(t, handled)
	tasks := m.project.Categories[0].Tasks
	require.Len(t, tasks, 5)
	assert.Equal(t, "First", tasks[1].Title)
	assert.Equal(t, domain.StatusCompleted, tasks[1].Status)
	assert.Equal(t, "Second", tasks[2].Title)
	assert.Equal(t, domain.StatusTodo, tasks[2].Status)
	assert.Equal(t, "t2", tasks[3].ID, "existing tasks shift below the batch")

	assert.True(t, m.ui.Modes.IsNormal())
	pos, ok := m.selectedPosition()
	require.True(t, ok)
	assert.Equal(t, "Second", m.project.Categories[0].Tasks[pos.TaskIndex].Title)
	assert.Equal(t, "Added 2 tasks", m.ui.Screen.StatusMsg)
}

func TestPasteLinesWhileAdding_PrependsTypedText(t *testing.T) {
	m := newTestModel(t, sampleProject())
	startAddTaskAt(t, m, 0, 0)
	m.ui.Edit.input.SetValue("pre")

	require.True(t, m.pasteLinesWhileAdding("fix\nnext"))

	tasks := m.project.Categories[0].Tasks
	assert.Equal(t, "pre fix", tasks[1].Title)
	assert.Equal(t, "next", tasks[2].Title)
}

func TestPasteLinesWhileAdding_SingleLineFallsThrough(t *testing.T) {
	m := newTestModel(t, sampleProject())
	startAddTaskAt(t, m, 0, 0)

	assert.False(t, m.pasteLinesWhileAdding("just one line"))
	assert.False(t, m.pasteLinesWhileAdding("trailing newline\n"))
	assert.True(t, m.ui.Modes.IsEdit(), "single-line pastes stay in the inline editor")
}

func TestPasteLinesWhileAdding_NotWhileEditingExisting(t *testing.T) {
	m := newTestModel(t, sampleProject())
	require.True(t, m.selectTaskIdx(0, 0))
	m.startEditing() // editing t1, not adding

	assert.False(t, m.pasteLinesWhileAdding("a\nb"))
	assert.Len(t, m.project.Categories[0].Tasks, 3)
}

func TestPasteLinesWhileAdding_UndoesAsOneBatch(t *testing.T) {
	m := newTestModel(t, sampleProject())
	startAddTaskAt(t, m, 0, 0)
	require.True(t, m.pasteLinesWhileAdding("a\nb\nc"))
	require.Len(t, m.project.Categories[0].Tasks, 6)

	m.undo()

	assert.Len(t, m.project.Categories[0].Tasks, 3)
	for _, task := range m.project.Categories[0].Tasks {
		assert.NotEmpty(t, task.ID)
	}
	// Guard: no leftover selection on a vanished row.
	_, ok := m.selectedPosition()
	assert.True(t, ok)
}
