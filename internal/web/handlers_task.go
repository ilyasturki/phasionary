package web

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"phasionary/internal/domain"
)

// parseEstimate accepts a non-negative minute count or empty (= 0). Reject
// negatives and non-numerics so a tampered <input min="0"> can't poison the
// stored value.
func parseEstimate(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, errors.New("estimate must be a whole number of minutes")
	}
	if n < 0 {
		return 0, errors.New("estimate must be zero or greater")
	}
	return n, nil
}

// taskFields carries the writable fields of a task. It is the shared unit the
// HTML form path and the JSON API path both build and validate, so the two
// surfaces create tasks through identical logic and can't drift.
type taskFields struct {
	Title       string
	Status      string
	Priority    string
	Estimate    int
	Description string
}

// Sentinel validation errors, so each surface maps them to its own presentation
// (an HTML form message vs. a JSON error body).
var (
	errTaskTitleRequired = errors.New("title is required")
	errTaskBadStatus     = errors.New("invalid status")
	errTaskBadPriority   = errors.New("invalid priority")
)

// validateTaskFields checks the title and the enum-like status/priority fields.
// Status must already be defaulted by the caller (empty is invalid). Estimate is
// validated by the caller — the form path parses it from a string, the JSON path
// receives an int — so it is not re-checked here.
func validateTaskFields(f taskFields) error {
	if f.Title == "" {
		return errTaskTitleRequired
	}
	if domain.ValidateStatus(f.Status) != nil {
		return errTaskBadStatus
	}
	if domain.ValidatePriority(f.Priority) != nil {
		return errTaskBadPriority
	}
	return nil
}

// findTask resolves cid→category and tid→task within p, returning a pointer to
// the stored task so callers can mutate it in place under the project lock.
func findTask(p *domain.Project, cid, tid string) (*domain.Task, error) {
	cidx, err := p.FindCategoryByID(cid)
	if err != nil {
		return nil, err
	}
	tidx, err := p.Categories[cidx].FindTaskByID(tid)
	if err != nil {
		return nil, err
	}
	return &p.Categories[cidx].Tasks[tidx], nil
}

// applyNewTask builds a task from f and appends it to category cid in p. It
// assumes f has already passed validateTaskFields. Returns the created task
// (identical to the stored copy) so callers can echo it back.
func applyNewTask(p *domain.Project, cid string, f taskFields) (domain.Task, error) {
	idx, err := p.FindCategoryByID(cid)
	if err != nil {
		return domain.Task{}, err
	}
	task, err := domain.NewTask(f.Title)
	if err != nil {
		return domain.Task{}, err
	}
	_ = task.SetStatus(f.Status)
	_ = task.SetPriority(f.Priority)
	task.SetEstimate(f.Estimate)
	task.Description = f.Description
	p.Categories[idx].AddTask(task)
	return task, nil
}

func (s *Server) handleTaskNew(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("pid")
	cid := r.PathValue("cid")
	project, err := s.store.LoadProjectByID(pid)
	if err != nil {
		s.mutationError(w, err)
		return
	}
	idx, err := project.FindCategoryByID(cid)
	if err != nil {
		s.mutationError(w, err)
		return
	}
	s.render(w, "task_form", taskFormData{
		Title:    "New task",
		Project:  project,
		Category: project.Categories[idx],
		Task:     domain.Task{Status: domain.StatusTodo},
		IsNew:    true,
	})
}

