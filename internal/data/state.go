package data

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

type State struct {
	LastProjectID string `json:"last_project_id"`
}

type StateManager struct {
	path  string
	state State
}

func NewStateManager(dataDir string) *StateManager {
	return &StateManager{
		path: filepath.Join(dataDir, "state.json"),
	}
}

func (m *StateManager) Load() error {
	data, err := os.ReadFile(m.path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			m.state = State{}
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &m.state)
}

func (m *StateManager) Save() error {
	data, err := json.MarshalIndent(m.state, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(m.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(m.path, data, 0o644)
}

func (m *StateManager) GetLastProjectID() string {
	return m.state.LastProjectID
}

func (m *StateManager) SetLastProjectID(id string) error {
	m.state.LastProjectID = id
	return m.Save()
}
