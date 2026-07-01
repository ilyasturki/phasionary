package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"phasionary/internal/data"
	"phasionary/internal/domain"
)

// newJSONRequest builds a request with a JSON body and the matching header.
func newJSONRequest(t *testing.T, method, target string, body any) *http.Request {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req := httptest.NewRequest(method, target, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// decodeJSON unmarshals a JSON response body into v, failing the test on error.
func decodeJSON(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	body := readBody(t, resp)
	if err := json.Unmarshal([]byte(body), v); err != nil {
		t.Fatalf("decode json: %v (body: %s)", err, body)
	}
}

func assertStatus(t *testing.T, resp *http.Response, want int) {
	t.Helper()
	if resp.StatusCode != want {
		t.Fatalf("status: want %d, got %d", want, resp.StatusCode)
	}
}

func assertJSONContentType(t *testing.T, resp *http.Response) {
	t.Helper()
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("Content-Type: want application/json, got %q", ct)
	}
}

// assertJSONError asserts resp carries a JSON body with a non-empty "error".
func assertJSONError(t *testing.T, resp *http.Response) {
	t.Helper()
	assertJSONContentType(t, resp)
	var body map[string]string
	decodeJSON(t, resp, &body)
	if body["error"] == "" {
		t.Fatalf("want JSON error body, got %v", body)
	}
}

func TestAPIProjectsListEmpty(t *testing.T) {
	srv, _ := newTestServer(t, "")
	resp := do(t, srv, newRequest(t, "GET", "/api/v1/projects", nil))
	assertStatus(t, resp, http.StatusOK)
	assertJSONContentType(t, resp)
	var projects []domain.Project
	decodeJSON(t, resp, &projects)
	if len(projects) != 0 {
		t.Fatalf("want empty array, got %d projects", len(projects))
	}
}

func TestAPIProjectsList(t *testing.T) {
	srv, store := newTestServer(t, "")
	if _, err := store.CreateProject("Alpha"); err != nil {
		t.Fatalf("create: %v", err)
	}
	resp := do(t, srv, newRequest(t, "GET", "/api/v1/projects", nil))
	assertStatus(t, resp, http.StatusOK)
	var projects []domain.Project
	decodeJSON(t, resp, &projects)
	if len(projects) != 1 {
		t.Fatalf("want 1 project, got %d", len(projects))
	}
	if projects[0].Name != "Alpha" {
		t.Fatalf("project name: want Alpha, got %q", projects[0].Name)
	}
	// Full project payload: categories must be present (list returns full projects).
	if len(projects[0].Categories) == 0 {
		t.Fatalf("expected categories in listed project")
	}
}

func TestAPIAuthNoTokenReturnsJSON401(t *testing.T) {
	srv, _ := newTestServer(t, "secret-token")
	resp := do(t, srv, newRequest(t, "GET", "/api/v1/projects", nil))
	// API auth failures return JSON, never a redirect.
	assertStatus(t, resp, http.StatusUnauthorized)
	if loc := resp.Header.Get("Location"); loc != "" {
		t.Fatalf("API auth failure must not redirect, got Location %q", loc)
	}
	assertJSONError(t, resp)
}

func TestAPIAuthWrongBearerReturnsJSON401(t *testing.T) {
	srv, _ := newTestServer(t, "secret-token")
	req := newRequest(t, "GET", "/api/v1/projects", nil)
	req.Header.Set("Authorization", "Bearer nope")
	resp := do(t, srv, req)
	assertStatus(t, resp, http.StatusUnauthorized)
	assertJSONError(t, resp)
}

func TestAPIAuthUnauthPostReturnsJSON401(t *testing.T) {
	srv, _ := newTestServer(t, "secret-token")
	resp := do(t, srv, newRequest(t, "POST",
		"/api/v1/projects/x/categories/y/tasks", nil))
	assertStatus(t, resp, http.StatusUnauthorized)
	assertJSONContentType(t, resp)
}

func TestAPIAuthCorrectBearerAccepted(t *testing.T) {
	srv, _ := newTestServer(t, "secret-token")
	req := newRequest(t, "GET", "/api/v1/projects", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	resp := do(t, srv, req)
	assertStatus(t, resp, http.StatusOK)
}

func TestAPIProjectGet(t *testing.T) {
	srv, store := newTestServer(t, "")
	p, err := store.CreateProject("Alpha")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	resp := do(t, srv, newRequest(t, "GET", "/api/v1/projects/"+p.ID, nil))
	assertStatus(t, resp, http.StatusOK)
	assertJSONContentType(t, resp)
	var got domain.Project
	decodeJSON(t, resp, &got)
	if got.ID != p.ID || got.Name != "Alpha" {
		t.Fatalf("project: want {%s Alpha}, got {%s %s}", p.ID, got.ID, got.Name)
	}
	if len(got.Categories) == 0 || len(got.Categories[0].Tasks) == 0 {
		t.Fatalf("expected full project with categories and tasks")
	}
}

func TestAPIProjectGetNotFound(t *testing.T) {
	srv, _ := newTestServer(t, "")
	resp := do(t, srv, newRequest(t, "GET", "/api/v1/projects/bogus-id", nil))
	assertStatus(t, resp, http.StatusNotFound)
	assertJSONError(t, resp)
}

// findTaskInStore reloads the project and returns the task with id tid, or nil.
func findTaskInStore(t *testing.T, store *data.Store, pid, tid string) *domain.Task {
	t.Helper()
	p, err := store.LoadProjectByID(pid)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	for _, c := range p.Categories {
		for i := range c.Tasks {
			if c.Tasks[i].ID == tid {
				return &c.Tasks[i]
			}
		}
	}
	return nil
}

func TestAPITaskCreate(t *testing.T) {
	srv, store := newTestServer(t, "")
	p, _ := store.CreateProject("Test")
	pid, cid := p.ID, p.Categories[0].ID

	resp := do(t, srv, newJSONRequest(t, "POST",
		"/api/v1/projects/"+pid+"/categories/"+cid+"/tasks",
		map[string]any{
			"title":            "Ship the API",
			"status":           domain.StatusInProgress,
			"priority":         domain.PriorityHigh,
			"estimate_minutes": 90,
			"description":      "with tests",
		}))
	assertStatus(t, resp, http.StatusCreated)
	assertJSONContentType(t, resp)
	var got domain.Task
	decodeJSON(t, resp, &got)
	if got.ID == "" || got.CreatedAt == "" {
		t.Fatalf("created task missing generated fields: %+v", got)
	}
	if got.Title != "Ship the API" || got.Status != domain.StatusInProgress ||
		got.Priority != domain.PriorityHigh || got.EstimateMinutes != 90 ||
		got.Description != "with tests" {
		t.Fatalf("created task fields wrong: %+v", got)
	}
	// Persisted, not just echoed.
	stored := findTaskInStore(t, store, pid, got.ID)
	if stored == nil {
		t.Fatalf("task not persisted to store")
	}
	if stored.Title != "Ship the API" || stored.Status != domain.StatusInProgress {
		t.Fatalf("persisted task fields wrong: %+v", stored)
	}
}

func TestAPITaskCreateDefaultsStatusToTodo(t *testing.T) {
	srv, store := newTestServer(t, "")
	p, _ := store.CreateProject("Test")
	pid, cid := p.ID, p.Categories[0].ID

	resp := do(t, srv, newJSONRequest(t, "POST",
		"/api/v1/projects/"+pid+"/categories/"+cid+"/tasks",
		map[string]any{"title": "No status given"}))
	assertStatus(t, resp, http.StatusCreated)
	var got domain.Task
	decodeJSON(t, resp, &got)
	if got.Status != domain.StatusTodo {
		t.Fatalf("default status: want todo, got %q", got.Status)
	}
}

func TestAPITaskCreateValidation(t *testing.T) {
	srv, store := newTestServer(t, "")
	p, _ := store.CreateProject("Test")
	pid, cid := p.ID, p.Categories[0].ID
	base := "/api/v1/projects/" + pid + "/categories/" + cid + "/tasks"

	cases := []struct {
		name string
		body map[string]any
	}{
		{"missing title", map[string]any{"status": "todo"}},
		{"blank title", map[string]any{"title": "   "}},
		{"invalid status", map[string]any{"title": "x", "status": "bogus"}},
		{"invalid priority", map[string]any{"title": "x", "priority": "bogus"}},
		{"negative estimate", map[string]any{"title": "x", "estimate_minutes": -5}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := do(t, srv, newJSONRequest(t, "POST", base, tc.body))
			assertStatus(t, resp, http.StatusBadRequest)
			assertJSONError(t, resp)
		})
	}
}

func TestAPITaskCreateMalformedJSON(t *testing.T) {
	srv, store := newTestServer(t, "")
	p, _ := store.CreateProject("Test")
	pid, cid := p.ID, p.Categories[0].ID

	req := httptest.NewRequest("POST",
		"/api/v1/projects/"+pid+"/categories/"+cid+"/tasks",
		strings.NewReader("{not valid json"))
	req.Header.Set("Content-Type", "application/json")
	resp := do(t, srv, req)
	assertStatus(t, resp, http.StatusBadRequest)
	assertJSONContentType(t, resp)
}

func TestAPITaskCreateNotFound(t *testing.T) {
	srv, store := newTestServer(t, "")
	p, _ := store.CreateProject("Test")
	pid, cid := p.ID, p.Categories[0].ID

	cases := []struct {
		name, path string
	}{
		{"unknown project", "/api/v1/projects/bogus/categories/" + cid + "/tasks"},
		{"unknown category", "/api/v1/projects/" + pid + "/categories/bogus/tasks"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := do(t, srv, newJSONRequest(t, "POST", tc.path,
				map[string]any{"title": "x"}))
			assertStatus(t, resp, http.StatusNotFound)
			assertJSONContentType(t, resp)
		})
	}
}

