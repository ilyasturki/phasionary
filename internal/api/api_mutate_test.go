package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"phasionary/internal/data"
	"phasionary/internal/domain"
)

// seedTask creates a project with one extra task and returns the ids needed to
// address it: project, category, task.
func seedTask(t *testing.T, srv *Server, store *data.Store, title string) (string, string, string) {
	t.Helper()
	p, err := store.CreateProject("Test")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	pid, cid := p.ID, p.Categories[0].ID
	resp := do(t, srv, newJSONRequest(t, "POST",
		"/api/v1/projects/"+pid+"/categories/"+cid+"/tasks",
		map[string]any{"title": title}))
	assertStatus(t, resp, http.StatusCreated)
	var created domain.Task
	decodeJSON(t, resp, &created)
	return pid, cid, created.ID
}

func taskPath(pid, cid, tid string) string {
	return "/api/v1/projects/" + pid + "/categories/" + cid + "/tasks/" + tid
}

// writeProjectFixture writes a project file directly. The store stamps its own
// UpdatedAt on every save, so this is the only way a test can pin the timestamps
// the sort is meant to order by.
func writeProjectFixture(t *testing.T, store *data.Store, id, name, updatedAt string) {
	t.Helper()
	raw, err := json.Marshal(domain.Project{
		ID:        id,
		Name:      name,
		CreatedAt: updatedAt,
		UpdatedAt: updatedAt,
	})
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	path := filepath.Join(store.Dir, id+".json")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}

func TestAPIProjectsListSortedByUpdatedAtDesc(t *testing.T) {
	srv, store := newTestServer(t, "")
	// Names run counter to recency, so a lingering alphabetical sort (what the
	// store itself returns) produces the opposite order and fails the test.
	writeProjectFixture(t, store, "p-alpha", "Alpha", "2026-01-01T00:00:00Z")
	writeProjectFixture(t, store, "p-zulu", "Zulu", "2026-06-01T00:00:00Z")

	resp := do(t, srv, newRequest(t, "GET", "/api/v1/projects", nil))
	assertStatus(t, resp, http.StatusOK)
	var projects []domain.Project
	decodeJSON(t, resp, &projects)
	if len(projects) != 2 {
		t.Fatalf("want 2 projects, got %d", len(projects))
	}
	if projects[0].ID != "p-zulu" || projects[1].ID != "p-alpha" {
		t.Fatalf("want [p-zulu p-alpha], got [%s %s]", projects[0].ID, projects[1].ID)
	}
}

func TestAPIProjectsListTiesBreakByName(t *testing.T) {
	srv, store := newTestServer(t, "")
	// Second-resolution timestamps collide often enough that the tie-break is
	// the common path, not an edge case.
	const same = "2026-03-01T12:00:00Z"
	writeProjectFixture(t, store, "p-zulu", "Zulu", same)
	writeProjectFixture(t, store, "p-alpha", "Alpha", same)

	resp := do(t, srv, newRequest(t, "GET", "/api/v1/projects", nil))
	assertStatus(t, resp, http.StatusOK)
	var projects []domain.Project
	decodeJSON(t, resp, &projects)
	if len(projects) != 2 {
		t.Fatalf("want 2 projects, got %d", len(projects))
	}
	if projects[0].Name != "Alpha" || projects[1].Name != "Zulu" {
		t.Fatalf("want [Alpha Zulu], got [%s %s]", projects[0].Name, projects[1].Name)
	}
}

func TestAPITaskUpdatePartial(t *testing.T) {
	srv, store := newTestServer(t, "")
	pid, cid, tid := seedTask(t, srv, store, "Original")

	resp := do(t, srv, newJSONRequest(t, "PATCH", taskPath(pid, cid, tid),
		map[string]any{
			"title":            "Renamed",
			"status":           domain.StatusInProgress,
			"priority":         domain.PriorityHigh,
			"estimate_minutes": 45,
			"description":      "notes",
		}))
	assertStatus(t, resp, http.StatusOK)
	assertJSONContentType(t, resp)

	var got domain.Task
	decodeJSON(t, resp, &got)
	if got.Title != "Renamed" || got.Status != domain.StatusInProgress ||
		got.Priority != domain.PriorityHigh || got.EstimateMinutes != 45 ||
		got.Description != "notes" {
		t.Fatalf("updated task fields wrong: %+v", got)
	}

	stored := findTaskInStore(t, store, pid, tid)
	if stored == nil || stored.Title != "Renamed" || stored.Description != "notes" {
		t.Fatalf("update not persisted: %+v", stored)
	}
}

