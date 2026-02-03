package data

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

type State struct {
	DirectoryProjects map[string]string `json:"directory_projects,omitempty"`
	ProjectOrder      []string          `json:"project_order,omitempty"`
}

type StateManager struct {
	path       string
	currentDir string
	state      State
}

func NewStateManager(dataDir, workingDir string) *StateManager {
	return &StateManager{
		path:       filepath.Join(dataDir, "state.json"),
		currentDir: workingDir,
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

	// Use a temporary struct to handle migration from old format
	var raw struct {
		LastProjectID     string            `json:"last_project_id"`
		DirectoryProjects map[string]string `json:"directory_projects,omitempty"`
		ProjectOrder      []string          `json:"project_order,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	m.state.DirectoryProjects = raw.DirectoryProjects
	m.state.ProjectOrder = raw.ProjectOrder
	if m.state.DirectoryProjects == nil {
		m.state.DirectoryProjects = make(map[string]string)
	}

	// Migrate old last_project_id to directory_projects[""] if not already set
	if raw.LastProjectID != "" {
		if _, ok := m.state.DirectoryProjects[""]; !ok {
			m.state.DirectoryProjects[""] = raw.LastProjectID
		}
	}

	return nil
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
	if m.state.DirectoryProjects == nil {
		return ""
	}
	if m.currentDir != "" {
		if id, ok := m.state.DirectoryProjects[m.currentDir]; ok {
			return id
		}
	}
	return m.state.DirectoryProjects[""]
}

func (m *StateManager) SetLastProjectID(id string) error {
	if m.state.DirectoryProjects == nil {
		m.state.DirectoryProjects = make(map[string]string)
	}
	if m.currentDir != "" {
		m.state.DirectoryProjects[m.currentDir] = id
	}
	m.state.DirectoryProjects[""] = id
	return m.Save()
}

func (m *StateManager) GetProjectOrder() []string {
	return m.state.ProjectOrder
}

func (m *StateManager) SetProjectOrder(order []string) error {
	m.state.ProjectOrder = order
	return m.Save()
}
