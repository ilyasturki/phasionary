package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"phasionary/internal/data"
	"phasionary/internal/domain"
)

func newTestServer(t *testing.T, token string) (*Server, *data.Store) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "projects")
	store := data.NewStore(dir)
	if err := store.Ensure(); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	srv, err := New(store, Config{Addr: "127.0.0.1:0", Token: token})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return srv, store
}

func newRequest(t *testing.T, method, target string, form url.Values) *http.Request {
	t.Helper()
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req := httptest.NewRequest(method, target, body)
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	return req
}

func do(t *testing.T, srv *Server, req *http.Request) *http.Response {
	t.Helper()
	rec := httptest.NewRecorder()
	srv.handler().ServeHTTP(rec, req)
	return rec.Result()
}

var (
	categoryIDRe = regexp.MustCompile(`id="category-([a-f0-9-]+)"`)
	taskIDRe     = regexp.MustCompile(`id="task-([a-f0-9-]+)"`)
)

func firstCategoryID(t *testing.T, html string) string {
	t.Helper()
	m := categoryIDRe.FindStringSubmatch(html)
	if m == nil {
		t.Fatalf("no category id found in html")
	}
	return m[1]
}

func firstTaskID(t *testing.T, html string) string {
	t.Helper()
	m := taskIDRe.FindStringSubmatch(html)
	if m == nil {
		t.Fatalf("no task id found in html")
	}
	return m[1]
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(b)
}

func TestIndexRedirectsToProjects(t *testing.T) {
	srv, _ := newTestServer(t, "")
	resp := do(t, srv, newRequest(t, "GET", "/", nil))
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("status: want 303, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Location"); got != "/projects" {
		t.Fatalf("location: want /projects, got %q", got)
	}
}

func TestProjectsListEmpty(t *testing.T) {
	srv, _ := newTestServer(t, "")
	resp := do(t, srv, newRequest(t, "GET", "/projects", nil))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: want 200, got %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, "No projects yet") {
		t.Fatalf("expected empty state, got: %s", body)
	}
}

func TestProjectCreateAndShow(t *testing.T) {
	srv, store := newTestServer(t, "")

	resp := do(t, srv, newRequest(t, "POST", "/projects", url.Values{"name": {"Mobile Test"}}))
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("create: want 303, got %d", resp.StatusCode)
	}
	loc := resp.Header.Get("Location")
	if !strings.HasPrefix(loc, "/projects/") {
		t.Fatalf("create location: want /projects/{id}, got %q", loc)
	}

	projects, _ := store.ListProjects()
	if len(projects) != 1 {
		t.Fatalf("project count: want 1, got %d", len(projects))
	}
	if projects[0].Name != "Mobile Test" {
		t.Fatalf("project name: want Mobile Test, got %q", projects[0].Name)
	}

	resp = do(t, srv, newRequest(t, "GET", loc, nil))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("show: want 200, got %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, "Mobile Test") {
		t.Fatalf("show body missing name")
	}
}

func TestProjectCreateRejectsEmptyName(t *testing.T) {
	srv, _ := newTestServer(t, "")
	resp := do(t, srv, newRequest(t, "POST", "/projects", url.Values{"name": {""}}))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: want 200 (form re-render), got %d", resp.StatusCode)
	}
	if !strings.Contains(readBody(t, resp), "Name is required") {
		t.Fatalf("expected validation error in re-rendered form")
	}
}

func TestProjectShow404(t *testing.T) {
	srv, _ := newTestServer(t, "")
	resp := do(t, srv, newRequest(t, "GET", "/projects/bogus-id", nil))
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: want 404, got %d", resp.StatusCode)
	}
}

func setupProjectWithTask(t *testing.T, srv *Server, store *data.Store) (pid, cid, tid string) {
	t.Helper()
	p, err := store.CreateProject("Test")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	pid = p.ID

	resp := do(t, srv, newRequest(t, "GET", "/projects/"+pid, nil))
	html := readBody(t, resp)
	cid = firstCategoryID(t, html)
	tid = firstTaskID(t, html)
	return pid, cid, tid
}

