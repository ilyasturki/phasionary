package app

import (
	"os"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"phasionary/internal/app/modes"
	"phasionary/internal/app/selection"
)

// descriptionMarkerRow renders the description textarea and returns the index of
// the single display row whose ">" gutter is drawn (i.e. the cursor's row).
func descriptionMarkerRow(t *testing.T, m *model) int {
	t.Helper()
	require.NotNil(t, m.ui.DescriptionEdit.cursorRow)
	*m.ui.DescriptionEdit.cursorRow = descriptionCursorDisplayRow(m.ui.DescriptionEdit.textarea)
	rows := strings.Split(ansi.Strip(m.ui.DescriptionEdit.textarea.View()), "\n")
	found := -1
	for i, r := range rows {
		if strings.HasPrefix(r, "> ") {
			require.Equal(t, -1, found, "expected one gutter marker, found rows %d and %d", found, i)
			found = i
		}
	}
	require.NotEqual(t, -1, found, "no > gutter marker in:\n%s", strings.Join(rows, "\n"))
	return found
}

func TestDescriptionEdit_GutterMarkerFollowsCursor(t *testing.T) {
	p := sampleProject()
	p.Categories[0].Tasks[0].Description = "alpha\nbeta\ngamma"
	m := newTestModel(t, p)
	m.ui.Screen.Width = 60
	m.ui.Screen.Height = 24
	_ = m.startDescriptionInlineEdit(0, 0)

	// Cursor starts at the end → last logical line.
	require.Equal(t, 2, descriptionMarkerRow(t, m))

	// Moving up walks the marker with the cursor, one row at a time.
	_, _ = m.handleDescriptionEditKey(tea.KeyPressMsg{Code: tea.KeyUp})
	require.Equal(t, 1, descriptionMarkerRow(t, m))
	_, _ = m.handleDescriptionEditKey(tea.KeyPressMsg{Code: tea.KeyUp})
	require.Equal(t, 0, descriptionMarkerRow(t, m))
}

func TestStartDescriptionInlineEdit_EntersModeAndPrefillsTextarea(t *testing.T) {
	p := sampleProject()
	p.Categories[0].Tasks[0].Description = "Hello\nworld"
	m := newTestModel(t, p)

	_ = m.startDescriptionInlineEdit(0, 0)

	assert.True(t, m.ui.Modes.IsDescriptionEdit(), "should be in description-edit mode")
	assert.Equal(t, "Hello\nworld", m.ui.DescriptionEdit.textarea.Value())
	assert.Equal(t, 0, m.ui.DescriptionEdit.categoryIndex)
	assert.Equal(t, 0, m.ui.DescriptionEdit.taskIndex)
	assert.Equal(t, "Hello\nworld", m.ui.DescriptionEdit.original)
	assert.False(t, m.ui.DescriptionEdit.creating)
}

func TestStartDescriptionInlineEdit_CreatingWhenEmpty(t *testing.T) {
	m := newTestModel(t, sampleProject()) // t1 has no description
	_ = m.startDescriptionInlineEdit(0, 0)
	assert.True(t, m.ui.DescriptionEdit.creating, "creating flag set when starting from empty description")
}

func TestFinishDescriptionEdit_SavesAndLeavesMode(t *testing.T) {
	p := sampleProject()
	m := newTestModel(t, p)
	_ = m.startDescriptionInlineEdit(0, 0)
	m.ui.DescriptionEdit.textarea.SetValue("New body line 1\nline 2")
	m.finishDescriptionEdit()

	assert.True(t, m.ui.Modes.IsNormal())
	assert.Equal(t, "New body line 1\nline 2", m.project.Categories[0].Tasks[0].Description)
	assert.Equal(t, "Description added", m.ui.Screen.StatusMsg)
}

func TestFinishDescriptionEdit_EmptyValueClearsDescription(t *testing.T) {
	p := sampleProject()
	p.Categories[0].Tasks[0].Description = "to be removed"
	m := newTestModel(t, p)
	_ = m.startDescriptionInlineEdit(0, 0)
	m.ui.DescriptionEdit.textarea.SetValue("   ") // whitespace only
	m.finishDescriptionEdit()

	assert.Equal(t, "", m.project.Categories[0].Tasks[0].Description)
	assert.Equal(t, "Description cleared", m.ui.Screen.StatusMsg)
}

func TestCancelDescriptionEdit_DiscardsChanges(t *testing.T) {
	p := sampleProject()
	p.Categories[0].Tasks[0].Description = "original"
	m := newTestModel(t, p)
	_ = m.startDescriptionInlineEdit(0, 0)
	m.ui.DescriptionEdit.textarea.SetValue("would-be-new")
	m.cancelDescriptionEdit()

	assert.True(t, m.ui.Modes.IsNormal())
	assert.Equal(t, "original", m.project.Categories[0].Tasks[0].Description,
		"cancel must not persist the typed value")
}

