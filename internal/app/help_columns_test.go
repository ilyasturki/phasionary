package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func focusedHelpEntry(t *testing.T, m *model) helpEntry {
	t.Helper()
	body := m.helpBody()
	require.NotEmpty(t, body.targets)
	require.Less(t, m.ui.Help.Focused, len(body.targets))
	return *body.entryAt(body.targets[m.ui.Help.Focused])
}

func TestHelpColumn_OnTheCard(t *testing.T) {
	for _, tc := range []struct {
		name string
		move func(t *testing.T, m *model)
		want string
	}{
		{"crosses over", func(_ *testing.T, m *model) { m.moveHelpColumn(1) }, "a"},
		{"no wrap right", func(_ *testing.T, m *model) { m.moveHelpColumn(1); m.moveHelpColumn(1) }, "a"},
		{"no wrap left", func(_ *testing.T, m *model) { m.moveHelpColumn(-1) }, "j / k"},
		// `d` sits on the blank between Move and Do, so `search` above is nearer
		// than `cycle status` below.
		{"nearest line", func(t *testing.T, m *model) {
			m.moveHelpColumn(1)
			m.moveHelpFocus(3)
			require.Equal(t, "d", focusedHelpEntry(t, m).keys)
			m.moveHelpColumn(-1)
		}, "/"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newTestModel(t, sampleProject())
			require.Equal(t, "j / k", focusedHelpEntry(t, m).keys)
			tc.move(t, m)
			assert.Equal(t, tc.want, focusedHelpEntry(t, m).keys)
		})
	}
}

func TestHelpColumn_InertOnOneColumn(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(t *testing.T, m *model)
	}{
		{"full reference", func(t *testing.T, m *model) { helpExpand(t, m); m.moveHelpFocus(2) }},
		{"filtering", func(t *testing.T, m *model) { helpFiltering(t, m, "cat") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newTestModel(t, sampleProject())
			tc.setup(t, m)

			before := focusedHelpEntry(t, m)
			m.moveHelpColumn(1)
			m.moveHelpColumn(-1)
			assert.Equal(t, before, focusedHelpEntry(t, m))
		})
	}
}
