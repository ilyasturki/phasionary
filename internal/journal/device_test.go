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

func TestOpenIfConfigured(t *testing.T) {
	dir := t.TempDir()

	rec, err := OpenIfConfigured(dir)
	require.NoError(t, err)
	assert.Nil(t, rec, "no device file → sync off")

	_, err = MintDevice(dir, "https://example.test", "tok")
	require.NoError(t, err)
	rec, err = OpenIfConfigured(dir)
	require.NoError(t, err)
	require.NotNil(t, rec)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "device.json"), []byte("{"), 0o600))
	_, err = OpenIfConfigured(dir)
	assert.Error(t, err, "corrupt device file must not fall back to local-only")
}
