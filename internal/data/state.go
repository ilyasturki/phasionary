package data

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

type State struct {
	DirectoryProjects map[string]string   `json:"directory_projects,omitempty"`
	ProjectOrder      []string            `json:"project_order,omitempty"`
	FoldedCategories  map[string][]string `json:"folded_categories,omitempty"`
}

// StateRepository abstracts persistence of UI state. Implemented by *StateManager.
type StateRepository interface {
	GetProjectForDir() string
	SetProjectForDir(id string) error
	GetProjectOrder() []string
	SetProjectOrder(order []string) error
	GetFoldedCategories(projectID string) []string
	SetFoldedCategories(projectID string, categoryIDs []string) error
	DeleteFoldedCategories(projectID string) error
}

type StateManager struct {
	path       string
	currentDir string
	state      State
}

var _ StateRepository = (*StateManager)(nil)

// NewStateManager creates a state manager that persists to <stateDir>/state.json.
// currentDir is the working directory that owns the linked project; pass "" for an
// unscoped (global fallback) entry.
func NewStateManager(stateDir, currentDir string) *StateManager {
	return &StateManager{
		path:       filepath.Join(stateDir, "state.json"),
		currentDir: currentDir,
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
		LastProjectID     string              `json:"last_project_id"`
		DirectoryProjects map[string]string   `json:"directory_projects,omitempty"`
		ProjectOrder      []string            `json:"project_order,omitempty"`
		FoldedCategories  map[string][]string `json:"folded_categories,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	m.state.DirectoryProjects = raw.DirectoryProjects
	m.state.ProjectOrder = raw.ProjectOrder
	m.state.FoldedCategories = raw.FoldedCategories
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

// GetProjectForDir returns the project ID owned by the manager's current
// directory, or "" if none is linked.
func (m *StateManager) GetProjectForDir() string {
	if m.state.DirectoryProjects == nil {
		return ""
	}
	return m.state.DirectoryProjects[m.currentDir]
}

// SetProjectForDir links the manager's current directory to the given project ID.
func (m *StateManager) SetProjectForDir(id string) error {
	if m.state.DirectoryProjects == nil {
		m.state.DirectoryProjects = make(map[string]string)
	}
	m.state.DirectoryProjects[m.currentDir] = id
	return m.Save()
}

// UnlinkDir removes the link for the manager's current directory. Returns the
// previously linked project ID (empty if none was linked).
func (m *StateManager) UnlinkDir() (string, error) {
	if m.state.DirectoryProjects == nil {
		return "", nil
	}
	prev, ok := m.state.DirectoryProjects[m.currentDir]
	if !ok {
		return "", nil
	}
	delete(m.state.DirectoryProjects, m.currentDir)
	return prev, m.Save()
}

func (m *StateManager) GetProjectOrder() []string {
	return m.state.ProjectOrder
}

func (m *StateManager) SetProjectOrder(order []string) error {
	m.state.ProjectOrder = order
	return m.Save()
}

func (m *StateManager) GetFoldedCategories(projectID string) []string {
	if m.state.FoldedCategories == nil {
		return nil
	}
	return m.state.FoldedCategories[projectID]
}

func (m *StateManager) SetFoldedCategories(projectID string, categoryIDs []string) error {
	if m.state.FoldedCategories == nil {
		m.state.FoldedCategories = make(map[string][]string)
	}
	if len(categoryIDs) == 0 {
		delete(m.state.FoldedCategories, projectID)
	} else {
		m.state.FoldedCategories[projectID] = categoryIDs
	}
	return m.Save()
}

func (m *StateManager) DeleteFoldedCategories(projectID string) error {
	if m.state.FoldedCategories == nil {
		return nil
	}
	delete(m.state.FoldedCategories, projectID)
	return m.Save()
}