func TestAPITaskSetStatus(t *testing.T) {
	srv, store := newTestServer(t, "")
	p, _ := store.CreateProject("Test")
	pid := p.ID
	cid := p.Categories[0].ID
	tid := p.Categories[0].Tasks[0].ID

	resp := do(t, srv, newJSONRequest(t, "POST",
		"/api/v1/projects/"+pid+"/categories/"+cid+"/tasks/"+tid+"/status",
		map[string]any{"status": domain.StatusCompleted}))
	assertStatus(t, resp, http.StatusOK)
	assertJSONContentType(t, resp)
	var got domain.Task
	decodeJSON(t, resp, &got)
	if got.Status != domain.StatusCompleted {
		t.Fatalf("status: want completed, got %q", got.Status)
	}
	if got.CompletionDate == "" {
		t.Fatalf("completing a task must set completion_date")
	}
	// Persisted.
	stored := findTaskInStore(t, store, pid, tid)
	if stored == nil || stored.Status != domain.StatusCompleted {
		t.Fatalf("status change not persisted: %+v", stored)
	}
}

func TestAPITaskSetStatusClearsCompletionOnReopen(t *testing.T) {
	srv, store := newTestServer(t, "")
	p, _ := store.CreateProject("Test")
	pid := p.ID
	cid := p.Categories[0].ID
	tid := p.Categories[0].Tasks[0].ID
	base := "/api/v1/projects/" + pid + "/categories/" + cid + "/tasks/" + tid + "/status"

	do(t, srv, newJSONRequest(t, "POST", base, map[string]any{"status": domain.StatusCompleted}))
	resp := do(t, srv, newJSONRequest(t, "POST", base, map[string]any{"status": domain.StatusTodo}))
	assertStatus(t, resp, http.StatusOK)
	var got domain.Task
	decodeJSON(t, resp, &got)
	if got.Status != domain.StatusTodo || got.CompletionDate != "" {
		t.Fatalf("reopening must clear completion_date: %+v", got)
	}
}

