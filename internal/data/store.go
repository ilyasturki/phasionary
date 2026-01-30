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

var ErrProjectNotFound = errors.New("project not found")

type ProjectRepository interface {
	ListProjects() ([]domain.Project, error)
	LoadProject(selector string) (domain.Project, error)
	SaveProject(project domain.Project) error
	CreateProject(name string) (domain.Project, error)
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
		projects = append(projects, project)
	}
	sort.Slice(projects, func(i, j int) bool {
		return strings.ToLower(projects[i].Name) < strings.ToLower(projects[j].Name)
	})
	return projects, nil
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

func (s *Store) SaveProject(project domain.Project) error {
	project.UpdatedAt = domain.NowTimestamp()
	path := s.projectPath(project.ID)
	data, err := json.MarshalIndent(project, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(s.Dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (s *Store) CreateProject(name string) (domain.Project, error) {
	projects, err := s.ListProjects()
	if err != nil {
		return domain.Project{}, err
	}
	needle := domain.NormalizeName(name)
	for _, project := range projects {
		if domain.NormalizeName(project.Name) == needle {
			return domain.Project{}, fmt.Errorf("project %q already exists", name)
		}
	}
	project, err := domain.NewProject(name)
	if err != nil {
		return domain.Project{}, err
	}
	project.Categories, err = s.defaultCategories()
	if err != nil {
		return domain.Project{}, err
	}
	if err := s.SaveProject(project); err != nil {
		return domain.Project{}, err
	}
	return project, nil
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
