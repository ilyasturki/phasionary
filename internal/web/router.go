package web

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

func (s *Server) registerRoutes(mux *http.ServeMux) {
	mux.Handle("GET /static/", s.staticHandler)
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /{$}", s.handleIndex)

	mux.HandleFunc("GET /projects", s.handleProjectsList)
	mux.HandleFunc("GET /projects/new", s.handleProjectNew)
	mux.HandleFunc("POST /projects", s.handleProjectCreate)
	mux.HandleFunc("GET /projects/{pid}", s.handleProjectShow)
	mux.HandleFunc("GET /projects/{pid}/edit", s.handleProjectEdit)
	mux.HandleFunc("POST /projects/{pid}", s.handleProjectUpdate)
	mux.HandleFunc("DELETE /projects/{pid}", s.handleProjectDelete)
	// Plain-form fallback for the delete button when htmx is unavailable;
	// without it a POST would route to handleProjectUpdate and fail
	// validation rather than deleting.
	mux.HandleFunc("POST /projects/{pid}/delete", s.handleProjectDelete)

	mux.HandleFunc("GET /projects/{pid}/categories/new", s.handleCategoryNew)
	mux.HandleFunc("POST /projects/{pid}/categories", s.handleCategoryCreate)
	mux.HandleFunc("GET /projects/{pid}/categories/{cid}/edit", s.handleCategoryEdit)
	mux.HandleFunc("POST /projects/{pid}/categories/{cid}", s.handleCategoryUpdate)
	mux.HandleFunc("DELETE /projects/{pid}/categories/{cid}", s.handleCategoryDelete)
	mux.HandleFunc("POST /projects/{pid}/categories/{cid}/move", s.handleCategoryMove)

	mux.HandleFunc("GET /projects/{pid}/categories/{cid}/tasks/new", s.handleTaskNew)
	mux.HandleFunc("POST /projects/{pid}/categories/{cid}/tasks", s.handleTaskCreate)
	mux.HandleFunc("GET /projects/{pid}/categories/{cid}/tasks/{tid}/edit", s.handleTaskEdit)
	mux.HandleFunc("POST /projects/{pid}/categories/{cid}/tasks/{tid}", s.handleTaskUpdate)
	mux.HandleFunc("DELETE /projects/{pid}/categories/{cid}/tasks/{tid}", s.handleTaskDelete)
	mux.HandleFunc("POST /projects/{pid}/categories/{cid}/tasks/{tid}/status", s.handleTaskStatus)
	mux.HandleFunc("POST /projects/{pid}/categories/{cid}/tasks/{tid}/priority", s.handleTaskPriority)
	mux.HandleFunc("POST /projects/{pid}/categories/{cid}/tasks/{tid}/move", s.handleTaskMove)
}
