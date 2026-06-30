package app

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"phasionary/internal/domain"
)

func makeProjects(n int) []domain.Project {
	ps := make([]domain.Project, n)
	for i := range ps {
		ps[i] = domain.Project{ID: fmt.Sprintf("p%02d", i), Name: fmt.Sprintf("Project %02d", i)}
	}
	return ps
}

// newPickerModel builds a model in picker mode with n projects and a fixed
// screen size (width/height in columns/rows, as the app sees them).
func newPickerModel(t *testing.T, n, width, height int) *model {
	t.Helper()
	m := newTestModel(t, sampleProject())
	m.ui.Screen.Width = width
	m.ui.Screen.Height = height
	m.ui.Picker = ProjectPickerState{projects: makeProjects(n)}
	m.ui.Modes.ToProjectPicker()
	return m
}

func TestHalfPage(t *testing.T) {
	assert.Equal(t, 3, halfPage(7))
	assert.Equal(t, 5, halfPage(10))
	assert.Equal(t, 1, halfPage(2))
	assert.Equal(t, 1, halfPage(1), "never returns less than one row")
	assert.Equal(t, 1, halfPage(0))
}

func TestPickerEnsureVisible_ScrolloffAndClamp(t *testing.T) {
	p := &ProjectPickerState{projects: makeProjects(30)}
	const visible = 7

	p.selected = 0
	p.ensureVisible(visible)
	assert.Equal(t, 0, p.scrollOffset, "at the top the offset pins to 0, scrolloff notwithstanding")

	// Moving onto the last visible row engages scrolloff: the window scrolls a
	// row early so one context row trails below the cursor.
	p.selected = 6
	p.ensureVisible(visible)
	assert.Equal(t, 1, p.scrollOffset)

	// The bottom clamps the offset to len-visible so the last project sits flush.
	p.selected = 29
	p.ensureVisible(visible)
	assert.Equal(t, 23, p.scrollOffset)

	// Jumping back to the top drags the offset all the way up.
	p.selected = 0
	p.ensureVisible(visible)
	assert.Equal(t, 0, p.scrollOffset)
}

func TestPickerEnsureVisible_NewProjectLeavesScrollAlone(t *testing.T) {
	// New Project is pinned outside the window, so selecting it must not move
	// the projects scroll — just keep it clamped.
	p := &ProjectPickerState{projects: makeProjects(30), selected: 20, scrollOffset: 15}
	p.onNew = true
	p.ensureVisible(7)
	assert.Equal(t, 15, p.scrollOffset)
}

func TestPickerMoveSelection_CrossesNewProjectBoundary(t *testing.T) {
	p := &ProjectPickerState{projects: makeProjects(10)}

	// Up from the first project lands on the pinned New Project row.
	p.selected = 0
	p.moveSelection(-1, 6)
	assert.True(t, p.onNew, "k at the top selects New Project")

	// Down from New Project re-enters the project list at the top.
	p.moveSelection(1, 6)
	assert.False(t, p.onNew)
	assert.Equal(t, 0, p.selected)

	// A big jump down clamps to the last project, never past it.
	p.moveSelection(999, 6)
	assert.False(t, p.onNew)
	assert.Equal(t, 9, p.selected)
}

func TestPickerJumps(t *testing.T) {
	p := &ProjectPickerState{projects: makeProjects(30)}
	const visible = 7

	p.jumpToLast(visible)
	assert.False(t, p.onNew)
	assert.Equal(t, 29, p.selected, "G targets the last project, not New Project")
	assert.Equal(t, 23, p.scrollOffset)

	p.jumpToFirst(visible)
	assert.False(t, p.onNew)
	assert.Equal(t, 0, p.selected)
	assert.Equal(t, 0, p.scrollOffset)
}

func TestPickerVisibleCount_AdaptsToTerminalHeight(t *testing.T) {
	// Width 160 keeps the footer hints on a single row, so the budget is a
	// stable Height - chrome - 1 - reserve.
	m := newPickerModel(t, 50, 160, 30)
	require.Equal(t, 1, m.pickerHintRows(), "test assumes single-line hints")

	assert.Equal(t, 30-pickerChromeRows-1-pickerScrollReserve, m.pickerVisibleCount()) // 18

	m.ui.Screen.Height = 50
	assert.Equal(t, 50-pickerChromeRows-1-pickerScrollReserve, m.pickerVisibleCount()) // 38
}

func TestPickerVisibleCount_FitsAllWhenTall(t *testing.T) {
	m := newPickerModel(t, 5, 160, 40) // 5 projects, far below the budget
	assert.Equal(t, 5, m.pickerVisibleCount())
}

