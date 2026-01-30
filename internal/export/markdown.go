package export

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"

	"phasionary/internal/domain"
)

var (
	projectHeaderRe  = regexp.MustCompile(`^#\s+(.+)$`)
	categoryHeaderRe = regexp.MustCompile(`^##\s+(.+)$`)
	taskLineRe       = regexp.MustCompile(`^-\s+\[([ x\-~])\]\s+(.+)$`)
	prioritySuffixRe = regexp.MustCompile(`\s+\((high|medium|low)\)\s*$`)
)

func statusToMarker(status string) string {
	switch status {
	case domain.StatusCompleted:
		return "x"
	case domain.StatusCancelled:
		return "-"
	case domain.StatusInProgress:
		return "~"
	default:
		return " "
	}
}

func markerToStatus(marker string) string {
	switch marker {
	case "x":
		return domain.StatusCompleted
	case "-":
		return domain.StatusCancelled
	case "~":
		return domain.StatusInProgress
	default:
		return domain.StatusTodo
	}
}

func ExportMarkdown(project domain.Project, w io.Writer) error {
	if _, err := fmt.Fprintf(w, "# %s\n", project.Name); err != nil {
		return err
	}
	if project.Description != "" {
		if _, err := fmt.Fprintf(w, "\n%s\n", project.Description); err != nil {
			return err
		}
	}
	for _, cat := range project.Categories {
		if _, err := fmt.Fprintf(w, "\n## %s\n\n", cat.Name); err != nil {
			return err
		}
		for _, task := range cat.Tasks {
			marker := statusToMarker(task.Status)
			line := fmt.Sprintf("- [%s] %s", marker, task.Title)
			if task.Priority != "" {
				line += fmt.Sprintf(" (%s)", task.Priority)
			}
			if _, err := fmt.Fprintln(w, line); err != nil {
				return err
			}
		}
	}
	return nil
}

func ImportMarkdown(r io.Reader, projectName string) (domain.Project, error) {
	scanner := bufio.NewScanner(r)

	var parsedName string
	var descLines []string
	var categories []categoryData
	var currentCategory *categoryData
	inDescription := false

	for scanner.Scan() {
		line := scanner.Text()

		if m := projectHeaderRe.FindStringSubmatch(line); m != nil {
			parsedName = strings.TrimSpace(m[1])
			inDescription = true
			continue
		}

		if m := categoryHeaderRe.FindStringSubmatch(line); m != nil {
			if currentCategory != nil {
				categories = append(categories, *currentCategory)
			}
			currentCategory = &categoryData{name: strings.TrimSpace(m[1])}
			inDescription = false
			continue
		}

		if m := taskLineRe.FindStringSubmatch(line); m != nil && currentCategory != nil {
			marker := m[1]
			title := strings.TrimSpace(m[2])
			var priority string
			if pm := prioritySuffixRe.FindStringSubmatch(title); pm != nil {
				priority = pm[1]
				title = strings.TrimSpace(prioritySuffixRe.ReplaceAllString(title, ""))
			}
			currentCategory.tasks = append(currentCategory.tasks, taskData{
				title:    title,
				status:   markerToStatus(marker),
				priority: priority,
			})
			continue
		}

		if inDescription && currentCategory == nil {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				descLines = append(descLines, trimmed)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return domain.Project{}, err
	}
	if currentCategory != nil {
		categories = append(categories, *currentCategory)
	}

	name := projectName
	if name == "" {
		name = parsedName
	}
	if name == "" {
		name = "Imported Project"
	}

	project, err := domain.NewProject(name, strings.Join(descLines, " "))
	if err != nil {
		return domain.Project{}, err
	}

	for _, cd := range categories {
		cat, err := domain.NewCategory(cd.name)
		if err != nil {
			return domain.Project{}, err
		}
		for _, td := range cd.tasks {
			task, err := domain.NewTask(td.title)
			if err != nil {
				return domain.Project{}, err
			}
			if err := task.SetStatus(td.status); err != nil {
				return domain.Project{}, err
			}
			if td.priority != "" {
				if err := task.SetPriority(td.priority); err != nil {
					return domain.Project{}, err
				}
			}
			cat.Tasks = append(cat.Tasks, task)
		}
		project.Categories = append(project.Categories, cat)
	}

	return project, nil
}

type categoryData struct {
	name  string
	tasks []taskData
}

type taskData struct {
	title    string
	status   string
	priority string
}
