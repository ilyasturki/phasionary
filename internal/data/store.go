package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"phasionary/internal/domain"
)

var (
	ErrProjectNotFound      = errors.New("project not found")
	ErrDuplicateProjectName = errors.New("project name already exists")
)

type ProjectRepository interface {
	ListProjects() ([]domain.Project, error)
	LoadProject(selector string) (domain.Project, error)
	LoadProjectByID(id string) (domain.Project, error)
	SaveProjectLocked(project domain.Project) error
	WithProjectLocked(id string, fn func(*domain.Project) error) (domain.Project, error)
	CreateProject(name string) (domain.Project, error)
	RenameProject(id, newName string) (domain.Project, error)
	DeleteProject(id string) error
}

// Store manages JSON persistence in a directory.
type Store struct {
	Dir string
}

var _ ProjectRepository = (*Store)(nil)

func NewStore(dir string) *Store {
	return &Store{Dir: dir}
}

func (s *Store) Ensure() error {
	return os.MkdirAll(s.Dir, 0o755)
}

func (s *Store) ListProjects() ([]domain.Project, error) {
	entries, err := os.ReadDir(s.Dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []domain.Project{}, nil
		}
		return nil, err
	}
	projects := make([]domain.Project, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(s.Dir, entry.Name())
		project, err := s.loadProjectFile(path)
		if err != nil {
			return nil, err
		}
		if project.ID == "" {
			continue
		}
		projects = append(projects, project)
	}
	sort.Slice(projects, func(i, j int) bool {
		return strings.ToLower(projects[i].Name) < strings.ToLower(projects[j].Name)
	})
	return projects, nil
}

func (s *Store) LoadProjectByID(id string) (domain.Project, error) {
	path := s.projectPath(id)
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return domain.Project{}, ErrProjectNotFound
	}
	project, err := s.loadProjectFile(path)
	if err != nil {
		return domain.Project{}, err
	}
	// A blank id would cause saveProjectAtomic to write to "<dir>/.json",
	// silently splitting state; treat such files as not-found rather than
	// loading them as a zero-ID project.
	if project.ID == "" {
		return domain.Project{}, ErrProjectNotFound
	}
	return project, nil
}

func (s *Store) LoadProject(selector string) (domain.Project, error) {
	projects, err := s.ListProjects()
	if err != nil {
		return domain.Project{}, err
	}
	if len(projects) == 0 {
		return domain.Project{}, ErrProjectNotFound
	}
	if strings.TrimSpace(selector) == "" {
		return projects[0], nil
	}
	needle := domain.NormalizeName(selector)
	for _, project := range projects {
		if strings.EqualFold(project.ID, selector) || domain.NormalizeName(project.Name) == needle {
			return project, nil
		}
	}
	return domain.Project{}, ErrProjectNotFound
}