func TestAPITaskSetStatusValidation(t *testing.T) {
	srv, store := newTestServer(t, "")
	p, _ := store.CreateProject("Test")
	pid := p.ID
	cid := p.Categories[0].ID
	tid := p.Categories[0].Tasks[0].ID
	base := "/api/v1/projects/" + pid + "/categories/" + cid + "/tasks/" + tid + "/status"

	cases := []struct {
		name string
		body map[string]any
	}{
		{"invalid status", map[string]any{"status": "bogus"}},
		{"missing status", map[string]any{}},
		{"empty status", map[string]any{"status": ""}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := do(t, srv, newJSONRequest(t, "POST", base, tc.body))
			assertStatus(t, resp, http.StatusBadRequest)
			assertJSONContentType(t, resp)
		})
	}
}

func TestAPITaskSetStatusNotFound(t *testing.T) {
	srv, store := newTestServer(t, "")
	p, _ := store.CreateProject("Test")
	pid := p.ID
	cid := p.Categories[0].ID
	tid := p.Categories[0].Tasks[0].ID

	cases := []struct {
		name, path string
	}{
		{"unknown project", "/api/v1/projects/bogus/categories/" + cid + "/tasks/" + tid + "/status"},
		{"unknown category", "/api/v1/projects/" + pid + "/categories/bogus/tasks/" + tid + "/status"},
		{"unknown task", "/api/v1/projects/" + pid + "/categories/" + cid + "/tasks/bogus/status"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := do(t, srv, newJSONRequest(t, "POST", tc.path,
				map[string]any{"status": domain.StatusCompleted}))
			assertStatus(t, resp, http.StatusNotFound)
			assertJSONContentType(t, resp)
		})
	}
}
