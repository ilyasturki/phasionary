package web

import (
	"errors"
	"net/http"
	"strings"

	"phasionary/internal/domain"
)

func (s *Server) handleCategoryNew(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("pid")
	project, err := s.store.LoadProjectByID(pid)
	if err != nil {
		s.mutationError(w, err)
		return
	}
	s.render(w, "category_form", categoryFormData{
		Title:   "New category",
		Project: project,
		IsNew:   true,
	})
}

func (s *Server) handleCategoryCreate(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("pid")
	if err := r.ParseForm(); err != nil {
		s.badRequest(w, "invalid form")
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))

	var renderError string
	project, err := s.withProject(pid, func(p *domain.Project) error {
		if name == "" {
			renderError = "Name is required."
			return errNoChange
		}
		if _, err := p.AddCategoryNamed(name); err != nil {
			if errors.Is(err, domain.ErrDuplicateCategoryName) {
				renderError = "A category with that name already exists."
				return errNoChange
			}
			return err
		}
		return nil
	})
	if err != nil {
		s.mutationError(w, err)
		return
	}
	if renderError != "" {
		s.render(w, "category_form", categoryFormData{
			Title:    "New category",
			Project:  project,
			Category: domain.Category{Name: name},
			IsNew:    true,
			Error:    renderError,
		})
		return
	}
	http.Redirect(w, r, "/projects/"+pid, http.StatusSeeOther)
}

func (s *Server) handleCategoryEdit(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("pid")
	cid := r.PathValue("cid")
	project, err := s.store.LoadProjectByID(pid)
	if err != nil {
		s.mutationError(w, err)
		return
	}
	idx, err := project.FindCategoryByID(cid)
	if err != nil {
		s.mutationError(w, err)
		return
	}
	s.render(w, "category_form", categoryFormData{
		Title:    "Rename " + project.Categories[idx].Name,
		Project:  project,
		Category: project.Categories[idx],
		IsNew:    false,
	})
}

func (s *Server) handleCategoryUpdate(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("pid")
	cid := r.PathValue("cid")
	if err := r.ParseForm(); err != nil {
		s.badRequest(w, "invalid form")
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))

	var renderError string
	project, err := s.withProject(pid, func(p *domain.Project) error {
		idx, err := p.FindCategoryByID(cid)
		if err != nil {
			return err
		}
		if name == "" {
			renderError = "Name is required."
			return errNoChange
		}
		if err := p.RenameCategory(idx, name); err != nil {
			if errors.Is(err, domain.ErrDuplicateCategoryName) {
				renderError = "A category with that name already exists."
				return errNoChange
			}
			return err
		}
		return nil
	})
	if err != nil {
		s.mutationError(w, err)
		return
	}
	if renderError != "" {
		idx, _ := project.FindCategoryByID(cid)
		cat := domain.Category{Name: name}
		if idx >= 0 {
			cat = project.Categories[idx]
			cat.Name = name
		}
		s.render(w, "category_form", categoryFormData{
			Title:    "Rename category",
			Project:  project,
			Category: cat,
			IsNew:    false,
			Error:    renderError,
		})
		return
	}
	http.Redirect(w, r, "/projects/"+pid, http.StatusSeeOther)
}

func (s *Server) handleCategoryDelete(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("pid")
	cid := r.PathValue("cid")
	_, err := s.withProject(pid, func(p *domain.Project) error {
		idx, err := p.FindCategoryByID(cid)
		if err != nil {
			return err
		}
		return p.RemoveCategory(idx)
	})
	if err != nil {
		s.mutationError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleCategoryMove(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("pid")
	cid := r.PathValue("cid")
	delta, err := parseMoveDir(r)
	if err != nil {
		s.badRequest(w, err.Error())
		return
	}
	project, err := s.withProject(pid, func(p *domain.Project) error {
		idx, err := p.FindCategoryByID(cid)
		if err != nil {
			return err
		}
		// Swallow edge errors so the partial re-renders the unchanged state.
		if mvErr := p.MoveCategory(idx, delta); mvErr != nil {
			return errNoChange
		}
		return nil
	})
	if err != nil {
		s.mutationError(w, err)
		return
	}
	s.renderPartial(w, "categories", categoriesViewData{
		Project:    project,
		Categories: project.Categories,
	})
}
