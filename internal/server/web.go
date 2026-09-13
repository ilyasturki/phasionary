package server

import (
	"io/fs"
	"net"
	"net/http"
	"path"
	"slices"
	"strings"
)

const contentSecurityPolicy = "default-src 'none'; script-src 'self'; style-src 'self'; " +
	"img-src 'self' data:; font-src 'self'; connect-src 'self'; manifest-src 'self'; " +
	"worker-src 'self'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'"

const notBuiltMessage = "The web app was not built into this binary.\n\nBuild it with `nix build .#phasionary-server`, " +
	"or run `just build` before\n`go build`. The sync API under /v1 works either way."

func webHandler(files fs.FS) http.Handler {
	_, err := fs.Stat(files, "index.html")
	built := err == nil
	server := http.FileServerFS(files)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", contentSecurityPolicy)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")

		if !built {
			http.Error(w, notBuiltMessage, http.StatusServiceUnavailable)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if _, err := fs.Stat(files, name); err != nil {
			if strings.HasPrefix(r.URL.Path, "/assets/") {
				http.NotFound(w, r)
				return
			}
			r = r.Clone(r.Context())
			r.URL.Path = "/"
		}
		// /assets/ names are content-hashed; everything else must revalidate on deploy.
		cc := "no-cache"
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			cc = "public, max-age=31536000, immutable"
		}
		w.Header().Set("Cache-Control", cc)
		server.ServeHTTP(w, r)
	})
}

// IP literals always pass: a rebinding page cannot send Host: <ip>, and a
// request by IP is plain cross-origin with nothing readable.
func hostAllowed(hostHeader string, allowed []string) bool {
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
	// RFC 6761: browsers resolve localhost without DNS, so it cannot be rebound.
	if strings.EqualFold(host, "localhost") {
		return true
	}
	return slices.ContainsFunc(allowed, func(name string) bool { return strings.EqualFold(host, name) })
}

func hostMiddleware(allowed []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !hostAllowed(r.Host, allowed) {
			writeError(w, http.StatusMisdirectedRequest, "unrecognised Host header; pass --allowed-host to permit this name")
			return
		}
		next.ServeHTTP(w, r)
	})
}
