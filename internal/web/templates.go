package web

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"phasionary/internal/domain"
)

func (s *Server) parseTemplates() error {
	funcs := template.FuncMap{
		"statusLabel":    statusLabel,
		"statusSlug":     statusSlug,
		"statusBadge":    statusBadge,
		"priorityLabel":  priorityLabel,
		"prioritySlug":   prioritySlug,
		"priorityBadge":  priorityBadge,
		"formatEstimate": formatEstimate,
		"countsLine":     countsLine,
		"dict":           dict,
	}

	pages := map[string][]string{
		"projects": {
			"templates/base.html",
			"templates/pages/projects.html",
		},
		"project": {
			"templates/base.html",
			"templates/pages/project.html",
			"templates/partials/categories.html",
			"templates/partials/category.html",
			"templates/partials/task_row.html",
		},
		"project_form": {
			"templates/base.html",
			"templates/forms/project_form.html",
		},
		"category_form": {
			"templates/base.html",
			"templates/forms/category_form.html",
		},
		"task_form": {
			"templates/base.html",
			"templates/forms/task_form.html",
		},
	}

	s.pages = make(map[string]*template.Template, len(pages))
	for name, files := range pages {
		t, err := template.New("base").Funcs(funcs).ParseFS(templatesFS, files...)
		if err != nil {
			return fmt.Errorf("parse %s: %w", name, err)
		}
		s.pages[name] = t
	}

	partials := map[string][]string{
		"task_row":   {"templates/partials/task_row.html"},
		"category":   {"templates/partials/category.html", "templates/partials/task_row.html"},
		"categories": {"templates/partials/categories.html", "templates/partials/category.html", "templates/partials/task_row.html"},
	}
	s.partials = make(map[string]*template.Template, len(partials))
	for name, files := range partials {
		t, err := template.New("partial").Funcs(funcs).ParseFS(templatesFS, files...)
		if err != nil {
			return fmt.Errorf("parse partial %s: %w", name, err)
		}
		s.partials[name] = t
	}

	return nil
}

func (s *Server) render(w http.ResponseWriter, name string, data any) {
	tmpl, ok := s.pages[name]
	if !ok {
		http.Error(w, "template not found: "+name, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		log.Printf("template render error (%s): %v", name, err)
	}
}

func (s *Server) renderPartial(w http.ResponseWriter, name string, data any) {
	tmpl, ok := s.partials[name]
	if !ok {
		http.Error(w, "partial not found: "+name, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("partial render error (%s): %v", name, err)
	}
}

type projectsPageData struct {
	Title    string
	Projects []domain.Project
}

type projectPageData struct {
	Title   string
	Project domain.Project
}

type projectFormData struct {
	Title   string
	Project domain.Project
	IsNew   bool
	Error   string
}

type categoryFormData struct {
	Title    string
	Project  domain.Project
	Category domain.Category
	IsNew    bool
	Error    string
}

type taskFormData struct {
	Title    string
	Project  domain.Project
	Category domain.Category
	Task     domain.Task
	IsNew    bool
	Error    string
}

type taskRowData struct {
	Project  domain.Project
	Category domain.Category
	Task     domain.Task
}

type categoryViewData struct {
	Project  domain.Project
	Category domain.Category
}

type categoriesViewData struct {
	Project    domain.Project
	Categories []domain.Category
}

func dict(values ...any) (map[string]any, error) {
	if len(values)%2 != 0 {
		return nil, errors.New("dict: odd number of values")
	}
	m := make(map[string]any, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict: key at index %d is not a string", i)
		}
		m[key] = values[i+1]
	}
	return m, nil
}

func statusLabel(s string) string {
	switch s {
	case domain.StatusTodo:
		return "Todo"
	case domain.StatusInProgress:
		return "In progress"
	case domain.StatusCompleted:
		return "Done"
	case domain.StatusCancelled:
		return "Cancelled"
	}
	return s
}

func statusSlug(s string) string {
	return strings.ReplaceAll(s, "_", "-")
}

func priorityLabel(p string) string {
	switch p {
	case domain.PriorityHigh:
		return "High"
	case domain.PriorityMedium:
		return "Med"
	case domain.PriorityLow:
		return "Low"
	}
	return "—"
}

func prioritySlug(p string) string {
	if p == "" {
		return "none"
	}
	return p
}

// statusBadge returns the CSS classes for the status toggle button: a
// basecoat .btn-sm with a status-specific color override defined in style.css.
func statusBadge(s string) string {
	return "btn-sm status-" + statusSlug(s)
}

// priorityBadge returns the CSS classes for the priority toggle button.
// Returns "" for empty priority — the template renders a muted em-dash
// placeholder instead.
func priorityBadge(p string) string {
	if p == "" {
		return ""
	}
	return "btn-sm priority-" + p
}

func formatEstimate(m int) string {
	if m == 0 {
		return ""
	}
	if m < 60 {
		return strconv.Itoa(m) + "m"
	}
	h := m / 60
	rem := m % 60
	if rem == 0 {
		return strconv.Itoa(h) + "h"
	}
	return fmt.Sprintf("%dh%dm", h, rem)
}

func countsLine(c domain.StatusCounts) string {
	var parts []string
	if c.InProgress > 0 {
		parts = append(parts, fmt.Sprintf("%d in progress", c.InProgress))
	}
	if c.Todo > 0 {
		parts = append(parts, fmt.Sprintf("%d todo", c.Todo))
	}
	if c.Completed > 0 {
		parts = append(parts, fmt.Sprintf("%d done", c.Completed))
	}
	if c.Cancelled > 0 {
		parts = append(parts, fmt.Sprintf("%d cancelled", c.Cancelled))
	}
	if len(parts) == 0 {
		return "no tasks"
	}
	return strings.Join(parts, " · ")
}
