package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"phasionary/internal/domain"
	"phasionary/internal/operations"
)

// writeJSON serializes v as the response body with the given status code.
// Encoding failures are logged, not surfaced, since the header is already sent.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("api json encode error: %v", err)
	}
}

// writeJSONError writes a consistent {"error": msg} body with the given status.
func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// maxAPIBodyBytes caps request bodies so a malformed or hostile client can't
// force the server to buffer an unbounded payload.
const maxAPIBodyBytes = 1 << 20 // 1 MiB

// decodeJSONBody decodes a size-limited JSON request body into v. It returns a
// client-friendly message rather than leaking the raw decoder error.
func decodeJSONBody(w http.ResponseWriter, r *http.Request, v any) error {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxAPIBodyBytes))
	if err := dec.Decode(v); err != nil {
		return errors.New("invalid JSON body")
	}
	return nil
}

// apiError renders err as JSON using the shared errorStatus classifier, which
// maps not-found errors to 404 and anything unexpected to 500. Only unexpected
// 500s are logged; expected 404s are not.
func (s *Server) apiError(w http.ResponseWriter, err error) {
	status, msg := errorStatus(err)
	if status == http.StatusInternalServerError {
		log.Printf("api internal error: %v", err)
	}
	writeJSONError(w, status, msg)
}

func (s *Server) handleAPIProjectsList(w http.ResponseWriter, r *http.Request) {
	projects, err := s.store.ListProjects()
	if err != nil {
		s.apiError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, projects)
}

func (s *Server) handleAPIProjectGet(w http.ResponseWriter, r *http.Request) {
	project, err := s.store.LoadProjectByID(r.PathValue("pid"))
	if err != nil {
		s.apiError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, project)
}

// taskCreateRequest is the quick-capture payload. Only title is required; the
// field names mirror the domain.Task JSON tags so clients use one vocabulary.
type taskCreateRequest struct {
	Title           string `json:"title"`
	Status          string `json:"status"`
	Priority        string `json:"priority"`
	EstimateMinutes int    `json:"estimate_minutes"`
	Description     string `json:"description"`
}

func (s *Server) handleAPITaskCreate(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("pid")
	cid := r.PathValue("cid")

	var req taskCreateRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	status := req.Status
	if status == "" {
		status = domain.StatusTodo
	}
	fields := operations.TaskFields{
		Title:       strings.TrimSpace(req.Title),
		Status:      status,
		Priority:    req.Priority,
		Estimate:    req.EstimateMinutes,
		Description: strings.TrimRight(req.Description, "\n"),
	}
	if err := validateTaskFields(fields); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if fields.Estimate < 0 {
		writeJSONError(w, http.StatusBadRequest, "estimate_minutes must be zero or greater")
		return
	}

	var created domain.Task
	_, err := s.withProject(pid, func(p *domain.Project) error {
		var e error
		created, e = operations.CreateTask(p, cid, fields)
		return e
	})
	if err != nil {
		s.apiError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// taskStatusRequest sets a task's status explicitly (unlike the HTML endpoint,
// which cycles). Explicit set keeps the API idempotent and predictable.
type taskStatusRequest struct {
	Status string `json:"status"`
}

func (s *Server) handleAPITaskStatus(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("pid")
	cid := r.PathValue("cid")
	tid := r.PathValue("tid")

	var req taskStatusRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	// Validate up-front so an invalid status is a 400 without touching disk.
	if err := domain.ValidateStatus(req.Status); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid status")
		return
	}

	var updated domain.Task
	_, err := s.withProject(pid, func(p *domain.Project) error {
		var e error
		updated, e = operations.SetTaskStatus(p, cid, tid, req.Status)
		return e
	})
	if err != nil {
		s.apiError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}