func TestHandleDescriptionEditKey_EscCancels(t *testing.T) {
	m := newTestModel(t, sampleProject())
	_ = m.startDescriptionInlineEdit(0, 0)
	m.ui.DescriptionEdit.textarea.SetValue("typed but not saved")

	_, _ = m.handleDescriptionEditKey(tea.KeyPressMsg{Code: tea.KeyEscape})

	assert.True(t, m.ui.Modes.IsNormal())
	assert.Equal(t, "", m.project.Categories[0].Tasks[0].Description)
}

func TestHandleDescriptionEditKey_EnterSaves(t *testing.T) {
	p := sampleProject()
	p.Categories[0].Tasks[0].Description = "existing body"
	m := newTestModel(t, p)
	_ = m.startDescriptionInlineEdit(0, 0)

	_, _ = m.handleDescriptionEditKey(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.True(t, m.ui.Modes.IsNormal(), "plain enter saves and leaves edit mode")
	assert.Equal(t, "existing body", m.project.Categories[0].Tasks[0].Description)
}

func TestHandleDescriptionEditKey_ShiftEnterInsertsNewline(t *testing.T) {
	// Both wire encodings of Shift+Enter must insert a newline without saving:
	// a genuine "shift+enter" and the "alt+enter" (ESC+CR) that kitty/ghostty
	// deliver.
	for _, key := range []tea.KeyPressMsg{
		{Code: tea.KeyEnter, Mod: tea.ModShift},
		{Code: tea.KeyEnter, Mod: tea.ModAlt},
	} {
		p := sampleProject()
		p.Categories[0].Tasks[0].Description = "abc" // cursor starts at end
		m := newTestModel(t, p)
		_ = m.startDescriptionInlineEdit(0, 0)

		_, _ = m.handleDescriptionEditKey(key)

		assert.True(t, m.ui.Modes.IsDescriptionEdit(), "%s must not save", key.String())
		assert.Equal(t, "abc\n", m.ui.DescriptionEdit.textarea.Value(),
			"%s must insert a newline at the cursor", key.String())
	}
}

func TestHandleDescriptionEditKey_CtrlSSaves(t *testing.T) {
	m := newTestModel(t, sampleProject())
	_ = m.startDescriptionInlineEdit(0, 0)
	m.ui.DescriptionEdit.textarea.SetValue("saved via ctrl+s")

	_, _ = m.handleDescriptionEditKey(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})

	assert.True(t, m.ui.Modes.IsNormal())
	assert.Equal(t, "saved via ctrl+s", m.project.Categories[0].Tasks[0].Description)
}

func TestEditOrFocusDescription_TaskWithoutDescription_EntersInlineEditMode(t *testing.T) {
	// Regression guard for the v2 of this feature: pressing shift+enter on a
	// task with no description used to open the external editor; it should
	// now drop the user straight into the in-app textarea.
	m := newTestModel(t, sampleProject())
	m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
		return p.Kind == selection.FocusTask && p.CategoryIndex == 0 && p.TaskIndex == 0
	})

	_ = m.editOrFocusDescription()

	require.Equal(t, modes.ModeDescriptionEdit, m.ui.Modes.Current())
	assert.Equal(t, "", m.ui.DescriptionEdit.textarea.Value())
	assert.True(t, m.ui.DescriptionEdit.creating)
}

func TestStartExternalEdit_FromDescriptionRow_EditsWholeTask(t *testing.T) {
	// `e` on a description row should open the external editor with the full
	// title + description body, not just the description.
	p := sampleProject()
	p.Categories[0].Tasks[0].Description = "Body text"
	m := newTestModel(t, p)
	m.ui.Screen.ExpandDescriptions = true
	m.rebuildPositions()
	m.ui.Selection.SelectByPredicate(func(pp selection.Position) bool {
		return pp.Kind == selection.FocusDescription && pp.CategoryIndex == 0 && pp.TaskIndex == 0
	})

	// We can't easily run the editor process inside a unit test, but we can
	// verify that the state transition uses ItemType = FocusTask.
	_ = m.startExternalEdit()
	assert.Equal(t, selection.FocusTask, m.ui.ExternalEdit.ItemType,
		"external edit from a description row must be scoped to the whole task")
	assert.Equal(t, 0, m.ui.ExternalEdit.CategoryIndex)
	assert.Equal(t, 0, m.ui.ExternalEdit.TaskIndex)
	// Clean up the temp file that startExternalEdit created.
	if m.ui.ExternalEdit.TempFilePath != "" {
		_ = os.Remove(m.ui.ExternalEdit.TempFilePath)
	}
}
