package journal

import (
	"slices"

	"phasionary/internal/domain"
)

// A nil old means no file existed (create, import): the diff is the full
// creation cascade. Drafts are ordered for apply-on-arrival: creates before
// reorders, deletes after moves.
func DiffProjects(old *domain.Project, updated domain.Project) []Draft {
	pid := updated.ID
	var drafts []Draft
	base := domain.Project{ID: pid}
	if old == nil {
		drafts = append(drafts, Draft{
			Kind:      KindProjectCreate,
			ProjectID: pid,
			Fields:    map[string]any{"name": updated.Name, "created_at": updated.CreatedAt},
		})
	} else {
		base = *old
		if base.Name != updated.Name {
			drafts = append(drafts, Draft{
				Kind:      KindProjectUpdate,
				ProjectID: pid,
				Fields:    map[string]any{"name": updated.Name},
			})
		}
	}

	oldCats := make(map[string]domain.Category, len(base.Categories))
	for _, c := range base.Categories {
		oldCats[c.ID] = c
	}
	newCats := make(map[string]domain.Category, len(updated.Categories))
	for _, c := range updated.Categories {
		newCats[c.ID] = c
	}

	type taskLoc struct {
		catID string
		task  *domain.Task
	}
	oldTasks := make(map[string]taskLoc)
	for _, c := range base.Categories {
		for i := range c.Tasks {
			oldTasks[c.Tasks[i].ID] = taskLoc{catID: c.ID, task: &c.Tasks[i]}
		}
	}

	for _, c := range updated.Categories {
		prev, existed := oldCats[c.ID]
		if !existed {
			drafts = append(drafts, Draft{
				Kind:      KindCategoryCreate,
				ProjectID: pid,
				EntityID:  c.ID,
				Fields:    categoryCreateFields(c),
			})
		} else if f := categoryChangedFields(prev, c); len(f) > 0 {
			drafts = append(drafts, Draft{
				Kind:      KindCategoryUpdate,
				ProjectID: pid,
				EntityID:  c.ID,
				Fields:    f,
			})
		}
	}

	seen := make(map[string]bool, len(oldTasks))
	for _, c := range updated.Categories {
		for _, t := range c.Tasks {
			seen[t.ID] = true
			prev, existed := oldTasks[t.ID]
			if !existed {
				drafts = append(drafts, Draft{
					Kind:      KindTaskCreate,
					ProjectID: pid,
					EntityID:  t.ID,
					Fields:    taskCreateFields(t, c.ID),
				})
				continue
			}
			if prev.catID != c.ID {
				drafts = append(drafts, Draft{
					Kind:      KindTaskMove,
					ProjectID: pid,
					EntityID:  t.ID,
					Fields:    map[string]any{"category_id": c.ID},
				})
			}
			if f := taskChangedFields(*prev.task, t); len(f) > 0 {
				drafts = append(drafts, Draft{
					Kind:      KindTaskUpdate,
					ProjectID: pid,
					EntityID:  t.ID,
					Fields:    f,
				})
			}
		}
	}

	// A vanished entity leaves no trace in the file; this tombstone is the only
	// thing distinguishing "deleted" from "never seen".
	for _, c := range base.Categories {
		for _, t := range c.Tasks {
			if !seen[t.ID] {
				drafts = append(drafts, Draft{
					Kind:      KindTaskDelete,
					ProjectID: pid,
					EntityID:  t.ID,
				})
			}
		}
	}
	for _, c := range base.Categories {
		if _, kept := newCats[c.ID]; !kept {
			drafts = append(drafts, Draft{
				Kind:      KindCategoryDelete,
				ProjectID: pid,
				EntityID:  c.ID,
			})
		}
	}

	// Any change to a list's ID sequence — insert, delete, move — emits one
	// reorder carrying the whole sequence.
	for _, c := range updated.Categories {
		newIDs := taskIDs(c)
		var oldIDs []string
		if prev, existed := oldCats[c.ID]; existed {
			oldIDs = taskIDs(prev)
		}
		if !slices.Equal(oldIDs, newIDs) {
			drafts = append(drafts, Draft{
				Kind:      KindCategoryReorder,
				ProjectID: pid,
				EntityID:  c.ID,
				Fields:    map[string]any{"task_ids": newIDs},
			})
		}
	}
	if newCatIDs := categoryIDs(updated); !slices.Equal(categoryIDs(base), newCatIDs) {
		drafts = append(drafts, Draft{
			Kind:      KindProjectReorder,
			ProjectID: pid,
			Fields:    map[string]any{"category_ids": newCatIDs},
		})
	}

	return drafts
}

