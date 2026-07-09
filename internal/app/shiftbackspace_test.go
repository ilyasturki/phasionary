package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// shiftBackspaceMsg is the key an enhanced-keyboard-protocol terminal delivers
// when the user presses Shift+Backspace: a Backspace code carrying the Shift
// modifier, distinct from a plain Backspace.
func shiftBackspaceMsg() tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: tea.KeyBackspace, Mod: tea.ModShift}
}

// TestShiftBackspaceStringForm pins the reason the fix is needed: Shift+Backspace
// stringifies distinctly from "backspace", so bubbles' delete-backward binding
// (which lists only "backspace"/"ctrl+h") ignores it without normalization.
func TestShiftBackspaceStringForm(t *testing.T) {
	assert.Equal(t, "shift+backspace", shiftBackspaceMsg().String())
	assert.Equal(t, "backspace", normalizeKey(shiftBackspaceMsg()).String(),
		"normalizeKey must strip the Shift modifier so the input treats it as Backspace")
}

// TestShiftBackspaceDeletesInTextInput is the regression guard for the single-line
// inline editor: Shift+Backspace must delete a character exactly like Backspace.
func TestShiftBackspaceDeletesInTextInput(t *testing.T) {
	p := sampleProject()
	m := newTestModel(t, p)
	selectFirstTask(m)

	m.startEditing()
	require.True(t, m.ui.Modes.IsEdit())
	m.ui.Edit.input.SetValue("abc")
	m.ui.Edit.input.CursorEnd()

	m.handleKeyMsg(shiftBackspaceMsg())

	assert.Equal(t, "ab", m.ui.Edit.input.Value(),
		"shift+backspace should delete one character like backspace")
}

// TestShiftBackspaceDeletesInDescriptionEditor guards the multi-line textarea,
// which shares the same delete-backward binding gap as the single-line input.
func TestShiftBackspaceDeletesInDescriptionEditor(t *testing.T) {
	p := sampleProject()
	p.Categories[0].Tasks[0].Description = "hello"
	m := newTestModel(t, p)
	selectFirstTask(m)

	require.NotNil(t, m.startDescriptionInlineEdit(0, 0))
	require.True(t, m.ui.Modes.IsDescriptionEdit())
	m.ui.DescriptionEdit.textarea.SetValue("abc")
	m.ui.DescriptionEdit.textarea.CursorEnd()

	m.handleKeyMsg(shiftBackspaceMsg())

	assert.Equal(t, "ab", m.ui.DescriptionEdit.textarea.Value(),
		"shift+backspace should delete one character in the description editor")
}
