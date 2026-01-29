package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Manager handles loading and saving configuration.
type Manager struct {
	path string
	cfg  Config
}

// NewManager creates a new config manager for the given config file path.
func NewManager(configPath string) *Manager {
	return &Manager{
		path: configPath,
		cfg:  DefaultConfig(),
	}
}

// Load reads the config from disk. Creates a default config file if missing.
func (m *Manager) Load() error {
	data, err := os.ReadFile(m.path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return m.Save()
		}
		return fmt.Errorf("reading config %s: %w", m.path, err)
	}
	if err := json.Unmarshal(data, &m.cfg); err != nil {
		return fmt.Errorf("parsing config %s: %w", m.path, err)
	}
	return nil
}

// Save writes the current config to disk. Creates the directory if needed.
func (m *Manager) Save() error {
	dir := filepath.Dir(m.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config directory %s: %w", dir, err)
	}
	data, err := json.MarshalIndent(m.cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	if err := os.WriteFile(m.path, data, 0o644); err != nil {
		return fmt.Errorf("writing config %s: %w", m.path, err)
	}
	return nil
}

// Get returns the current config.
func (m *Manager) Get() Config {
	return m.cfg
}

// Update applies a function to modify the config and saves it.
func (m *Manager) Update(fn func(*Config)) error {
	fn(&m.cfg)
	return m.Save()
}
