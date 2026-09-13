package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

var builtFS = fstest.MapFS{
	"index.html":             {Data: []byte("<!doctype html><title>phasionary</title>")},
	"assets/index-abc123.js": {Data: []byte("export default 1")},
	"sw.js":                  {Data: []byte("// worker")},
}

func get(t *testing.T, h http.Handler, path string) *http.Response {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec.Result()
}

func TestWebServesTheAppAndItsRoutes(t *testing.T) {
	h := webHandler(builtFS)

	for _, path := range []string{"/", "/sync", "/projects/anything"} {
		resp := get(t, h, path)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s: status %d, want 200", path, resp.StatusCode)
		}
		if got := resp.Header.Get("Content-Security-Policy"); got != contentSecurityPolicy {
			t.Errorf("GET %s: CSP %q", path, got)
		}
	}

	if resp := get(t, h, "/assets/missing.js"); resp.StatusCode != http.StatusNotFound {
		t.Errorf("missing asset: status %d, want 404", resp.StatusCode)
	}
	if got := get(t, h, "/assets/index-abc123.js").Header.Get("Cache-Control"); !strings.Contains(got, "immutable") {
		t.Errorf("hashed asset Cache-Control = %q, want immutable", got)
	}
	if got := get(t, h, "/sw.js").Header.Get("Cache-Control"); got != "no-cache" {
		t.Errorf("sw.js Cache-Control = %q, want no-cache", got)
	}
}

func TestWebReportsAMissingBundle(t *testing.T) {
	resp := get(t, webHandler(fstest.MapFS{}), "/")
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status %d, want 503", resp.StatusCode)
	}
}

func TestHostAllowed(t *testing.T) {
	allowed := []string{"phas.example.net"}
	cases := []struct {
		host string
		want bool
	}{
		{"127.0.0.1:7777", true},
		{"100.64.1.2", true},
		{"[::1]:7777", true},
		{"localhost:7777", true},
		{"LocalHost", true},
		{"phas.example.net", true},
		{"PHAS.example.NET:443", true},
		{"evil.example", false},
		{"", false},
	}
	for _, c := range cases {
		if got := hostAllowed(c.host, allowed); got != c.want {
			t.Errorf("hostAllowed(%q) = %v, want %v", c.host, got, c.want)
		}
	}
}

func TestHostMiddlewareRejectsUnknownNames(t *testing.T) {
	resp := get(t, hostMiddleware(nil, http.NotFoundHandler()), "http://evil.example/")
	if resp.StatusCode != http.StatusMisdirectedRequest {
		t.Fatalf("status %d, want 421", resp.StatusCode)
	}
}

func TestAPIPathsNeverFallBackToTheApp(t *testing.T) {
	resp := get(t, newEnv(t).srv.Config.Handler, "http://127.0.0.1/v1/nope")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status %d, want 404", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("Content-Type = %q, want JSON", ct)
	}
}
