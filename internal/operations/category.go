package operations

import (
	"strings"

	"phasionary/internal/domain"
)

// CreateCategory adds a category named name (trimmed) to p. Returns
// ErrNameRequired for an empty name or domain.ErrDuplicateCategoryName when the
// name collides case-insensitively with an existing category.
func CreateCategory(p *domain.Project, name string) (domain.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return domain.Category{}, ErrNameRequired
	}
	if err := domain.ValidateLine(name); err != nil {
		return domain.Category{}, err
	}
	cat, err := p.AddCategoryNamed(name)
	if err != nil {
		return domain.Category{}, err
	}
	return *cat, nil
}

// RenameCategory renames the category identified by categoryID to name
// (trimmed). Returns ErrNameRequired, domain.ErrCategoryNotFound, or
// domain.ErrDuplicateCategoryName.
func RenameCategory(p *domain.Project, categoryID, name string) (domain.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return domain.Category{}, ErrNameRequired
	}
	if err := domain.ValidateLine(name); err != nil {
		return domain.Category{}, err
	}
	idx, err := p.FindCategoryByID(categoryID)
	if err != nil {
		return domain.Category{}, err
	}
	if err := p.RenameCategory(idx, name); err != nil {
		return domain.Category{}, err
	}
	return p.Categories[idx], nil
}

// MoveCategory shifts the identified category by delta positions. A move that
// would fall off either end changes nothing and returns ErrNoChange.
func MoveCategory(p *domain.Project, categoryID string, delta int) (domain.Category, error) {
	idx, err := p.FindCategoryByID(categoryID)
	if err != nil {
		return domain.Category{}, err
	}
	dst := idx + delta
	if dst < 0 || dst >= len(p.Categories) {
		return p.Categories[idx], ErrNoChange
	}
	if err := p.MoveCategory(idx, delta); err != nil {
		return domain.Category{}, err
	}
	return p.Categories[dst], nil
}

// DeleteCategory removes the identified category and, with it, every task it
// held (the cascade is inherent in removing the category from the project).
func DeleteCategory(p *domain.Project, categoryID string) error {
	idx, err := p.FindCategoryByID(categoryID)
	if err != nil {
		return err
	}
	return p.RemoveCategory(idx)
}
