package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// focusedHelpEntry names where the help cursor sits, as "keys/desc".
func focusedHelpEntry(t *testing.T, m *model) helpEntry {
	t.Helper()
	body := m.helpBody()
	require.NotEmpty(t, body.targets)
	require.Less(t, m.ui.Help.Focused, len(body.targets))
	return *body.entryAt(body.targets[m.ui.Help.Focused])
}

func TestHelpColumn_CrossesToTheOtherColumn(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Modes.ToHelp()
	m.ensureHelpVisible()

	require.Equal(t, "j / k", focusedHelpEntry(t, m).keys)
	m.moveHelpColumn(1)
	assert.Equal(t, "a", focusedHelpEntry(t, m).keys)
}

// The card's two columns are different lengths, so crossing back lands on the
// nearest entry rather than the one the cursor left from.
func TestHelpColumn_LandsOnTheNearestLine(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Modes.ToHelp()
	m.ensureHelpVisible()

	m.moveHelpColumn(1)
	m.moveHelpFocus(3) // down to `d / delete`, the last of Create & edit
	require.Equal(t, "d", focusedHelpEntry(t, m).keys)

	// `d` sits on the blank line between Move and Do, so `search` above it is
	// closer than `cycle status` below.
	m.moveHelpColumn(-1)
	assert.Equal(t, "/", focusedHelpEntry(t, m).keys)
}

func TestHelpColumn_DoesNotWrap(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Modes.ToHelp()
	m.ensureHelpVisible()

	m.moveHelpColumn(-1)
	assert.Equal(t, "j / k", focusedHelpEntry(t, m).keys)

	m.moveHelpColumn(1)
	m.moveHelpColumn(1)
	assert.Equal(t, "a", focusedHelpEntry(t, m).keys)
}

// The full reference is one column, so there is nothing to cross to.
func TestHelpColumn_InertOnTheReference(t *testing.T) {
	m := newTestModel(t, sampleProject())
	helpExpand(t, m)
	m.moveHelpFocus(2)

	before := focusedHelpEntry(t, m)
	m.moveHelpColumn(1)
	m.moveHelpColumn(-1)
	assert.Equal(t, before, focusedHelpEntry(t, m))
}

func TestHelpColumn_InertWhileFiltering(t *testing.T) {
	m := newTestModel(t, sampleProject())
	helpFiltering(t, m, "cat")

	before := focusedHelpEntry(t, m)
	m.moveHelpColumn(1)
	assert.Equal(t, before, focusedHelpEntry(t, m))
}
