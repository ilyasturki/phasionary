package api

import (
	"crypto/subtle"
	"log"
	"net"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

// tokenMatches reports whether provided equals the configured token, compared
// in constant time. An empty provided value never matches.
func (s *Server) tokenMatches(provided string) bool {
	if provided == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(s.cfg.Token)) == 1
}

// bearerToken extracts the token from an "Authorization: Bearer <token>" header,
// returning "" when the header is absent or not a Bearer credential.
func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if len(h) >= 7 && strings.EqualFold(h[:7], "Bearer ") {
		return h[7:]
	}
	return ""
}

// authMiddleware requires a valid Bearer token on every request when a token is
// configured. Failures return a JSON 401 — never an HTML page — since the only
// clients are the mobile app and scripts.
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	if s.cfg.Token == "" {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.tokenMatches(bearerToken(r)) {
			next.ServeHTTP(w, r)
			return
		}
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
	})
}

// hostAllowed reports whether the Host header of a request may be served.
//
// DNS rebinding is a name attack: the page keeps its own hostname (evil.example)
// while the DNS answer flips to 127.0.0.1, so the browser treats the response as
// same-origin and can read it. What never changes is the Host header — it still
// carries the attacker's name. Rejecting names we don't recognise therefore
// kills the whole class.
//
// IP literals are allowed unconditionally, which is what keeps this from
// breaking real deployments: a phone hitting a LAN or Tailscale address sends
// Host: 100.x.y.z, and a page cannot mount a rebinding attack while sending an
// IP literal — reaching the server by IP is plain cross-origin, where the
// absence of any CORS header already stops the response being read. So the
// check costs LAN and Tailscale users nothing, and only a named host (MagicDNS,
// a reverse proxy) needs --allowed-host.
func (s *Server) hostAllowed(hostHeader string) bool {
	if hostHeader == "" {
		return false
	}
	host := hostHeader
	if h, _, err := net.SplitHostPort(hostHeader); err == nil {
		host = h
	}
	host = strings.TrimSuffix(strings.TrimPrefix(host, "["), "]")

	if net.ParseIP(host) != nil {
		return true
	}
	// localhost is reserved to the loopback address (RFC 6761) and browsers
	// resolve it without consulting DNS, so it cannot be rebound.
	if strings.EqualFold(host, "localhost") {
		return true
	}
	for _, allowed := range s.cfg.AllowedHosts {
		if strings.EqualFold(host, allowed) {
			return true
		}
	}
	return false
}

// hostMiddleware rejects requests carrying an unrecognised Host. It sits
// outside authMiddleware so a rebinding attempt is refused before any
// credential comparison happens.
func (s *Server) hostMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.hostAllowed(r.Host) {
			writeJSONError(w, http.StatusMisdirectedRequest,
				"unrecognised Host header; pass --allowed-host to permit this name")
			return
		}
		next.ServeHTTP(w, r)
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

// safeForLog renders an attacker-controlled string as a quoted Go literal, so
// control bytes become escapes (\x1b) instead of reaching the terminal. The
// request path arrives already percent-decoded, and this middleware is the
// outermost one — it logs even requests authMiddleware rejects — so an
// unescaped write would hand any unauthenticated caller a terminal-escape
// primitive (OSC 52 writes the operator's clipboard) against whoever is
// watching the server's output.
func safeForLog(s string) string { return strconv.Quote(s) }

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		log.Printf("%s %s %d %s", r.Method, safeForLog(r.URL.Path), rw.status, time.Since(start))
	})
}

func panicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic %s %s: %v\n%s", r.Method, safeForLog(r.URL.Path), rec, debug.Stack())
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
