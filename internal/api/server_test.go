package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"phasionary/internal/data"
)

func newTestServer(t *testing.T, token string) (*Server, *data.Store) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "projects")
	store := data.NewStore(dir)
	if err := store.Ensure(); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	srv := New(store, Config{Addr: "127.0.0.1:0", Token: token})
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

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(b)
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