func (s *Server) handleTaskCreate(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("pid")
	cid := r.PathValue("cid")
	if err := r.ParseForm(); err != nil {
		s.badRequest(w, "invalid form")
		return
	}
	title := strings.TrimSpace(r.FormValue("title"))
	status := r.FormValue("status")
	if status == "" {
		status = domain.StatusTodo
	}
	priority := r.FormValue("priority")
	estimate, estErr := parseEstimate(r.FormValue("estimate"))
	description := strings.TrimRight(r.FormValue("description"), "\n")

	renderForm := func(errMsg string) {
		project, err := s.store.LoadProjectByID(pid)
		if err != nil {
			s.mutationError(w, err)
			return
		}
		idx, err := project.FindCategoryByID(cid)
		if err != nil {
			s.mutationError(w, err)
			return
		}
		s.render(w, "task_form", taskFormData{
			Title:    "New task",
			Project:  project,
			Category: project.Categories[idx],
			Task: domain.Task{
				Title:           title,
				Status:          status,
				Priority:        priority,
				EstimateMinutes: estimate,
				Description:     description,
			},
			IsNew: true,
			Error: errMsg,
		})
	}

	fields := taskFields{
		Title:       title,
		Status:      status,
		Priority:    priority,
		Estimate:    estimate,
		Description: description,
	}
	// Validate enum-like fields up-front so a tampered <select> shows a form
	// error rather than a 500.
	if err := validateTaskFields(fields); err != nil {
		switch {
		case errors.Is(err, errTaskTitleRequired):
			renderForm("Title is required.")
		case errors.Is(err, errTaskBadStatus):
			renderForm("Invalid status.")
		case errors.Is(err, errTaskBadPriority):
			renderForm("Invalid priority.")
		default:
			renderForm(err.Error())
		}
		return
	}
	if estErr != nil {
		renderForm(estErr.Error())
		return
	}

	_, err := s.withProject(pid, func(p *domain.Project) error {
		_, err := applyNewTask(p, cid, fields)
		return err
	})
	if err != nil {
		s.mutationError(w, err)
		return
	}
	http.Redirect(w, r, "/projects/"+pid, http.StatusSeeOther)
}

func (s *Server) handleTaskEdit(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("pid")
	cid := r.PathValue("cid")
	tid := r.PathValue("tid")
	project, err := s.store.LoadProjectByID(pid)
	if err != nil {
		s.mutationError(w, err)
		return
	}
	cidx, err := project.FindCategoryByID(cid)
	if err != nil {
		s.mutationError(w, err)
		return
	}
	tidx, err := project.Categories[cidx].FindTaskByID(tid)
	if err != nil {
		s.mutationError(w, err)
		return
	}
	s.render(w, "task_form", taskFormData{
		Title:    "Edit task",
		Project:  project,
		Category: project.Categories[cidx],
		Task:     project.Categories[cidx].Tasks[tidx],
		IsNew:    false,
	})
}

func (s *Server) handleTaskUpdate(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("pid")
	cid := r.PathValue("cid")
	tid := r.PathValue("tid")
	if err := r.ParseForm(); err != nil {
		s.badRequest(w, "invalid form")
		return
	}
	title := strings.TrimSpace(r.FormValue("title"))
	status := r.FormValue("status")
	priority := r.FormValue("priority")
	estimate, estErr := parseEstimate(r.FormValue("estimate"))
	description := strings.TrimRight(r.FormValue("description"), "\n")

	renderForm := func(errMsg string) {
		project, err := s.store.LoadProjectByID(pid)
		if err != nil {
			s.mutationError(w, err)
			return
		}
		cidx, err := project.FindCategoryByID(cid)
		if err != nil {
			s.mutationError(w, err)
			return
		}
		// Preserve the user's just-typed values; carry only ID over from the
		// persisted task so the form action URL still points at the right row.
		task := domain.Task{
			ID:              tid,
			Title:           title,
			Status:          status,
			Priority:        priority,
			EstimateMinutes: estimate,
			Description:     description,
		}
		s.render(w, "task_form", taskFormData{
			Title:    "Edit task",
			Project:  project,
			Category: project.Categories[cidx],
			Task:     task,
			IsNew:    false,
			Error:    errMsg,
		})
	}

	if title == "" {
		renderForm("Title is required.")
		return
	}
	probe := domain.Task{}
	if err := probe.SetStatus(status); err != nil {
		renderForm("Invalid status.")
		return
	}
	if err := probe.SetPriority(priority); err != nil {
		renderForm("Invalid priority.")
		return
	}
	if estErr != nil {
		renderForm(estErr.Error())
		return
	}

	_, err := s.withProject(pid, func(p *domain.Project) error {
		cidx, err := p.FindCategoryByID(cid)
		if err != nil {
			return err
		}
		tidx, err := p.Categories[cidx].FindTaskByID(tid)
		if err != nil {
			return err
		}
		task := &p.Categories[cidx].Tasks[tidx]
		task.Title = title
		_ = task.SetStatus(status)
		_ = task.SetPriority(priority)
		task.SetEstimate(estimate)
		task.Description = description
		return nil
	})
	if err != nil {
		s.mutationError(w, err)
		return
	}
	http.Redirect(w, r, "/projects/"+pid, http.StatusSeeOther)
}