func TestTaskStatusCycle(t *testing.T) {
	srv, store := newTestServer(t, "")
	pid, cid, tid := setupProjectWithTask(t, srv, store)

	resp := do(t, srv, newRequest(t, "POST",
		"/projects/"+pid+"/categories/"+cid+"/tasks/"+tid+"/status", nil))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: want 200, got %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, `class="task-row`) {
		t.Fatalf("expected task_row partial, got: %s", body)
	}

	// Reload project and verify status actually changed
	p, _ := store.LoadProject(pid)
	for _, c := range p.Categories {
		if c.ID == cid {
			for _, tk := range c.Tasks {
				if tk.ID == tid {
					if tk.Status == domain.StatusTodo {
						t.Fatalf("status did not cycle (still todo)")
					}
					return
				}
			}
		}
	}
	t.Fatalf("task not found after cycle")
}

func TestTaskPriorityCycle(t *testing.T) {
	srv, store := newTestServer(t, "")
	pid, cid, tid := setupProjectWithTask(t, srv, store)

	// Start by clearing existing priority via update form
	for i := 0; i < 5; i++ {
		do(t, srv, newRequest(t, "POST",
			"/projects/"+pid+"/categories/"+cid+"/tasks/"+tid+"/priority", nil))
	}
	// After 5 cycles starting from any priority, we've wrapped around at least once.
	// Just verify the endpoint stays 200 and returns a row partial.
	resp := do(t, srv, newRequest(t, "POST",
		"/projects/"+pid+"/categories/"+cid+"/tasks/"+tid+"/priority", nil))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: want 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(readBody(t, resp), "priority-") {
		t.Fatalf("expected priority class in row partial")
	}
}

func TestTaskMoveDownThenUp(t *testing.T) {
	srv, store := newTestServer(t, "")
	pid, cid, _ := setupProjectWithTask(t, srv, store)

	p, _ := store.LoadProject(pid)
	var cat *domain.Category
	for i := range p.Categories {
		if p.Categories[i].ID == cid {
			cat = &p.Categories[i]
			break
		}
	}
	if cat == nil || len(cat.Tasks) < 2 {
		t.Skipf("need at least 2 tasks; got %d", len(cat.Tasks))
	}
	firstID := cat.Tasks[0].ID
	secondID := cat.Tasks[1].ID

	resp := do(t, srv, newRequest(t, "POST",
		"/projects/"+pid+"/categories/"+cid+"/tasks/"+firstID+"/move?dir=down", nil))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("move: want 200, got %d", resp.StatusCode)
	}

	p, _ = store.LoadProject(pid)
	for _, c := range p.Categories {
		if c.ID == cid {
			if c.Tasks[0].ID != secondID || c.Tasks[1].ID != firstID {
				t.Fatalf("order did not swap: got [%s, %s]", c.Tasks[0].ID, c.Tasks[1].ID)
			}
			return
		}
	}
}

func TestTaskMoveUpAtEdgeIsNoop(t *testing.T) {
	srv, store := newTestServer(t, "")
	pid, cid, tid := setupProjectWithTask(t, srv, store)

	resp := do(t, srv, newRequest(t, "POST",
		"/projects/"+pid+"/categories/"+cid+"/tasks/"+tid+"/move?dir=up", nil))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: want 200, got %d", resp.StatusCode)
	}

	p, _ := store.LoadProject(pid)
	for _, c := range p.Categories {
		if c.ID == cid {
			if c.Tasks[0].ID != tid {
				t.Fatalf("task moved past edge: first task is now %s, want %s", c.Tasks[0].ID, tid)
			}
			return
		}
	}
}

func TestTaskDelete(t *testing.T) {
	srv, store := newTestServer(t, "")
	pid, cid, tid := setupProjectWithTask(t, srv, store)

	resp := do(t, srv, newRequest(t, "DELETE",
		"/projects/"+pid+"/categories/"+cid+"/tasks/"+tid, nil))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("delete: want 200, got %d", resp.StatusCode)
	}

	p, _ := store.LoadProject(pid)
	for _, c := range p.Categories {
		if c.ID == cid {
			for _, tk := range c.Tasks {
				if tk.ID == tid {
					t.Fatalf("task still present after delete")
				}
			}
			return
		}
	}
}