func taskIDs(c domain.Category) []string {
	ids := make([]string, 0, len(c.Tasks))
	for _, t := range c.Tasks {
		ids = append(ids, t.ID)
	}
	return ids
}

func categoryIDs(p domain.Project) []string {
	ids := make([]string, 0, len(p.Categories))
	for _, c := range p.Categories {
		ids = append(ids, c.ID)
	}
	return ids
}

// setIfChanged records a field only when it moved. An empty new value is still
// recorded: absent means untouched, so a clear must travel.
func setIfChanged[T comparable](f map[string]any, key string, old, updated T) {
	if old != updated {
		f[key] = updated
	}
}

// setIfNonZero records a field only when it carries a value: on a create,
// absent and zero are the same to a reader starting from the zero value.
func setIfNonZero[T comparable](f map[string]any, key string, v T) {
	var zero T
	if v != zero {
		f[key] = v
	}
}

func categoryCreateFields(c domain.Category) map[string]any {
	f := map[string]any{"name": c.Name, "created_at": c.CreatedAt}
	setIfNonZero(f, "estimate_minutes", c.EstimateMinutes)
	return f
}

// UpdatedAt is deliberately not compared: the domain bumps it on every
// contained-task mutation, which would journal a category change per task edit.
func categoryChangedFields(old, updated domain.Category) map[string]any {
	f := map[string]any{}
	setIfChanged(f, "name", old.Name, updated.Name)
	setIfChanged(f, "estimate_minutes", old.EstimateMinutes, updated.EstimateMinutes)
	return f
}

func taskCreateFields(t domain.Task, categoryID string) map[string]any {
	f := map[string]any{
		"category_id": categoryID,
		"created_at":  t.CreatedAt,
		"updated_at":  t.UpdatedAt,
	}
	setIfNonZero(f, "title", t.Title)
	setIfNonZero(f, "status", t.Status)
	setIfNonZero(f, "priority", t.Priority)
	setIfNonZero(f, "completion_date", t.CompletionDate)
	setIfNonZero(f, "estimate_minutes", t.EstimateMinutes)
	setIfNonZero(f, "description", t.Description)
	setIfNonZero(f, "kind", t.Kind)
	setIfNonZero(f, "tag_color", t.TagColor)
	setIfNonZero(f, "tag_label", t.TagLabel)
	return f
}

// UpdatedAt alone is not an edit, but rides along with any real change so
// replicas converge on one display timestamp.
func taskChangedFields(old, updated domain.Task) map[string]any {
	f := map[string]any{}
	setIfChanged(f, "title", old.Title, updated.Title)
	setIfChanged(f, "status", old.Status, updated.Status)
	setIfChanged(f, "priority", old.Priority, updated.Priority)
	setIfChanged(f, "completion_date", old.CompletionDate, updated.CompletionDate)
	setIfChanged(f, "estimate_minutes", old.EstimateMinutes, updated.EstimateMinutes)
	setIfChanged(f, "description", old.Description, updated.Description)
	setIfChanged(f, "tag_color", old.TagColor, updated.TagColor)
	setIfChanged(f, "tag_label", old.TagLabel, updated.TagLabel)
	if len(f) == 0 {
		return nil
	}
	setIfChanged(f, "updated_at", old.UpdatedAt, updated.UpdatedAt)
	return f
}
