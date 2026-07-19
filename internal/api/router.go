package api

import "net/http"

func (s *Server) handler() http.Handler {
	mux := http.NewServeMux()
	s.registerRoutes(mux)

	var h http.Handler = mux
	h = s.authMiddleware(h)
	h = panicMiddleware(h)
	h = loggingMiddleware(h)
	return h
}

// registerRoutes wires the /api/v1 JSON API. authMiddleware requires a valid
// Bearer token (when one is configured) and returns JSON 401s on failure.
func (s *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/projects", s.handleAPIProjectsList)
	mux.HandleFunc("GET /api/v1/projects/{pid}", s.handleAPIProjectGet)
	mux.HandleFunc("GET /api/v1/projects/{pid}/folds", s.handleAPIFoldsGet)
	mux.HandleFunc("PUT /api/v1/projects/{pid}/folds", s.handleAPIFoldsSet)
	mux.HandleFunc("POST /api/v1/projects/{pid}/categories", s.handleAPICategoryCreate)
	mux.HandleFunc("POST /api/v1/projects/{pid}/categories/{cid}/tasks", s.handleAPITaskCreate)
	mux.HandleFunc("PATCH /api/v1/projects/{pid}/categories/{cid}/tasks/{tid}", s.handleAPITaskUpdate)
	mux.HandleFunc("DELETE /api/v1/projects/{pid}/categories/{cid}/tasks/{tid}", s.handleAPITaskDelete)
	mux.HandleFunc("POST /api/v1/projects/{pid}/categories/{cid}/tasks/{tid}/status", s.handleAPITaskStatus)
}
