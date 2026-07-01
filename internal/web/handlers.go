package web

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"phasionary/internal/data"
	"phasionary/internal/domain"
)

// errNoChange is returned from a withProject closure to signal "the load was
// fine but nothing actually changed; skip the save." Lets edge-case no-ops
// (e.g. tap "move up" on the first item) avoid a disk write that would only
// bump UpdatedAt. If the closure mutated the project before returning
// errNoChange, withProject discards those mutations by reloading from disk.
var errNoChange = errors.New("no change")

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/projects", http.StatusSeeOther)
}

// withProject runs fn inside Store.WithProjectLocked. If fn returns
// errNoChange, the save is skipped and the project is reloaded from disk so
// the caller renders only persisted state, never speculative in-memory
// mutations the closure may have applied before bailing.
func (s *Server) withProject(pid string, fn func(*domain.Project) error) (domain.Project, error) {
	project, err := s.store.WithProjectLocked(pid, fn)
	if errors.Is(err, errNoChange) {
		return s.store.LoadProjectByID(pid)
	}
	return project, err
}

func (s *Server) notFound(w http.ResponseWriter, msg string) {
	http.Error(w, msg, http.StatusNotFound)
}

func (s *Server) badRequest(w http.ResponseWriter, msg string) {
	http.Error(w, msg, http.StatusBadRequest)
}

func (s *Server) internalError(w http.ResponseWriter, err error) {
	log.Printf("internal error: %v", err)
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

// errorStatus classifies a store/domain error into an HTTP status and a
// client-safe message. The HTML surface (mutationError) and the JSON surface
// (apiError) both go through it, so the two agree on which errors are 404 vs
// 500 without each maintaining its own copy of the mapping.
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

// mutationError renders err as an HTML error using the shared errorStatus
// classifier. Used by every mutation handler.
func (s *Server) mutationError(w http.ResponseWriter, err error) {
	status, msg := errorStatus(err)
	if status == http.StatusInternalServerError {
		s.internalError(w, err)
		return
	}
	http.Error(w, msg, status)
}

func parseMoveDir(r *http.Request) (int, error) {
	switch strings.ToLower(r.URL.Query().Get("dir")) {
	case "up":
		return -1, nil
	case "down":
		return 1, nil
	}
	return 0, errors.New("invalid dir; expected up or down")
}
