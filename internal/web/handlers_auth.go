package web

import (
	"net/http"
	"net/url"
	"strings"
)

type loginPageData struct {
	Title string
	Next  string
	Error string
}

// safeNext sanitizes a post-login redirect target to a local path, guarding
// against open redirects. It must be a rooted path ("/...") with no scheme or
// host. Backslashes and control bytes are rejected outright: browsers fold "\"
// to "/", so "/\evil.com" would be re-parsed as the protocol-relative
// "//evil.com". Anything that fails these checks falls back to "/".
func safeNext(next string) string {
	if next == "" || !strings.HasPrefix(next, "/") {
		return "/"
	}
	if strings.ContainsAny(next, "\\\x00\t\r\n") {
		return "/"
	}
	u, err := url.Parse(next)
	if err != nil || u.Scheme != "" || u.Host != "" || strings.HasPrefix(u.Path, "//") {
		return "/"
	}
	return next
}

// handleLoginForm renders the token entry page. It is exempt from auth so an
// unauthenticated visitor can reach it, and is only registered when a token is
// configured (see registerRoutes), so there is no auth-disabled case to handle.
func (s *Server) handleLoginForm(w http.ResponseWriter, r *http.Request) {
	next := safeNext(r.URL.Query().Get("next"))
	if s.hasValidSession(r) {
		http.Redirect(w, r, next, http.StatusSeeOther)
		return
	}
	// The form carries the shared secret; keep it out of browser/proxy caches.
	w.Header().Set("Cache-Control", "no-store")
	s.render(w, "login", loginPageData{Title: "Sign in", Next: next})
}

// handleLogin verifies the submitted token and, on success, issues the session
// cookie and redirects to the originally requested page. Like handleLoginForm
// it is only registered when a token is configured.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.badRequest(w, "invalid form")
		return
	}
	next := safeNext(r.PostFormValue("next"))
	if s.tokenMatches(r.PostFormValue("token")) {
		setSessionCookie(w, s.cfg.Token)
		http.Redirect(w, r, next, http.StatusSeeOther)
		return
	}
	// Wrong token: re-render the form with an error at 401. No cookie is set.
	w.Header().Set("Cache-Control", "no-store")
	s.renderStatus(w, http.StatusUnauthorized, "login", loginPageData{
		Title: "Sign in",
		Next:  next,
		Error: "Incorrect token. Try again.",
	})
}
