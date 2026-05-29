package web

import (
	"errors"
	"net/http"
	"strings"

	"phasionary/internal/data"
	"phasionary/internal/domain"
)

func (s *Server) handleProjectsList(w http.ResponseWriter, r *http.Request) {
	projects, err := s.store.ListProjects()
	if err != nil {
		s.internalError(w, err)
		return
	}
	s.render(w, "projects", projectsPageData{
		Title:    "Projects",
		Projects: projects,
	})
}

func (s *Server) handleProjectNew(w http.ResponseWriter, r *http.Request) {
	s.render(w, "project_form", projectFormData{
		Title: "New project",
		IsNew: true,
	})
}

func (s *Server) handleProjectCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.badRequest(w, "invalid form")
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		s.render(w, "project_form", projectFormData{
			Title: "New project",
			IsNew: true,
			Error: "Name is required.",
		})
		return
	}
	project, err := s.store.CreateProject(name)
	if err != nil {
		if errors.Is(err, data.ErrDuplicateProjectName) {
			s.render(w, "project_form", projectFormData{
				Title:   "New project",
				IsNew:   true,
				Project: domain.Project{Name: name},
				Error:   "A project with that name already exists.",
			})
			return
		}
		s.internalError(w, err)
		return
	}
	http.Redirect(w, r, "/projects/"+project.ID, http.StatusSeeOther)
}

func (s *Server) handleProjectShow(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("pid")
	project, err := s.store.LoadProjectByID(pid)
	if err != nil {
		if errors.Is(err, data.ErrProjectNotFound) {
			s.notFound(w, "project not found")
			return
		}
		s.internalError(w, err)
		return
	}
	s.render(w, "project", projectPageData{
		Title:   project.Name,
		Project: project,
	})
}

func (s *Server) handleProjectEdit(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("pid")
	project, err := s.store.LoadProjectByID(pid)
	if err != nil {
		if errors.Is(err, data.ErrProjectNotFound) {
			s.notFound(w, "project not found")
			return
		}
		s.internalError(w, err)
		return
	}
	s.render(w, "project_form", projectFormData{
		Title:   "Rename " + project.Name,
		Project: project,
		IsNew:   false,
	})
}

func (s *Server) handleProjectUpdate(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("pid")
	if err := r.ParseForm(); err != nil {
		s.badRequest(w, "invalid form")
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		project, _ := s.store.LoadProjectByID(pid)
		s.render(w, "project_form", projectFormData{
			Title:   "Rename " + project.Name,
			Project: project,
			IsNew:   false,
			Error:   "Name is required.",
		})
		return
	}
	project, err := s.withProject(pid, func(p *domain.Project) error {
		p.Name = name
		return nil
	})
	if err != nil {
		s.mutationError(w, err)
		return
	}
	http.Redirect(w, r, "/projects/"+project.ID, http.StatusSeeOther)
}

func (s *Server) handleProjectDelete(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("pid")
	if err := s.store.DeleteProject(pid); err != nil {
		s.mutationError(w, err)
		return
	}
	// htmx clients honor HX-Redirect; plain clients see the 303.
	w.Header().Set("HX-Redirect", "/projects")
	w.Header().Set("Location", "/projects")
	w.WriteHeader(http.StatusSeeOther)
}
