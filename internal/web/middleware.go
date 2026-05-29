package web

import (
	"crypto/subtle"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

const sessionCookieName = "phasionary_session"

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	if s.cfg.Token == "" {
		return next
	}
	expected := []byte(s.cfg.Token)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}
		if c, err := r.Cookie(sessionCookieName); err == nil {
			if subtle.ConstantTimeCompare([]byte(c.Value), expected) == 1 {
				next.ServeHTTP(w, r)
				return
			}
		}
		hadQueryToken := r.URL.Query().Has("token")
		provided := r.URL.Query().Get("token")
		if provided == "" {
			if h := r.Header.Get("Authorization"); len(h) >= 7 && strings.EqualFold(h[:7], "Bearer ") {
				provided = h[7:]
			}
		}
		if provided != "" && subtle.ConstantTimeCompare([]byte(provided), expected) == 1 {
			// The server speaks plain HTTP, so don't set Secure — a Secure
			// cookie is silently dropped by browsers over HTTP and the
			// post-redirect request would re-trigger the auth path.
			// Operators terminating TLS in front of this server should
			// rely on a reverse proxy that re-stamps the cookie.
			http.SetCookie(w, &http.Cookie{
				Name:     sessionCookieName,
				Value:    s.cfg.Token,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   30 * 24 * 60 * 60,
			})
			// Strip ?token= and redirect to the clean URL so the secret doesn't
			// linger in browser history or referer headers. Only safe for GET:
			// 303 demotes other methods to GET, which would silently drop the
			// request body. For non-GET we just process the request inline and
			// trust the cookie to take over from the next request onwards.
			if hadQueryToken && r.Method == http.MethodGet {
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
		http.Error(w, "unauthorized", http.StatusUnauthorized)
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
