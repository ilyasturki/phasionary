package operations

import (
	"strings"

	"phasionary/internal/domain"
)

// TaskFields carries the writable fields of a task. It is the single unit the
// CLI and web front doors both build and hand to CreateTask, so the two
// surfaces create tasks through identical logic.
type TaskFields struct {
	Title       string
	Status      string
	Priority    string
	Estimate    int
	Description string
	TagColor    string
	TagLabel    string
}

// CreateTask builds a task from f and appends it to the category identified by
// categoryID. An empty Status defaults to todo; Status and Priority, when set,
// are applied through the domain setters so their side effects (UpdatedAt,
// CompletionDate) and validation hold. Returns ErrTitleRequired, an invalid
// status/priority error, or domain.ErrCategoryNotFound.
func CreateTask(p *domain.Project, categoryID string, f TaskFields) (domain.Task, error) {
	if strings.TrimSpace(f.Title) == "" {
		return domain.Task{}, ErrTitleRequired
	}
	if f.Estimate < 0 {
		return domain.Task{}, ErrNegativeEstimate
	}
	idx, err := p.FindCategoryByID(categoryID)
	if err != nil {
		return domain.Task{}, err
	}

	task, err := domain.NewTask(f.Title)
	if err != nil {
		return domain.Task{}, err
	}
	if f.Status != "" {
		if err := task.SetStatus(f.Status); err != nil {
			return domain.Task{}, err
		}
	}
	if f.Priority != "" {
		if err := task.SetPriority(f.Priority); err != nil {
			return domain.Task{}, err
		}
	}
	task.SetEstimate(f.Estimate)
	task.Description = f.Description
	if f.TagColor != "" || f.TagLabel != "" {
		if err := task.SetTagColor(f.TagColor); err != nil {
			return domain.Task{}, err
		}
		task.SetTagLabel(f.TagLabel)
	}

	p.Categories[idx].AddTask(task)
	return task, nil
}

// findTask resolves categoryID→category and taskID→task within p, returning a
// pointer to the stored task so callers mutate it in place. Returns
// domain.ErrCategoryNotFound or domain.ErrTaskNotFound.
func findTask(p *domain.Project, categoryID, taskID string) (*domain.Task, error) {
	cidx, err := p.FindCategoryByID(categoryID)
	if err != nil {
		return nil, err
	}
	tidx, err := p.Categories[cidx].FindTaskByID(taskID)
	if err != nil {
		return nil, err
	}
	return &p.Categories[cidx].Tasks[tidx], nil
}

// SetTaskStatus applies status to the identified task through domain.SetStatus,
// so completion-date bookkeeping and validation hold.
func SetTaskStatus(p *domain.Project, categoryID, taskID, status string) (domain.Task, error) {
	task, err := findTask(p, categoryID, taskID)
	if err != nil {
		return domain.Task{}, err
	}
	if err := task.SetStatus(status); err != nil {
		return domain.Task{}, err
	}
	return *task, nil
}

// SetTaskPriority applies priority to the identified task through
// domain.SetPriority.
func SetTaskPriority(p *domain.Project, categoryID, taskID, priority string) (domain.Task, error) {
	task, err := findTask(p, categoryID, taskID)
	if err != nil {
		return domain.Task{}, err
	}
	if err := task.SetPriority(priority); err != nil {
		return domain.Task{}, err
	}
	return *task, nil
}

// SetTaskEstimate sets the identified task's estimate in minutes.
func SetTaskEstimate(p *domain.Project, categoryID, taskID string, minutes int) (domain.Task, error) {
	if minutes < 0 {
		return domain.Task{}, ErrNegativeEstimate
	}
	task, err := findTask(p, categoryID, taskID)
	if err != nil {
		return domain.Task{}, err
	}
	task.SetEstimate(minutes)
	return *task, nil
}

// MoveTaskWithinCategory shifts the identified task by delta positions inside
// its category. A move that would fall off either end changes nothing and
// returns ErrNoChange, so a boundary tap doesn't trigger a save.
func MoveTaskWithinCategory(p *domain.Project, categoryID, taskID string, delta int) (domain.Task, error) {
	cidx, err := p.FindCategoryByID(categoryID)
	if err != nil {
		return domain.Task{}, err
	}
	cat := &p.Categories[cidx]
	tidx, err := cat.FindTaskByID(taskID)
	if err != nil {
		return domain.Task{}, err
	}
	dst := tidx + delta
	if dst < 0 || dst >= len(cat.Tasks) {
		return cat.Tasks[tidx], ErrNoChange
	}
	if err := cat.MoveTask(tidx, delta); err != nil {
		return domain.Task{}, err
	}
	return cat.Tasks[dst], nil
}

// MoveTaskToCategory moves the identified task out of srcCatID and into
// dstCatID. Propagates domain.MoveTask's error when the two categories are the
// same or an id does not resolve.
func MoveTaskToCategory(p *domain.Project, srcCatID, taskID, dstCatID string) (domain.Task, error) {
	srcIdx, err := p.FindCategoryByID(srcCatID)
	if err != nil {
		return domain.Task{}, err
	}
	tidx, err := p.Categories[srcIdx].FindTaskByID(taskID)
	if err != nil {
		return domain.Task{}, err
	}
	dstIdx, err := p.FindCategoryByID(dstCatID)
	if err != nil {
		return domain.Task{}, err
	}
	task := p.Categories[srcIdx].Tasks[tidx]
	if err := p.MoveTask(srcIdx, tidx, dstIdx); err != nil {
		return domain.Task{}, err
	}
	return task, nil
}
