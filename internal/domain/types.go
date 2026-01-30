package domain

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

const (
	StatusTodo       = "todo"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
	StatusCancelled  = "cancelled"

	PriorityHigh   = "high"
	PriorityMedium = "medium"
	PriorityLow    = "low"
)

var DefaultCategories = []string{"Feature", "Fix", "Ergonomy", "Documentation", "Research"}

// Project is stored as a single JSON file.
type Project struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	CreatedAt   string     `json:"created_at"`
	UpdatedAt   string     `json:"updated_at"`
	Categories  []Category `json:"categories"`
}

type Category struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	Tasks     []Task `json:"tasks"`
}

type Task struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Status         string `json:"status"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	Priority       string `json:"priority,omitempty"`
	CompletionDate string `json:"completion_date,omitempty"`
}

func NowTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func NormalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func NewID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// Format as UUID-like string.
	parts := []string{
		hex.EncodeToString(b[0:4]),
		hex.EncodeToString(b[4:6]),
		hex.EncodeToString(b[6:8]),
		hex.EncodeToString(b[8:10]),
		hex.EncodeToString(b[10:16]),
	}
	return strings.Join(parts, "-"), nil
}

func NewProject(name string, description string) (Project, error) {
	id, err := NewID()
	if err != nil {
		return Project{}, err
	}
	now := NowTimestamp()
	project := Project{
		ID:          id,
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
		Categories:  []Category{},
	}
	return project, nil
}

func NewCategory(name string) (Category, error) {
	id, err := NewID()
	if err != nil {
		return Category{}, err
	}
	return Category{
		ID:        id,
		Name:      name,
		CreatedAt: NowTimestamp(),
		Tasks:     []Task{},
	}, nil
}

func NewTask(title string) (Task, error) {
	id, err := NewID()
	if err != nil {
		return Task{}, err
	}
	now := NowTimestamp()
	return Task{
		ID:        id,
		Title:     title,
		Status:    StatusTodo,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func ValidateStatus(status string) error {
	switch status {
	case StatusTodo, StatusInProgress, StatusCompleted, StatusCancelled:
		return nil
	default:
		return errors.New("invalid status")
	}
}

func ValidatePriority(priority string) error {
	switch priority {
	case "", PriorityHigh, PriorityMedium, PriorityLow:
		return nil
	default:
		return errors.New("invalid priority")
	}
}

func (t *Task) SetStatus(status string) error {
	if err := ValidateStatus(status); err != nil {
		return err
	}
	t.Status = status
	t.UpdatedAt = NowTimestamp()
	if status == StatusCompleted {
		t.CompletionDate = NowTimestamp()
	} else {
		t.CompletionDate = ""
	}
	return nil
}

func (t *Task) SetPriority(priority string) error {
	if err := ValidatePriority(priority); err != nil {
		return err
	}
	t.Priority = priority
	t.UpdatedAt = NowTimestamp()
	return nil
}

