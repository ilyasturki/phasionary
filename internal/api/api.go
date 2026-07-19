package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sort"
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
	// Most recently touched first — the phone's picker is "what am I working
	// on", where the TUI's picker keeps a manual order (state.json
	// project_order). That's why the sort lives here and not in the store:
	// changing ListProjects would reorder the TUI too. UpdatedAt is UTC RFC3339,
	// so string comparison is chronological.
	sort.SliceStable(projects, func(i, j int) bool {
		if projects[i].UpdatedAt != projects[j].UpdatedAt {
			return projects[i].UpdatedAt > projects[j].UpdatedAt
		}
		return strings.ToLower(projects[i].Name) < strings.ToLower(projects[j].Name)
	})
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
	TagColor        string `json:"tag_color"`
	TagLabel        string `json:"tag_label"`
	// Kind is "" for an ordinary task or "separator" for a divider, whose only
	// field is its (optional) label in Title.
	Kind string `json:"kind"`
	// InsertAfter places the new row directly below that task; empty appends to
	// the end of the category. Separators are only useful between rows, so the
	// client always sends an anchor for them.
	InsertAfter string `json:"insert_after"`
}

func (s *Server) handleAPITaskCreate(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("pid")
	cid := r.PathValue("cid")

	var req taskCreateRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	// A separator carries only a label, so it skips the task defaults entirely —
	// giving it a default status is exactly the corruption the verbs reject.
	// The client's other fields are still forwarded verbatim so the verb can
	// reject them: dropping them here instead would silently accept a request
	// that asked for something impossible.
	var fields operations.TaskFields
	if req.Kind == domain.KindSeparator {
		fields = operations.TaskFields{
			Kind:        domain.KindSeparator,
			Title:       strings.TrimSpace(req.Title),
			Status:      req.Status,
			Priority:    req.Priority,
			Estimate:    req.EstimateMinutes,
			Description: req.Description,
			TagColor:    req.TagColor,
			TagLabel:    req.TagLabel,
		}
	} else {
		status := req.Status
		if status == "" {
			status = domain.StatusTodo
		}
		fields = operations.TaskFields{
			Title:       strings.TrimSpace(req.Title),
			Status:      status,
			Priority:    req.Priority,
			Estimate:    req.EstimateMinutes,
			Description: strings.TrimRight(req.Description, "\n"),
			TagColor:    req.TagColor,
			TagLabel:    req.TagLabel,
			Kind:        req.Kind,
		}
		if err := validateTaskFields(fields); err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		if fields.Estimate < 0 {
			writeJSONError(w, http.StatusBadRequest, "estimate_minutes must be zero or greater")
			return
		}
	}

	var created domain.Task
	_, err := s.withProject(pid, func(p *domain.Project) error {
		var e error
		created, e = operations.CreateTaskAfter(p, cid, fields, req.InsertAfter)
		return e
	})
	if err != nil {
		s.apiError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// taskUpdateRequest is the partial-edit payload for PATCH. Every field is a
// pointer so an omitted field is left untouched while an explicit "" or 0
// clears it — the mobile editor can send only what actually changed.
type taskUpdateRequest struct {
	Title           *string `json:"title"`
	Status          *string `json:"status"`
	Priority        *string `json:"priority"`
	EstimateMinutes *int    `json:"estimate_minutes"`
	Description     *string `json:"description"`
}

func (s *Server) handleAPITaskUpdate(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("pid")
	cid := r.PathValue("cid")
	tid := r.PathValue("tid")

	var req taskUpdateRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	// Validate the enum-like fields here so they 400 without touching disk;
	// domain's validators return plain errors that errorStatus can't classify.
	if req.Status != nil && domain.ValidateStatus(*req.Status) != nil {
		writeJSONError(w, http.StatusBadRequest, errTaskBadStatus.Error())
		return
	}
	if req.Priority != nil && domain.ValidatePriority(*req.Priority) != nil {
		writeJSONError(w, http.StatusBadRequest, errTaskBadPriority.Error())
		return
	}

	update := operations.TaskUpdate{
		Title:       req.Title,
		Status:      req.Status,
		Priority:    req.Priority,
		Estimate:    req.EstimateMinutes,
		Description: req.Description,
	}
	var updated domain.Task
	_, err := s.withProject(pid, func(p *domain.Project) error {
		var e error
		updated, e = operations.UpdateTask(p, cid, tid, update)
		return e
	})
	if err != nil {
		s.apiError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) handleAPITaskDelete(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("pid")
	cid := r.PathValue("cid")
	tid := r.PathValue("tid")

	_, err := s.withProject(pid, func(p *domain.Project) error {
		_, e := operations.DeleteTask(p, cid, tid)
		return e
	})
	if err != nil {
		s.apiError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// categoryCreateRequest is the payload for adding a category from the phone.
type categoryCreateRequest struct {
	Name string `json:"name"`
}

func (s *Server) handleAPICategoryCreate(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("pid")

	var req categoryCreateRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	var created domain.Category
	_, err := s.withProject(pid, func(p *domain.Project) error {
		var e error
		created, e = operations.CreateCategory(p, req.Name)
		return e
	})
	if err != nil {
		s.apiError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// foldsPayload is both the GET response and the PUT body. It mirrors the
// state.json key so the phone and the TUI speak one vocabulary about which
// categories are collapsed. PUT replaces the whole list, matching
// StateRepository.SetFoldedCategories.
type foldsPayload struct {
	FoldedCategories []string `json:"folded_categories"`
}

func (s *Server) handleAPIFoldsGet(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("pid")
	// Resolve the project first so an unknown id 404s instead of reporting an
	// empty fold set for a project that doesn't exist.
	if _, err := s.store.LoadProjectByID(pid); err != nil {
		s.apiError(w, err)
		return
	}
	folded := s.state.GetFoldedCategories(pid)
	if folded == nil {
		folded = []string{}
	}
	writeJSON(w, http.StatusOK, foldsPayload{FoldedCategories: folded})
}

func (s *Server) handleAPIFoldsSet(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("pid")
	if _, err := s.store.LoadProjectByID(pid); err != nil {
		s.apiError(w, err)
		return
	}

	var req foldsPayload
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	// Category IDs are not validated against the project: a fold entry for a
	// category deleted elsewhere is inert, and the TUI already tolerates it.
	if err := s.state.SetFoldedCategories(pid, req.FoldedCategories); err != nil {
		s.apiError(w, err)
		return
	}
	folded := req.FoldedCategories
	if folded == nil {
		folded = []string{}
	}
	writeJSON(w, http.StatusOK, foldsPayload{FoldedCategories: folded})
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
