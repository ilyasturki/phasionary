package cli

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/viper"

	"phasionary/internal/data"
	"phasionary/internal/domain"
)

var ErrNotFound = errors.New("not found")

func projectSelector(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return viper.GetString("project")
}

func loadStoreAndProject(selector string) (*data.Store, domain.Project, error) {
	store, err := storeFromViper()
	if err != nil {
		return nil, domain.Project{}, err
	}
	project, err := store.LoadProject(selector)
	if err != nil {
		return nil, domain.Project{}, err
	}
	return store, project, nil
}

func withProject(selector string, fn func(*data.Store, *domain.Project) error) error {
	store, project, err := loadStoreAndProject(selector)
	if err != nil {
		return err
	}
	if err := fn(store, &project); err != nil {
		return err
	}
	return store.SaveProject(project)
}

func viewProject(selector string, fn func(domain.Project) error) error {
	_, project, err := loadStoreAndProject(selector)
	if err != nil {
		return err
	}
	return fn(project)
}

var timeEstimateRe = regexp.MustCompile(`^(?:(\d+(?:\.\d+)?)h)?(?:(\d+)m?)?$`)

func parseTimeEstimate(input string) (int, error) {
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return 0, nil
	}

	if mins, err := strconv.Atoi(input); err == nil {
		return mins, nil
	}

	m := timeEstimateRe.FindStringSubmatch(input)
	if m == nil {
		return 0, fmt.Errorf("invalid time estimate format: %s", input)
	}

	var total float64
	if m[1] != "" {
		hours, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			return 0, err
		}
		total += hours * 60
	}
	if m[2] != "" {
		mins, err := strconv.Atoi(m[2])
		if err != nil {
			return 0, err
		}
		total += float64(mins)
	}

	return int(total), nil
}

type TaskRef struct {
	Task          *domain.Task
	CategoryName  string
	CategoryIndex int
	TaskIndex     int
}

func resolveTask(project domain.Project, selector string) (TaskRef, error) {
	selector = strings.TrimSpace(selector)
	if selector == "" {
		return TaskRef{}, ErrNotFound
	}

	needle := domain.NormalizeName(selector)
	const minPrefixLen = 4

	for cIdx := range project.Categories {
		for tIdx := range project.Categories[cIdx].Tasks {
			task := &project.Categories[cIdx].Tasks[tIdx]

			matched := task.ID == selector ||
				(len(selector) >= minPrefixLen && strings.HasPrefix(strings.ToLower(task.ID), strings.ToLower(selector))) ||
				domain.NormalizeName(task.Title) == needle

			if matched {
				return TaskRef{
					Task:          task,
					CategoryName:  project.Categories[cIdx].Name,
					CategoryIndex: cIdx,
					TaskIndex:     tIdx,
				}, nil
			}
		}
	}

	return TaskRef{}, ErrNotFound
}

func resolveCategory(project domain.Project, selector string) (*domain.Category, int, error) {
	selector = strings.TrimSpace(selector)
	if selector == "" {
		return nil, -1, ErrNotFound
	}

	needle := domain.NormalizeName(selector)
	minPrefixLen := 4

	for cIdx := range project.Categories {
		cat := &project.Categories[cIdx]

		if cat.ID == selector {
			return cat, cIdx, nil
		}

		if len(selector) >= minPrefixLen && strings.HasPrefix(strings.ToLower(cat.ID), strings.ToLower(selector)) {
			return cat, cIdx, nil
		}

		if domain.NormalizeName(cat.Name) == needle {
			return cat, cIdx, nil
		}
	}

	return nil, -1, ErrNotFound
}
