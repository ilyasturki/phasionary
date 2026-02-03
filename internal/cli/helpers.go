package cli

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"phasionary/internal/domain"
)

var ErrNotFound = errors.New("not found")

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

func resolveTask(project domain.Project, selector string) (*domain.Task, string, int, int, error) {
	selector = strings.TrimSpace(selector)
	if selector == "" {
		return nil, "", -1, -1, ErrNotFound
	}

	needle := domain.NormalizeName(selector)
	minPrefixLen := 4

	for cIdx := range project.Categories {
		for tIdx := range project.Categories[cIdx].Tasks {
			task := &project.Categories[cIdx].Tasks[tIdx]

			if task.ID == selector {
				return task, project.Categories[cIdx].Name, cIdx, tIdx, nil
			}

			if len(selector) >= minPrefixLen && strings.HasPrefix(strings.ToLower(task.ID), strings.ToLower(selector)) {
				return task, project.Categories[cIdx].Name, cIdx, tIdx, nil
			}

			if domain.NormalizeName(task.Title) == needle {
				return task, project.Categories[cIdx].Name, cIdx, tIdx, nil
			}
		}
	}

	return nil, "", -1, -1, ErrNotFound
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
