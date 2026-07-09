package api

import (
	"errors"
	"net/http"

	"phasionary/internal/data"
	"phasionary/internal/domain"
	"phasionary/internal/operations"
)

// withProject runs fn inside Store.WithProjectLocked. If fn returns
// operations.ErrNoChange, the save is skipped and the project is reloaded from
// disk so the caller renders only persisted state, never speculative in-memory
// mutations the verb may have applied before bailing.
func (s *Server) withProject(pid string, fn func(*domain.Project) error) (domain.Project, error) {
	project, err := s.store.WithProjectLocked(pid, fn)
	if errors.Is(err, operations.ErrNoChange) {
		return s.store.LoadProjectByID(pid)
	}
	return project, err
}

// errorStatus classifies a store/domain error into an HTTP status and a
// client-safe message. apiError goes through it so a not-found error maps to a
// 404 while anything unexpected maps to a 500.
func errorStatus(err error) (int, string) {
	switch {
	case errors.Is(err, data.ErrProjectNotFound):
		return http.StatusNotFound, "project not found"
	case errors.Is(err, domain.ErrCategoryNotFound):
		return http.StatusNotFound, "category not found"
	case errors.Is(err, domain.ErrTaskNotFound):
		return http.StatusNotFound, "task not found"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

// Sentinel validation errors for the task quick-capture payload; each maps to a
// distinct client-facing message in the 400 response.
var (
	errTaskTitleRequired = errors.New("title is required")
	errTaskBadStatus     = errors.New("invalid status")
	errTaskBadPriority   = errors.New("invalid priority")
	errTaskBadTagColor   = errors.New("invalid tag color")
)

// validateTaskFields checks the title and the enum-like status/priority fields.
// Status must already be defaulted by the caller (empty is invalid). Estimate is
// validated separately by the caller.
func validateTaskFields(f operations.TaskFields) error {
	if f.Title == "" {
		return errTaskTitleRequired
	}
	if domain.ValidateStatus(f.Status) != nil {
		return errTaskBadStatus
	}
	if domain.ValidatePriority(f.Priority) != nil {
		return errTaskBadPriority
	}
	if domain.ValidateTagColor(f.TagColor) != nil {
		return errTaskBadTagColor
	}
	return nil
}
