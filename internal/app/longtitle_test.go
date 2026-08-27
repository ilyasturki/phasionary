package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"phasionary/internal/app/modes"
	"phasionary/internal/app/selection"
	"phasionary/internal/domain"
	"phasionary/internal/ui"
)

// Titles carry no length limit. These tests pin the three places that used to
// depend on one: the list row, the inline editor, and the info dialog. Each is
// bounded on its own now, so a title of any length keeps the screen usable.

// longWords builds a space-separated string of n runes ending in a sentinel, so
// a test can tell whether a render reached the end of the text.
func longWords(n int) string {
	var b strings.Builder
	for i := 0; b.Len() < n; i++ {
		b.WriteString("word ")
	}
	return b.String()[:n-4] + "ZEND"
}

func longTitleProject(title string) domain.Project {
	return domain.Project{
		ID:   "p1",
		Name: "Probe",
		Categories: []domain.Category{{
			ID:   "c1",
			Name: "Cat",
			Tasks: []domain.Task{
				{ID: "t1", Title: "before", Status: domain.StatusTodo},
				{ID: "t2", Title: title, Status: domain.StatusTodo},
				{ID: "t3", Title: "after", Status: domain.StatusTodo},
			},
		}},
	}
}

// longTitleItem returns the layout item for the long-titled task, and its
// position index.
func longTitleItem(t *testing.T, m *model) (LayoutItem, int) {
	t.Helper()
	for _, item := range m.buildLayout().Items {
		if item.Kind == LayoutTask && item.TaskIndex == 1 {
			return item, item.PositionIndex
		}
	}
	t.Fatal("no layout item for the long-titled task")
	return LayoutItem{}, -1
}

func TestLongTitle_RowStaysBounded(t *testing.T) {
	for _, size := range []int{200, 1024, 100_000} {
		for _, dim := range []struct{ w, h int }{{40, 24}, {80, 24}, {120, 40}} {
			m := newTestModel(t, longTitleProject(longWords(size)))
			m.ui.Screen.Width, m.ui.Screen.Height = dim.w, dim.h
			m.invalidateLayout()

			item, _ := longTitleItem(t, m)
			assert.LessOrEqual(t, item.Height, ui.MaxLineRows,
				"len=%d %dx%d: a row must never outgrow the clamp", size, dim.w, dim.h)

			// The reservation and the render have to agree, or the viewport's row
			// arithmetic drifts and content falls off the bottom.
			drawn := strings.Count(m.renderLayoutItem(item), "\n") + 1
			assert.Equal(t, item.Height, drawn,
				"len=%d %dx%d: reserved %d rows, drew %d", size, dim.w, dim.h, item.Height, drawn)

			rows := strings.Count(m.renderView(), "\n") + 1
			assert.Equal(t, dim.h, rows, "len=%d %dx%d: the view must fill the screen exactly", size, dim.w, dim.h)
		}
	}
}

func TestLongTitle_SearchWindowFollowsTheMatch(t *testing.T) {
	// A match past the clamp would otherwise be highlighted on a row nobody can
	// see: n/N would jump the cursor to a task showing no visible hit.
	m := newTestModel(t, longTitleProject(longWords(4000)))
	m.ui.Screen.Width, m.ui.Screen.Height = 80, 24
	m.ui.Search.query = "ZEND"
	m.invalidateLayout()

	item, _ := longTitleItem(t, m)
	assert.Contains(t, m.renderLayoutItem(item), "ZEND", "the window should open on the match")
	assert.Equal(t, item.Height, strings.Count(m.renderLayoutItem(item), "\n")+1,
		"shifting the window must not change the row count")
}

func TestLongTitle_EditRowStaysOneRow(t *testing.T) {
	for _, size := range []int{20, 1024, 100_000} {
		m := newTestModel(t, longTitleProject("short"))
		m.ui.Screen.Width, m.ui.Screen.Height = 80, 24
		m.invalidateLayout()
		item, pos := longTitleItem(t, m)

		buffer := longWords(size)
		m.ui.Selection.SetSelected(pos)
		m.ui.Modes = modes.NewMachine(modes.ModeEdit)
		m.ui.Edit = newEditState(buffer, false, "", selection.FocusTask)

		// The layout sizes the row from the stored title, so a wrapping editor
		// would grow the row under it and push the bottom bar off the terminal.
		drawn := strings.Count(m.renderLayoutItem(item), "\n") + 1
		assert.Equal(t, 1, drawn, "len=%d: the edit row must stay one row", size)
		assert.LessOrEqual(t, strings.Count(m.renderView(), "\n")+1, 24, "len=%d: the view must not overflow the screen", size)

		// The cursor sits at the end of the buffer, so the window must have
		// scrolled to it rather than staying at the start of the text.
		assert.Contains(t, m.renderLayoutItem(item), "ZEND", "len=%d: the cursor's end of the buffer should be visible", size)
	}
}

func TestLongTitle_InfoDialogFitsAndScrolls(t *testing.T) {
	m := newTestModel(t, longTitleProject(longWords(4000)))
	m.ui.Screen.Width, m.ui.Screen.Height = 80, 24
	m.invalidateLayout()
	_, pos := longTitleItem(t, m)
	m.ui.Selection.SetSelected(pos)
	m.ui.Modes = modes.NewMachine(modes.ModeInfo)

	rows := strings.Count(m.infoView(), "\n") + 1
	assert.LessOrEqual(t, rows, 24, "an over-tall dialog is clipped by placeOverlay, hint and all")
	assert.Contains(t, m.renderView(), "close", "the hint must survive on screen")

	// The dialog is where the full title lives now, so its end has to be
	// reachable — unlike the clamped row in the list.
	assert.NotContains(t, m.infoView(), "ZEND", "the tail should start off screen")
	m.ui.Info.ScrollOffset = m.infoMaxScroll()
	assert.Contains(t, m.infoView(), "ZEND", "the end of the title must be reachable")
	assert.Contains(t, m.infoView(), "close", "the hint stays pinned below the window")

	m.ui.Info.ScrollOffset = 0
	before := m.infoView()
	m.scrollInfo(-5)
	assert.Equal(t, before, m.infoView(), "scrolling up at the top is a no-op")

	atBottom := m.infoMaxScroll()
	m.ui.Info.ScrollOffset = atBottom
	m.scrollInfo(1000)
	assert.Equal(t, atBottom, m.ui.Info.ScrollOffset, "scrolling past the end is a no-op")
}

func TestLongTitle_InfoDialogScrollKeys(t *testing.T) {
	m := newTestModel(t, longTitleProject(longWords(4000)))
	m.ui.Screen.Width, m.ui.Screen.Height = 80, 24
	m.invalidateLayout()
	_, pos := longTitleItem(t, m)
	m.ui.Selection.SetSelected(pos)
	m.ui.Modes = modes.NewMachine(modes.ModeInfo)

	press := func(key string) {
		*m = m.handleInfoKey(tea.KeyPressMsg{Code: []rune(key)[0], Text: key})
	}

	press("j")
	assert.Equal(t, 1, m.ui.Info.ScrollOffset)
	press("k")
	assert.Equal(t, 0, m.ui.Info.ScrollOffset)
	press("G")
	require.Positive(t, m.ui.Info.ScrollOffset)
	press("g")
	assert.Equal(t, 0, m.ui.Info.ScrollOffset)
}