func (s *Store) saveProjectAtomic(project domain.Project) error {
	project.UpdatedAt = domain.NowTimestamp()
	data, err := json.MarshalIndent(project, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(s.Dir, 0o755); err != nil {
		return err
	}
	path := s.projectPath(project.ID)
	tmp := s.tmpPath(project.ID)
	if err := writeFileSync(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	// Flush the directory entry so the rename survives a crash.
	if dir, derr := os.Open(s.Dir); derr == nil {
		_ = dir.Sync()
		_ = dir.Close()
	}
	return nil
}

// writeFileSync writes data to path and fsyncs the file before closing.
// os.WriteFile only returns once the bytes are in the page cache, which is
// not crash-safe; the atomic rename pattern needs a flushed tmp file.
func writeFileSync(path string, data []byte, perm os.FileMode) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

func (s *Store) CreateProject(name string) (domain.Project, error) {
	// Hold the global lock across the duplicate-name check and the save so
	// two racing CreateProject calls can't both observe a missing name and
	// each write their own project file.
	g, err := s.acquireGlobalLock()
	if err != nil {
		return domain.Project{}, err
	}
	defer g.Close()

	if err := s.checkProjectNameAvailableLocked(name, ""); err != nil {
		return domain.Project{}, err
	}
	project, err := domain.NewProject(name)
	if err != nil {
		return domain.Project{}, err
	}
	project.Categories, err = s.defaultCategories()
	if err != nil {
		return domain.Project{}, err
	}
	project.Categories = populateSampleTasks(project.Categories)
	if err := s.saveNewProjectLocked(project); err != nil {
		return domain.Project{}, err
	}
	return project, nil
}

// RenameProject updates a project's display name while holding the global
// lock, so the unique-name invariant survives concurrent renames/creates
// from other processes. excludeID lets the duplicate check skip the
// project being renamed.
func (s *Store) RenameProject(id, newName string) (domain.Project, error) {
	g, err := s.acquireGlobalLock()
	if err != nil {
		return domain.Project{}, err
	}
	defer g.Close()

	if err := s.checkProjectNameAvailableLocked(newName, id); err != nil {
		return domain.Project{}, err
	}
	// Per-project flock for the read-modify-write itself. Global lock is
	// already held, so the acquisition order is global → project.
	if _, err := os.Stat(s.projectPath(id)); errors.Is(err, fs.ErrNotExist) {
		return domain.Project{}, ErrProjectNotFound
	}
	pl, err := s.acquireProjectLock(id)
	if err != nil {
		return domain.Project{}, err
	}
	defer pl.Close()
	project, err := s.LoadProjectByID(id)
	if err != nil {
		return domain.Project{}, err
	}
	project.Name = newName
	if err := s.saveProjectAtomic(project); err != nil {
		return domain.Project{}, err
	}
	return project, nil
}

// checkProjectNameAvailableLocked returns ErrDuplicateProjectName if any
// existing project (other than excludeID) shares the normalized name.
// Caller must hold the global lock.
func (s *Store) checkProjectNameAvailableLocked(name, excludeID string) error {
	projects, err := s.ListProjects()
	if err != nil {
		return err
	}
	needle := domain.NormalizeName(name)
	for _, p := range projects {
		if p.ID == excludeID {
			continue
		}
		if domain.NormalizeName(p.Name) == needle {
			return fmt.Errorf("%w: %q", ErrDuplicateProjectName, name)
		}
	}
	return nil
}

func (s *Store) InitDefault() (domain.Project, error) {
	if err := s.Ensure(); err != nil {
		return domain.Project{}, err
	}
	projects, err := s.ListProjects()
	if err != nil {
		return domain.Project{}, err
	}
	if len(projects) > 0 {
		return projects[0], nil
	}
	return s.CreateProject("Default")
}

func (s *Store) defaultCategories() ([]domain.Category, error) {
	categories := make([]domain.Category, 0, len(domain.DefaultCategories))
	for _, name := range domain.DefaultCategories {
		category, err := domain.NewCategory(name)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func (s *Store) loadProjectFile(path string) (domain.Project, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return domain.Project{}, err
	}
	var project domain.Project
	if err := json.Unmarshal(data, &project); err != nil {
		return domain.Project{}, err
	}
	return project, nil
}

func (s *Store) projectPath(id string) string {
	return filepath.Join(s.Dir, fmt.Sprintf("%s.json", id))
}

func (s *Store) lockPath(id string) string { return s.projectPath(id) + ".lock" }
func (s *Store) tmpPath(id string) string  { return s.projectPath(id) + ".tmp" }

func (s *Store) DeleteProject(id string) error {
	// Hold the project flock while deleting so a concurrent writer can't
	// race in mid-delete. We keep the .lock file on disk afterwards — unlinking
	// it would break mutual exclusion for any other process that's already
	// flocked the old inode (a fresh OpenFile would get a new inode and lock
	// it independently).
	f, err := s.acquireProjectLock(id)
	if err != nil {
		return err
	}
	defer f.Close()
	path := s.projectPath(id)
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return ErrProjectNotFound
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	_ = os.Remove(s.tmpPath(id))
	return nil
}

type sampleTask struct {
	title    string
	status   string
	priority string
	estimate int
}

var sampleTasksByCategory = map[string][]sampleTask{
	"Feature": {
		{"Build the main dashboard", domain.StatusInProgress, domain.PriorityHigh, 480},
		{"Add user preferences panel", domain.StatusTodo, domain.PriorityMedium, 240},
	},
	"Fix": {
		{"Resolve login timeout issue", domain.StatusTodo, domain.PriorityHigh, 60},
		{"Fix date formatting in reports", domain.StatusCompleted, domain.PriorityLow, 30},
	},
	"Ergonomy": {
		{"Improve keyboard navigation", domain.StatusInProgress, domain.PriorityMedium, 120},
		{"Add dark mode support", domain.StatusTodo, domain.PriorityLow, 240},
	},
	"Documentation": {
		{"Write getting started guide", domain.StatusTodo, domain.PriorityMedium, 120},
		{"Document API endpoints", domain.StatusTodo, domain.PriorityLow, 480},
	},
	"Research": {
		{"Evaluate caching strategies", domain.StatusCompleted, domain.PriorityMedium, 240},
		{"Investigate performance bottlenecks", domain.StatusTodo, domain.PriorityHigh, 120},
	},
}

func populateSampleTasks(categories []domain.Category) []domain.Category {
	for i := range categories {
		samples, ok := sampleTasksByCategory[categories[i].Name]
		if !ok {
			continue
		}
		for _, s := range samples {
			task, err := domain.NewTask(s.title)
			if err != nil {
				continue
			}
			_ = task.SetStatus(s.status)
			_ = task.SetPriority(s.priority)
			task.SetEstimate(s.estimate)
			categories[i].AddTask(task)
		}
	}
	return categories
}
