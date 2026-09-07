package journal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMintAndLoadDevice(t *testing.T) {
	dir := t.TempDir()
	minted, err := MintDevice(dir, "https://example.test", "tok")
	require.NoError(t, err)
	assert.NotEmpty(t, minted.DeviceID)

	loaded, ok, err := LoadDevice(dir)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, minted, loaded)

	info, err := os.Stat(filepath.Join(dir, "device.json"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(), "credential file must be owner-only")

	_, err = MintDevice(dir, "https://other.test", "tok2")
	assert.Error(t, err, "re-minting would orphan the server's cursor for the old identity")
}

func TestRecorderFollowsEnrollment(t *testing.T) {
	dir := t.TempDir()
	rec := NewRecorder(dir)
	assert.False(t, rec.Active(), "no device file → sync off")
	require.NoError(t, rec.RecordDelete("p1"))
	ops, err := Open(dir).Entries()
	require.NoError(t, err)
	assert.Empty(t, ops, "nothing journaled while unenrolled")

	_, err = MintDevice(dir, "https://example.test", "tok")
	require.NoError(t, err)
	assert.True(t, rec.Active())
	require.NoError(t, rec.RecordDelete("p1"))
	ops, err = Open(dir).Entries()
	require.NoError(t, err)
	require.Len(t, ops, 1)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "device.json"), []byte("{"), 0o600))
	assert.Error(t, rec.RecordDelete("p1"), "corrupt device file must not fall back to local-only")
}

func TestUnenrollRemovesIdentityAndJournal(t *testing.T) {
	dir := t.TempDir()
	_, err := MintDevice(dir, "https://example.test", "tok")
	require.NoError(t, err)
	require.NoError(t, NewRecorder(dir).RecordDelete("p1"))
	require.NoError(t, Unenroll(dir))
	_, ok, err := LoadDevice(dir)
	require.NoError(t, err)
	assert.False(t, ok)
	ops, err := Open(dir).Entries()
	require.NoError(t, err)
	assert.Empty(t, ops)
	require.NoError(t, Unenroll(dir), "idempotent")
}
