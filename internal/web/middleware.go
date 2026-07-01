package web

import (
	"crypto/subtle"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"time"
)

const sessionCookieName = "phasionary_session"

// setSessionCookie issues the long-lived cookie carrying the shared token.
// It is intentionally not Secure: the server speaks plain HTTP, and a Secure
// cookie would be silently dropped by browsers, breaking auth. Operators
// terminating TLS upstream should have the reverse proxy re-stamp it.
func setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   30 * 24 * 60 * 60,
	})
}

// tokenMatches reports whether provided equals the configured token, compared
// in constant time. An empty provided value never matches.
func (s *Server) tokenMatches(provided string) bool {
	if provided == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(s.cfg.Token)) == 1
}

// hasValidSession reports whether the request carries a session cookie holding
// the configured token.
func (s *Server) hasValidSession(r *http.Request) bool {
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		return false
	}
	return s.tokenMatches(c.Value)
}

// isPublicPath reports whether p is reachable without authentication: static
// assets and the login endpoints (otherwise an unauthenticated visitor could
// never reach the form). Centralizing the allowlist keeps it from drifting
// from the routes it mirrors.
func isPublicPath(p string) bool {
	return strings.HasPrefix(p, "/static/") || p == "/login"
}

// isAPIPath reports whether p targets the JSON API. Auth failures on these
// paths return a JSON 401 instead of an HTML login redirect, so a script
// client never has to parse a login page to discover it lacks credentials.
func isAPIPath(p string) bool {
	return strings.HasPrefix(p, "/api/")
}

// requestToken extracts a caller-supplied token, preferring the ?token= query
// param and falling back to an "Authorization: Bearer" header. viaBearer
// reports that the token came from the header, which marks the caller as a
// script/API client rather than a browser.
func requestToken(r *http.Request) (token string, viaBearer bool) {
	if t := r.URL.Query().Get("token"); t != "" {
		return t, false
	}
	if h := r.Header.Get("Authorization"); len(h) >= 7 && strings.EqualFold(h[:7], "Bearer ") {
		return h[7:], true
	}
	return "", false
}

// loginTarget builds the /login URL to send an unauthenticated browser to,
// preserving where it was headed in ?next= — minus any (failed) ?token= so the
// attempted secret doesn't linger in browser history or the login form.
func loginTarget(r *http.Request) string {
	u := *r.URL
	if u.Query().Has("token") {
		q := u.Query()
		q.Del("token")
		u.RawQuery = q.Encode()
	}
	return "/login?next=" + url.QueryEscape(u.RequestURI())
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	if s.cfg.Token == "" {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isPublicPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		if s.hasValidSession(r) {
			next.ServeHTTP(w, r)
			return
		}
		provided, viaBearer := requestToken(r)
		if s.tokenMatches(provided) {
			setSessionCookie(w, s.cfg.Token)
			// Strip ?token= and redirect to the clean URL so the secret doesn't
			// linger in browser history or referer headers. Only safe for GET:
			// 303 demotes other methods to GET, which would silently drop the
			// request body. For non-GET we just process the request inline and
			// trust the cookie to take over from the next request onwards. A
			// Bearer token isn't in the URL, so there's nothing to strip.
			if !viaBearer && r.Method == http.MethodGet {
				q := r.URL.Query()
				q.Del("token")
				target := r.URL.Path
				if encoded := q.Encode(); encoded != "" {
					target += "?" + encoded
				}
				http.Redirect(w, r, target, http.StatusSeeOther)
				return
			}
			next.ServeHTTP(w, r)
			return
		}
		// Unauthenticated. API clients get a JSON 401 for any auth failure —
		// never an HTML login redirect or a plain-text body — regardless of
		// method or whether a (bad) Bearer token was supplied.
		if isAPIPath(r.URL.Path) {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		// htmx swaps response bodies into the page, so a normal redirect would
		// drop the login form into a fragment; HX-Redirect tells htmx to do a
		// full client-side navigation instead (for any method).
		target := loginTarget(r)
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Redirect", target)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// A Bearer client is a script/API caller; so is any non-GET request,
		// which can't be sent to the login form with a 303 (that demotes it to
		// GET and silently drops the body). Both get a plain 401.
		if viaBearer || r.Method != http.MethodGet {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		// Browser GET: send it to the login form, preserving the destination.
		http.Redirect(w, r, target, http.StatusSeeOther)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, rw.status, time.Since(start))
	})
}

func panicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic %s %s: %v\n%s", r.Method, r.URL.Path, rec, debug.Stack())
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
