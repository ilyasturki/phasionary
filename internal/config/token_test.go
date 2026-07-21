package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServeToken(t *testing.T) {
	t.Run("is unguessable and distinct per call", func(t *testing.T) {
		seen := make(map[string]bool, 64)
		for range 64 {
			token, err := NewServeToken()
			require.NoError(t, err)
			// 32 random bytes in unpadded base64url.
			assert.Len(t, token, 43)
			assert.False(t, seen[token], "token repeated: %q", token)
			seen[token] = true
		}
	})

	t.Run("is safe to paste anywhere", func(t *testing.T) {
		// It travels through a shell argument, an HTTP header and a phone's text
		// field, so it must not need escaping in any of them.
		token, err := NewServeToken()
		require.NoError(t, err)
		for _, r := range token {
			ok := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
				(r >= '0' && r <= '9') || r == '-' || r == '_'
			assert.True(t, ok, "unexpected character %q in token", r)
		}
	})
}

func TestConfigFilePermissions(t *testing.T) {
	t.Run("a new config is owner-only", func(t *testing.T) {
		// The file holds the serve token, which grants full read/write access to
		// every project. World-readable would hand it to every other account on
		// the machine.
		path := filepath.Join(t.TempDir(), "config.json")
		m := NewManager(path)
		require.NoError(t, m.Load()) // creates it

		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, configFileMode, info.Mode().Perm())
	})

	t.Run("an existing world-readable config is tightened on load", func(t *testing.T) {
		// Configs written before the token lived here are 0644, and os.WriteFile
		// keeps an existing file's mode — so without an explicit chmod, upgrading
		// and then generating a token would leave it world-readable.
		path := filepath.Join(t.TempDir(), "config.json")
		require.NoError(t, os.WriteFile(path, []byte("{}"), 0o644))

		m := NewManager(path)
		require.NoError(t, m.Load())

		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, configFileMode, info.Mode().Perm())
	})

	t.Run("saving a token keeps the file owner-only", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.json")
		require.NoError(t, os.WriteFile(path, []byte("{}"), 0o644))

		m := NewManager(path)
		require.NoError(t, m.Load())

		token, err := NewServeToken()
		require.NoError(t, err)
		require.NoError(t, m.Update(func(c *Config) { c.ServeToken = token }))

		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, configFileMode, info.Mode().Perm())

		// And it round-trips, so serve reuses the same token next run rather
		// than minting a new one and breaking the phone again.
		reloaded := NewManager(path)
		require.NoError(t, reloaded.Load())
		assert.Equal(t, token, reloaded.Get().ServeToken)
	})
}
