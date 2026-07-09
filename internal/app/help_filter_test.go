package app

import (
	"testing"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"phasionary/internal/ui"
)

// helpFiltering puts the model into help mode with an active `/` filter for the
// given query, then reconciles focus/scroll the way the key handler would.
func helpFiltering(t *testing.T, m *model, query string) {
	t.Helper()
	m.ui.Modes.ToHelp()
	ti := textinput.New()
	ti.SetValue(query)
	m.ui.Help = HelpState{Filtering: true, Filter: ti}
	m.ensureHelpVisible()
}

func TestHelpFilter_SlashEntersFiltering(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Modes.ToHelp()

	after, _ := m.handleHelpKey(tea.KeyPressMsg{Text: "/"})
	am := after.(model)
	assert.True(t, am.ui.Help.Filtering)
	assert.Equal(t, "", am.ui.Help.Filter.Value(), "filter starts empty")
	assert.Equal(t, 0, am.ui.Help.Focused)
}

func TestHelpFilter_EmptyQueryReturnsFullList(t *testing.T) {
	rows, focus := filteredHelpRows("")
	assert.Equal(t, helpRowsAll, rows)
	assert.Equal(t, helpFocusablesAll, focus)
}

func TestHelpFilter_NarrowsByLabel(t *testing.T) {
	rows, focus := filteredHelpRows("options")
	require.Len(t, focus, 1, "only the options binding matches its label")
	row := rows[focus[0]]
	b := normalBindings[row.bindingIndex]
	assert.Equal(t, "options", b.desc)
	assert.True(t, rows[0].header, "the section header is kept above the match")
}

func TestHelpFilter_MatchesByShortcutKey(t *testing.T) {
	// "?" appears only as the toggle-help binding's key, so keying it surfaces
	// exactly that row — which is disabled and therefore not runnable.
	rows, focus := filteredHelpRows("?")
	require.Len(t, focus, 1)
	row := rows[focus[0]]
	assert.Equal(t, "toggle help", normalBindings[row.bindingIndex].desc)
	assert.True(t, row.disabled)
}

func TestHelpFilter_MatchesDisplayOnlyVisualRow(t *testing.T) {
	// "markdown" only appears in the Visual-mode reference section, which is
	// searchable but not runnable.
	rows, focus := filteredHelpRows("markdown")
	assert.Empty(t, focus, "reference rows are display-only, never focusable")

	matched := false
	for _, r := range rows {
		if r.filter != "" && ui.Contains(r.filter, "markdown") {
			matched = true
		}
	}
	assert.True(t, matched, "the Visual-mode markdown row still shows up")
}

func TestHelpFilter_EnterRunsFocusedMatchAndCloses(t *testing.T) {
	m := newTestModel(t, sampleProject())
	helpFiltering(t, m, "options")

	after, _ := m.handleHelpKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	am := after.(model)
	assert.True(t, am.ui.Modes.IsOptions(), "Enter fires the matched binding")
}

func TestHelpFilter_EscClearsFilterButKeepsHelpOpen(t *testing.T) {
	m := newTestModel(t, sampleProject())
	helpFiltering(t, m, "options")

	after, _ := m.handleHelpKey(tea.KeyPressMsg{Code: tea.KeyEscape})
	am := after.(model)
	assert.False(t, am.ui.Help.Filtering, "Esc leaves the filter")
	assert.True(t, am.ui.Modes.IsHelp(), "Esc stays in the help dialog")
	assert.Equal(t, "", am.ui.Help.Filter.Value())
}

func TestHelpFilter_NoMatchHasNoFocusables(t *testing.T) {
	rows, focus := filteredHelpRows("zzzznomatch")
	assert.Empty(t, rows)
	assert.Empty(t, focus)
}