func TestAPITaskUpdateOmittedFieldsUntouched(t *testing.T) {
	srv, store := newTestServer(t, "")
	pid, cid, tid := seedTask(t, srv, store, "Keep me")

	// Seed a description, then send a PATCH that mentions only the status.
	resp := do(t, srv, newJSONRequest(t, "PATCH", taskPath(pid, cid, tid),
		map[string]any{"description": "keep this"}))
	assertStatus(t, resp, http.StatusOK)

	resp = do(t, srv, newJSONRequest(t, "PATCH", taskPath(pid, cid, tid),
		map[string]any{"status": domain.StatusCompleted}))
	assertStatus(t, resp, http.StatusOK)

	stored := findTaskInStore(t, store, pid, tid)
	if stored == nil {
		t.Fatal("task vanished")
	}
	if stored.Title != "Keep me" {
		t.Fatalf("omitted title was modified: %q", stored.Title)
	}
	if stored.Description != "keep this" {
		t.Fatalf("omitted description was modified: %q", stored.Description)
	}
	if stored.Status != domain.StatusCompleted {
		t.Fatalf("status not applied: %q", stored.Status)
	}
}

func TestAPITaskUpdateExplicitEmptyClearsDescription(t *testing.T) {
	srv, store := newTestServer(t, "")
	pid, cid, tid := seedTask(t, srv, store, "Task")

	resp := do(t, srv, newJSONRequest(t, "PATCH", taskPath(pid, cid, tid),
		map[string]any{"description": "temporary"}))
	assertStatus(t, resp, http.StatusOK)

	// An explicit "" must clear it — that's the distinction the pointer fields buy.
	resp = do(t, srv, newJSONRequest(t, "PATCH", taskPath(pid, cid, tid),
		map[string]any{"description": ""}))
	assertStatus(t, resp, http.StatusOK)

	stored := findTaskInStore(t, store, pid, tid)
	if stored == nil || stored.Description != "" {
		t.Fatalf("description not cleared: %+v", stored)
	}
}

func TestAPITaskUpdateRejectsBadInput(t *testing.T) {
	srv, store := newTestServer(t, "")
	pid, cid, tid := seedTask(t, srv, store, "Task")

	cases := []struct {
		name string
		body map[string]any
	}{
		{"blank title", map[string]any{"title": "   "}},
		{"bad status", map[string]any{"status": "nope"}},
		{"bad priority", map[string]any{"priority": "urgentish"}},
		{"negative estimate", map[string]any{"estimate_minutes": -5}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := do(t, srv, newJSONRequest(t, "PATCH", taskPath(pid, cid, tid), tc.body))
			assertStatus(t, resp, http.StatusBadRequest)
			assertJSONError(t, resp)
		})
	}

	// The rejected writes must not have altered the task.
	stored := findTaskInStore(t, store, pid, tid)
	if stored == nil || stored.Title != "Task" || stored.Status != domain.StatusTodo {
		t.Fatalf("rejected update leaked through: %+v", stored)
	}
}

func TestAPITaskUpdateNotFound(t *testing.T) {
	srv, store := newTestServer(t, "")
	pid, cid, _ := seedTask(t, srv, store, "Task")

	resp := do(t, srv, newJSONRequest(t, "PATCH", taskPath(pid, cid, "no-such-task"),
		map[string]any{"title": "x"}))
	assertStatus(t, resp, http.StatusNotFound)
	assertJSONError(t, resp)
}

func TestAPITaskDelete(t *testing.T) {
	srv, store := newTestServer(t, "")
	pid, cid, tid := seedTask(t, srv, store, "Doomed")

	resp := do(t, srv, newRequest(t, "DELETE", taskPath(pid, cid, tid), nil))
	assertStatus(t, resp, http.StatusNoContent)

	if stored := findTaskInStore(t, store, pid, tid); stored != nil {
		t.Fatalf("task still present after delete: %+v", stored)
	}
}

