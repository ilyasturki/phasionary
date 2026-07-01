package web

import (
	"errors"
	"net/http"
	"strings"

	"phasionary/internal/domain"
	"phasionary/internal/operations"
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

	project, err := s.withProject(pid, func(p *domain.Project) error {
		_, e := operations.CreateCategory(p, name)
		return e
	})
	switch {
	case err == nil:
		http.Redirect(w, r, "/projects/"+pid, http.StatusSeeOther)
	case errors.Is(err, operations.ErrNameRequired):
		s.renderCategoryFormError(w, project, name, "Name is required.")
	case errors.Is(err, domain.ErrDuplicateCategoryName):
		s.renderCategoryFormError(w, project, name, "A category with that name already exists.")
	default:
		s.mutationError(w, err)
	}
}

// renderCategoryFormError re-renders the new-category form with a message,
// echoing back the name the user just typed.
func (s *Server) renderCategoryFormError(w http.ResponseWriter, project domain.Project, name, msg string) {
	s.render(w, "category_form", categoryFormData{
		Title:    "New category",
		Project:  project,
		Category: domain.Category{Name: name},
		IsNew:    true,
		Error:    msg,
	})
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

	project, err := s.withProject(pid, func(p *domain.Project) error {
		_, e := operations.RenameCategory(p, cid, name)
		return e
	})
	switch {
	case err == nil:
		http.Redirect(w, r, "/projects/"+pid, http.StatusSeeOther)
	case errors.Is(err, operations.ErrNameRequired):
		s.renderCategoryRenameError(w, project, cid, name, "Name is required.")
	case errors.Is(err, domain.ErrDuplicateCategoryName):
		s.renderCategoryRenameError(w, project, cid, name, "A category with that name already exists.")
	default:
		s.mutationError(w, err)
	}
}

// renderCategoryRenameError re-renders the rename form with a message. If a
// concurrent writer deleted the category between the mutation and here, it
// surfaces a 404 instead of a form whose action URL is missing the id.
func (s *Server) renderCategoryRenameError(w http.ResponseWriter, project domain.Project, cid, name, msg string) {
	idx, err := project.FindCategoryByID(cid)
	if err != nil {
		s.mutationError(w, err)
		return
	}
	cat := project.Categories[idx]
	cat.Name = name
	s.render(w, "category_form", categoryFormData{
		Title:    "Rename category",
		Project:  project,
		Category: cat,
		IsNew:    false,
		Error:    msg,
	})
}

func (s *Server) handleCategoryDelete(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("pid")
	cid := r.PathValue("cid")
	_, err := s.withProject(pid, func(p *domain.Project) error {
		return operations.DeleteCategory(p, cid)
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
		_, e := operations.MoveCategory(p, cid, delta)
		return e
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