func TestPickerVisibleCount_ClampsToMinOnShort(t *testing.T) {
	m := newPickerModel(t, 50, 160, 12) // budget rounds below the floor
	assert.Equal(t, pickerMinVisible, m.pickerVisibleCount())
}

func TestPickerVisibleCount_FallbackBeforeFirstResize(t *testing.T) {
	m := newPickerModel(t, 50, 160, 0)
	assert.Equal(t, pickerFallbackVisible, m.pickerVisibleCount())

	few := newPickerModel(t, 5, 160, 0)
	assert.Equal(t, 5, few.pickerVisibleCount(), "still capped at the project count")
}

// TestPickerView_FitsWithinScreenHeight is the anti-clipping guard: across a
// range of heights, widths, and cursor positions, the rendered dialog never
// spills past the screen (which would clip the footer hints).
func TestPickerView_FitsWithinScreenHeight(t *testing.T) {
	for _, width := range []int{100, 160} {
		for height := 16; height <= 45; height++ {
			for _, pos := range []string{"top", "middle", "bottom", "new"} {
				m := newPickerModel(t, 40, width, height)
				switch pos {
				case "middle":
					m.ui.Picker.selected = 20
					m.ui.Picker.ensureVisible(m.pickerVisibleCount())
				case "bottom":
					m.ui.Picker.jumpToLast(m.pickerVisibleCount())
				case "new":
					m.ui.Picker.onNew = true
					m.ui.Picker.ensureVisible(m.pickerVisibleCount())
				}
				got := lipgloss.Height(m.projectPickerView())
				assert.LessOrEqualf(t, got, height,
					"dialog height %d exceeds screen %d (width=%d, pos=%s)",
					got, height, width, pos)
			}
		}
	}
}

func TestHandlePickerKey_PagingAndJumps(t *testing.T) {
	m := newPickerModel(t, 30, 160, 24)
	visible := m.pickerVisibleCount()
	require.Equal(t, 12, visible)
	m.ui.Picker.selected = 0

	// G jumps to the last project.
	after, _ := m.handleProjectPickerKey(tea.KeyPressMsg{Text: "G"})
	assert.Equal(t, 29, after.ui.Picker.selected)
	assert.False(t, after.ui.Picker.onNew)

	// g returns to the first project; k from there steps up onto pinned New Project.
	after, _ = after.handleProjectPickerKey(tea.KeyPressMsg{Text: "g"})
	after, _ = after.handleProjectPickerKey(tea.KeyPressMsg{Text: "k"})
	assert.True(t, after.ui.Picker.onNew)

	// Enter on New Project starts the inline add flow.
	after, _ = after.handleProjectPickerKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.True(t, after.ui.Picker.isAdding)
}

func TestHandlePickerKey_HalfPage(t *testing.T) {
	m := newPickerModel(t, 30, 160, 24)
	visible := m.pickerVisibleCount()
	m.ui.Picker.selected = 0

	after, _ := m.handleProjectPickerKey(tea.KeyPressMsg{Code: 'd', Mod: tea.ModCtrl})
	assert.Equal(t, halfPage(visible), after.ui.Picker.selected)

	before := after.ui.Picker.selected
	after, _ = after.handleProjectPickerKey(tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl})
	assert.Less(t, after.ui.Picker.selected, before)
}

func projectWithTasks(statuses ...string) domain.Project {
	tasks := make([]domain.Task, len(statuses))
	for i, s := range statuses {
		tasks[i] = domain.Task{ID: fmt.Sprintf("t%d", i), Status: s}
	}
	return domain.Project{Categories: []domain.Category{{Tasks: tasks}}}
}

func TestProjectProgressBadge(t *testing.T) {
	// Cancelled tasks drop out of both the numerator and denominator.
	badge, complete := projectProgressBadge(projectWithTasks(
		domain.StatusCompleted, domain.StatusTodo, domain.StatusInProgress, domain.StatusCancelled,
	))
	assert.Equal(t, "1/3", badge)
	assert.False(t, complete)

	// All active tasks done reads as complete with a check.
	badge, complete = projectProgressBadge(projectWithTasks(
		domain.StatusCompleted, domain.StatusCompleted, domain.StatusCancelled,
	))
	assert.Equal(t, "2/2 ✓", badge)
	assert.True(t, complete)

	// No countable tasks → no badge.
	badge, complete = projectProgressBadge(projectWithTasks(domain.StatusCancelled))
	assert.Equal(t, "", badge)
	assert.False(t, complete)

	badge, _ = projectProgressBadge(domain.Project{})
	assert.Equal(t, "", badge)
}