func TestAPITaskDeleteNotFound(t *testing.T) {
	srv, store := newTestServer(t, "")
	pid, cid, _ := seedTask(t, srv, store, "Task")

	resp := do(t, srv, newRequest(t, "DELETE", taskPath(pid, cid, "ghost"), nil))
	assertStatus(t, resp, http.StatusNotFound)
	assertJSONError(t, resp)
}

func TestAPICategoryCreate(t *testing.T) {
	srv, store := newTestServer(t, "")
	p, _ := store.CreateProject("Test")

	resp := do(t, srv, newJSONRequest(t, "POST",
		"/api/v1/projects/"+p.ID+"/categories", map[string]any{"name": "Backlog"}))
	assertStatus(t, resp, http.StatusCreated)
	assertJSONContentType(t, resp)

	var got domain.Category
	decodeJSON(t, resp, &got)
	if got.ID == "" || got.Name != "Backlog" {
		t.Fatalf("created category wrong: %+v", got)
	}

	reloaded, err := store.LoadProjectByID(p.ID)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	var found bool
	for _, c := range reloaded.Categories {
		if c.ID == got.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("category not persisted")
	}
}

func TestAPICategoryCreateRejectsBlankAndDuplicate(t *testing.T) {
	srv, store := newTestServer(t, "")
	p, _ := store.CreateProject("Test")
	existing := p.Categories[0].Name

	resp := do(t, srv, newJSONRequest(t, "POST",
		"/api/v1/projects/"+p.ID+"/categories", map[string]any{"name": "  "}))
	assertStatus(t, resp, http.StatusBadRequest)
	assertJSONError(t, resp)

	resp = do(t, srv, newJSONRequest(t, "POST",
		"/api/v1/projects/"+p.ID+"/categories", map[string]any{"name": existing}))
	assertStatus(t, resp, http.StatusConflict)
	assertJSONError(t, resp)
}

func TestAPIFoldsRoundTrip(t *testing.T) {
	srv, store := newTestServer(t, "")
	p, _ := store.CreateProject("Test")
	cid := p.Categories[0].ID

	// Nothing folded yet: an empty array, never null.
	resp := do(t, srv, newRequest(t, "GET", "/api/v1/projects/"+p.ID+"/folds", nil))
	assertStatus(t, resp, http.StatusOK)
	var got foldsPayload
	decodeJSON(t, resp, &got)
	if len(got.FoldedCategories) != 0 {
		t.Fatalf("want no folds, got %v", got.FoldedCategories)
	}

	resp = do(t, srv, newJSONRequest(t, "PUT", "/api/v1/projects/"+p.ID+"/folds",
		foldsPayload{FoldedCategories: []string{cid}}))
	assertStatus(t, resp, http.StatusOK)

	resp = do(t, srv, newRequest(t, "GET", "/api/v1/projects/"+p.ID+"/folds", nil))
	assertStatus(t, resp, http.StatusOK)
	decodeJSON(t, resp, &got)
	if len(got.FoldedCategories) != 1 || got.FoldedCategories[0] != cid {
		t.Fatalf("folds not persisted: %v", got.FoldedCategories)
	}

	// An empty list unfolds everything (matching SetFoldedCategories).
	resp = do(t, srv, newJSONRequest(t, "PUT", "/api/v1/projects/"+p.ID+"/folds",
		foldsPayload{FoldedCategories: []string{}}))
	assertStatus(t, resp, http.StatusOK)

	resp = do(t, srv, newRequest(t, "GET", "/api/v1/projects/"+p.ID+"/folds", nil))
	decodeJSON(t, resp, &got)
	if len(got.FoldedCategories) != 0 {
		t.Fatalf("want folds cleared, got %v", got.FoldedCategories)
	}
}

// seedSeparator inserts a separator below the given anchor and returns its id.
func seedSeparator(t *testing.T, srv *Server, pid, cid, afterTaskID, label string) string {
	t.Helper()
	resp := do(t, srv, newJSONRequest(t, "POST",
		"/api/v1/projects/"+pid+"/categories/"+cid+"/tasks",
		map[string]any{
			"kind":         domain.KindSeparator,
			"title":        label,
			"insert_after": afterTaskID,
		}))
	assertStatus(t, resp, http.StatusCreated)
	var created domain.Task
	decodeJSON(t, resp, &created)
	return created.ID
}