func (s *Server) handleTaskDelete(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("pid")
	cid := r.PathValue("cid")
	tid := r.PathValue("tid")
	_, err := s.withProject(pid, func(p *domain.Project) error {
		cidx, err := p.FindCategoryByID(cid)
		if err != nil {
			return err
		}
		tidx, err := p.Categories[cidx].FindTaskByID(tid)
		if err != nil {
			return err
		}
		return p.Categories[cidx].RemoveTask(tidx)
	})
	if err != nil {
		s.mutationError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleTaskStatus(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("pid")
	cid := r.PathValue("cid")
	tid := r.PathValue("tid")
	var cidx, tidx int
	project, err := s.withProject(pid, func(p *domain.Project) error {
		var err error
		cidx, err = p.FindCategoryByID(cid)
		if err != nil {
			return err
		}
		tidx, err = p.Categories[cidx].FindTaskByID(tid)
		if err != nil {
			return err
		}
		p.Categories[cidx].Tasks[tidx].CycleStatus()
		return nil
	})
	if err != nil {
		s.mutationError(w, err)
		return
	}
	s.renderPartial(w, "task_row", taskRowData{
		Project:  project,
		Category: project.Categories[cidx],
		Task:     project.Categories[cidx].Tasks[tidx],
	})
}

func (s *Server) handleTaskPriority(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("pid")
	cid := r.PathValue("cid")
	tid := r.PathValue("tid")
	var cidx, tidx int
	project, err := s.withProject(pid, func(p *domain.Project) error {
		var err error
		cidx, err = p.FindCategoryByID(cid)
		if err != nil {
			return err
		}
		tidx, err = p.Categories[cidx].FindTaskByID(tid)
		if err != nil {
			return err
		}
		p.Categories[cidx].Tasks[tidx].CyclePriority()
		return nil
	})
	if err != nil {
		s.mutationError(w, err)
		return
	}
	s.renderPartial(w, "task_row", taskRowData{
		Project:  project,
		Category: project.Categories[cidx],
		Task:     project.Categories[cidx].Tasks[tidx],
	})
}

func (s *Server) handleTaskMove(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("pid")
	cid := r.PathValue("cid")
	tid := r.PathValue("tid")
	delta, err := parseMoveDir(r)
	if err != nil {
		s.badRequest(w, err.Error())
		return
	}
	project, err := s.withProject(pid, func(p *domain.Project) error {
		cidx, err := p.FindCategoryByID(cid)
		if err != nil {
			return err
		}
		tidx, err := p.Categories[cidx].FindTaskByID(tid)
		if err != nil {
			return err
		}
		// Swallow edge errors so tapping past the boundary doesn't error,
		// and skip the save so we don't bump UpdatedAt for nothing.
		if mvErr := p.Categories[cidx].MoveTask(tidx, delta); mvErr != nil {
			return errNoChange
		}
		return nil
	})
	if err != nil {
		s.mutationError(w, err)
		return
	}
	// Re-resolve the category index against the post-call project — the
	// errNoChange branch of withProject reloads from disk without the
	// project flock, so any closure-captured cidx may be stale.
	cidx, err := project.FindCategoryByID(cid)
	if err != nil {
		s.mutationError(w, err)
		return
	}
	s.renderPartial(w, "category", categoryViewData{
		Project:  project,
		Category: project.Categories[cidx],
	})
}