func TestCategoryCreateRejectsDuplicate(t *testing.T) {
	srv, store := newTestServer(t, "")
	p, _ := store.CreateProject("Test")
	pid := p.ID
	dupe := p.Categories[0].Name

	resp := do(t, srv, newRequest(t, "POST",
		"/projects/"+pid+"/categories", url.Values{"name": {dupe}}))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: want 200 (form re-render), got %d", resp.StatusCode)
	}
	if !strings.Contains(readBody(t, resp), "already exists") {
		t.Fatalf("expected duplicate-name error")
	}
}

func TestCategoryDelete(t *testing.T) {
	srv, store := newTestServer(t, "")
	pid, cid, _ := setupProjectWithTask(t, srv, store)

	resp := do(t, srv, newRequest(t, "DELETE",
		"/projects/"+pid+"/categories/"+cid, nil))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("delete: want 200, got %d", resp.StatusCode)
	}

	p, _ := store.LoadProject(pid)
	for _, c := range p.Categories {
		if c.ID == cid {
			t.Fatalf("category still present after delete")
		}
	}
}

func TestProjectDeleteSendsRedirect(t *testing.T) {
	srv, store := newTestServer(t, "")
	p, _ := store.CreateProject("Test")
	pid := p.ID

	resp := do(t, srv, newRequest(t, "DELETE", "/projects/"+pid, nil))
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("status: want 303, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("HX-Redirect"); got != "/projects" {
		t.Fatalf("HX-Redirect: want /projects, got %q", got)
	}

	projects, _ := store.ListProjects()
	if len(projects) != 0 {
		t.Fatalf("expected no projects after delete, got %d", len(projects))
	}
}

func TestAuthMissingToken401(t *testing.T) {
	srv, _ := newTestServer(t, "secret-token")
	resp := do(t, srv, newRequest(t, "GET", "/projects", nil))
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: want 401, got %d", resp.StatusCode)
	}
}

func TestAuthBearerTokenAccepted(t *testing.T) {
	srv, _ := newTestServer(t, "secret-token")
	req := newRequest(t, "GET", "/projects", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	resp := do(t, srv, req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: want 200, got %d", resp.StatusCode)
	}
}

func TestAuthQueryTokenRedirectsAndSetsCookie(t *testing.T) {
	srv, _ := newTestServer(t, "secret-token")
	resp := do(t, srv, newRequest(t, "GET", "/projects?token=secret-token", nil))
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("status: want 303, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Location"); got != "/projects" {
		t.Fatalf("Location: want /projects (token stripped), got %q", got)
	}
	cookies := resp.Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == sessionCookieName && c.Value == "secret-token" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected session cookie")
	}
}

func TestAuthCookieAccepted(t *testing.T) {
	srv, _ := newTestServer(t, "secret-token")
	req := newRequest(t, "GET", "/projects", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "secret-token"})
	resp := do(t, srv, req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: want 200, got %d", resp.StatusCode)
	}
}

func TestAuthStaticAlwaysOpen(t *testing.T) {
	srv, _ := newTestServer(t, "secret-token")
	resp := do(t, srv, newRequest(t, "GET", "/static/style.css", nil))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: want 200, got %d", resp.StatusCode)
	}
}

func TestIsLoopbackAddr(t *testing.T) {
	cases := map[string]bool{
		"127.0.0.1:7777": true,
		"localhost:7777": true,
		"[::1]:7777":     true,
		"0.0.0.0:7777":   false,
		":7777":          false,
		"192.168.1.5:80": false,
		"":               false,
		"7777":           false,
	}
	for in, want := range cases {
		if got := IsLoopbackAddr(in); got != want {
			t.Errorf("IsLoopbackAddr(%q) = %v, want %v", in, got, want)
		}
	}
}