func TestAPITaskCreateInsertsAfterAnchor(t *testing.T) {
	srv, store := newTestServer(t, "")
	pid, cid, tid := seedTask(t, srv, store, "first")

	resp := do(t, srv, newJSONRequest(t, "POST",
		"/api/v1/projects/"+pid+"/categories/"+cid+"/tasks",
		map[string]any{"title": "inserted", "insert_after": tid}))
	assertStatus(t, resp, http.StatusCreated)
	var created domain.Task
	decodeJSON(t, resp, &created)

	p, err := store.LoadProjectByID(pid)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	tasks := p.Categories[0].Tasks
	var anchorIdx, newIdx = -1, -1
	for i, task := range tasks {
		switch task.ID {
		case tid:
			anchorIdx = i
		case created.ID:
			newIdx = i
		}
	}
	if newIdx != anchorIdx+1 {
		t.Fatalf("want the new task directly below the anchor (%d), got %d", anchorIdx+1, newIdx)
	}
}

func TestAPISeparatorCreate(t *testing.T) {
	srv, store := newTestServer(t, "")
	pid, cid, tid := seedTask(t, srv, store, "first")

	sid := seedSeparator(t, srv, pid, cid, tid, "Later")

	stored := findTaskInStore(t, store, pid, sid)
	if stored == nil {
		t.Fatal("separator not persisted")
	}
	if !stored.IsSeparator() {
		t.Fatalf("want a separator, got kind %q", stored.Kind)
	}
	if stored.Title != "Later" {
		t.Fatalf("label: want Later, got %q", stored.Title)
	}
	// The task defaults must not leak onto a divider.
	if stored.Status != "" {
		t.Fatalf("separator carries a status: %q", stored.Status)
	}
}

func TestAPISeparatorCreateAllowsNoLabel(t *testing.T) {
	srv, store := newTestServer(t, "")
	pid, cid, tid := seedTask(t, srv, store, "first")

	// An unlabeled separator is a plain rule — a blank title must not 400 here
	// the way it does for a task.
	sid := seedSeparator(t, srv, pid, cid, tid, "")

	stored := findTaskInStore(t, store, pid, sid)
	if stored == nil || stored.Title != "" {
		t.Fatalf("want an unlabeled separator, got %+v", stored)
	}
}

// Creating a separator with task fields must 400, not silently drop them. The
// verb has always rejected this; the handler used to strip the fields before
// the verb could see them, so the request came back 201 having quietly ignored
// half of what it was asked to do.
func TestAPISeparatorCreateRejectsTaskFields(t *testing.T) {
	srv, store := newTestServer(t, "")
	pid, cid, tid := seedTask(t, srv, store, "first")

	cases := []struct {
		name string
		body map[string]any
	}{
		{"status", map[string]any{"status": domain.StatusCompleted}},
		{"priority", map[string]any{"priority": domain.PriorityHigh}},
		{"estimate", map[string]any{"estimate_minutes": 90}},
		{"description", map[string]any{"description": "notes"}},
		{"tag", map[string]any{"tag_label": "wip"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := map[string]any{
				"kind":         domain.KindSeparator,
				"insert_after": tid,
			}
			for k, v := range tc.body {
				body[k] = v
			}
			resp := do(t, srv, newJSONRequest(t, "POST",
				"/api/v1/projects/"+pid+"/categories/"+cid+"/tasks", body))
			assertStatus(t, resp, http.StatusBadRequest)
			assertJSONError(t, resp)
		})
	}

	// Nothing should have been created by the rejected requests.
	p, err := store.LoadProjectByID(pid)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	for _, task := range p.Categories[0].Tasks {
		if task.IsSeparator() {
			t.Fatalf("a rejected request still created a separator: %+v", task)
		}
	}
}

func TestAPITaskCreateRejectsUnknownKind(t *testing.T) {
	srv, store := newTestServer(t, "")
	pid, cid, _ := seedTask(t, srv, store, "first")

	resp := do(t, srv, newJSONRequest(t, "POST",
		"/api/v1/projects/"+pid+"/categories/"+cid+"/tasks",
		map[string]any{"title": "x", "kind": "milestone"}))
	assertStatus(t, resp, http.StatusBadRequest)
	assertJSONError(t, resp)
}

