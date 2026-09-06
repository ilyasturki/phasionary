package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"phasionary/internal/config"
	"phasionary/internal/journal"
)

func TestSyncStatusUnconfigured(t *testing.T) {
	t.Setenv(config.EnvStatePath, t.TempDir())

	out, err := runCLI(t, t.TempDir(), "sync", "status")
	require.NoError(t, err)
	assert.Contains(t, out, "not configured")
}

func TestSyncStatusConfiguredCountsPending(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv(config.EnvStatePath, stateDir)
	device, err := journal.MintDevice(stateDir, "https://example.test", "tok")
	require.NoError(t, err)

	dataDir := t.TempDir()
	_, err = runCLI(t, dataDir, "init", "Synced")
	require.NoError(t, err)

	out, err := runCLI(t, dataDir, "sync", "status")
	require.NoError(t, err)
	assert.Contains(t, out, "Sync: configured")
	assert.Contains(t, out, device.DeviceID)
	assert.Contains(t, out, "https://example.test")
	assert.NotContains(t, out, "0 journaled", "the init must have journaled the creation cascade")
}

func TestCLIWritesJournalWhenEnrolled(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv(config.EnvStatePath, stateDir)
	_, err := journal.MintDevice(stateDir, "https://example.test", "tok")
	require.NoError(t, err)

	dataDir := t.TempDir()
	_, err = runCLI(t, dataDir, "init", "Synced")
	require.NoError(t, err)
	_, err = runCLI(t, dataDir, "task", "add", "Journaled task", "-p", "Synced", "--category", "Feature")
	require.NoError(t, err)

	ops, err := journal.Open(stateDir).Entries()
	require.NoError(t, err)
	require.NotEmpty(t, ops)
	assert.Equal(t, journal.KindProjectCreate, ops[0].Kind)

	var found bool
	for _, op := range ops {
		if op.Kind == journal.KindTaskCreate && op.Fields["title"] == "Journaled task" {
			found = true
		}
	}
	assert.True(t, found, "the task add must appear in the journal")
}

func TestCLIJournalsNothingWhenLocalOnly(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv(config.EnvStatePath, stateDir)

	dataDir := t.TempDir()
	_, err := runCLI(t, dataDir, "init", "Local")
	require.NoError(t, err)

	ops, err := journal.Open(stateDir).Entries()
	require.NoError(t, err)
	assert.Empty(t, ops)
}