func TestAPISeparatorRelabel(t *testing.T) {
	srv, store := newTestServer(t, "")
	pid, cid, tid := seedTask(t, srv, store, "first")
	sid := seedSeparator(t, srv, pid, cid, tid, "Later")

	resp := do(t, srv, newJSONRequest(t, "PATCH", taskPath(pid, cid, sid),
		map[string]any{"title": "Much later"}))
	assertStatus(t, resp, http.StatusOK)

	stored := findTaskInStore(t, store, pid, sid)
	if stored == nil || stored.Title != "Much later" {
		t.Fatalf("relabel not persisted: %+v", stored)
	}
	if !stored.IsSeparator() {
		t.Fatal("relabeling must not change the kind")
	}

	// Clearing the label is legal for a separator, unlike for a task.
	resp = do(t, srv, newJSONRequest(t, "PATCH", taskPath(pid, cid, sid),
		map[string]any{"title": ""}))
	assertStatus(t, resp, http.StatusOK)
	if stored := findTaskInStore(t, store, pid, sid); stored == nil || stored.Title != "" {
		t.Fatalf("want the label cleared, got %+v", stored)
	}
}

// The regression this whole pass exists for: a PATCH must not be able to turn a
// divider into a half-task.
func TestAPISeparatorRejectsTaskFields(t *testing.T) {
	srv, store := newTestServer(t, "")
	pid, cid, tid := seedTask(t, srv, store, "first")
	sid := seedSeparator(t, srv, pid, cid, tid, "Later")

	cases := []struct {
		name string
		body map[string]any
	}{
		{"status", map[string]any{"status": domain.StatusCompleted}},
		{"priority", map[string]any{"priority": domain.PriorityHigh}},
		{"estimate", map[string]any{"estimate_minutes": 30}},
		{"description", map[string]any{"description": "notes"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := do(t, srv, newJSONRequest(t, "PATCH", taskPath(pid, cid, sid), tc.body))
			assertStatus(t, resp, http.StatusBadRequest)
			assertJSONError(t, resp)
		})
	}

	stored := findTaskInStore(t, store, pid, sid)
	if stored == nil {
		t.Fatal("separator vanished")
	}
	if stored.Status != "" || stored.Priority != "" ||
		stored.EstimateMinutes != 0 || stored.Description != "" {
		t.Fatalf("separator was corrupted: %+v", stored)
	}
}

// The status endpoint is a separate door into the same field — it must be shut
// for separators too.
func TestAPISeparatorRejectsStatusEndpoint(t *testing.T) {
	srv, store := newTestServer(t, "")
	pid, cid, tid := seedTask(t, srv, store, "first")
	sid := seedSeparator(t, srv, pid, cid, tid, "Later")

	resp := do(t, srv, newJSONRequest(t, "POST",
		taskPath(pid, cid, sid)+"/status",
		map[string]any{"status": domain.StatusCompleted}))
	assertStatus(t, resp, http.StatusBadRequest)
	assertJSONError(t, resp)

	if stored := findTaskInStore(t, store, pid, sid); stored == nil || stored.Status != "" {
		t.Fatalf("separator took a status: %+v", stored)
	}
}

func TestAPISeparatorDelete(t *testing.T) {
	srv, store := newTestServer(t, "")
	pid, cid, tid := seedTask(t, srv, store, "first")
	sid := seedSeparator(t, srv, pid, cid, tid, "Later")

	resp := do(t, srv, newRequest(t, "DELETE", taskPath(pid, cid, sid), nil))
	assertStatus(t, resp, http.StatusNoContent)

	if stored := findTaskInStore(t, store, pid, sid); stored != nil {
		t.Fatalf("separator still present: %+v", stored)
	}
}

func TestAPIFoldsUnknownProject(t *testing.T) {
	srv, _ := newTestServer(t, "")

	resp := do(t, srv, newRequest(t, "GET", "/api/v1/projects/bogus/folds", nil))
	assertStatus(t, resp, http.StatusNotFound)
	assertJSONError(t, resp)

	resp = do(t, srv, newJSONRequest(t, "PUT", "/api/v1/projects/bogus/folds",
		foldsPayload{FoldedCategories: []string{"x"}}))
	assertStatus(t, resp, http.StatusNotFound)
	assertJSONError(t, resp)
}
