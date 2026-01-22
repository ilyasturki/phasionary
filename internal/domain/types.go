package domain

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"time"
)

const (
	StatusTodo       = "todo"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
	StatusCancelled  = "cancelled"

	SectionCurrent = "current"
	SectionFuture  = "future"
	SectionPast    = "past"

	PriorityHigh   = "high"
	PriorityMedium = "medium"
	PriorityLow    = "low"
)

const (
	EstimateMinutes = "minutes"
	EstimateHours   = "hours"
	EstimateDays    = "days"
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
	ID                string `json:"id"`
	Title             string `json:"title"`
	Description       string `json:"description,omitempty"`
	Status            string `json:"status"`
	Section           string `json:"section"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
	Deadline          string `json:"deadline,omitempty"`
	TimeEstimateValue int    `json:"time_estimate_value,omitempty"`
	TimeEstimateUnit  string `json:"time_estimate_unit,omitempty"`
	Priority          string `json:"priority,omitempty"`
	Notes             string `json:"notes,omitempty"`
	CompletionDate    string `json:"completion_date,omitempty"`
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
		Section:   SectionCurrent,
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

func ValidateSection(section string) error {
	switch section {
	case SectionCurrent, SectionFuture, SectionPast:
		return nil
	default:
		return errors.New("invalid section")
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

func SortTasks(tasks []Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		a, b := tasks[i], tasks[j]
		if rankPriority(a.Priority) != rankPriority(b.Priority) {
			return rankPriority(a.Priority) < rankPriority(b.Priority)
		}
		if dlA, okA := parseDeadline(a.Deadline); okA {
			if dlB, okB := parseDeadline(b.Deadline); okB {
				if !dlA.Equal(dlB) {
					return dlA.Before(dlB)
				}
			} else {
				return true
			}
		} else if _, okB := parseDeadline(b.Deadline); okB {
			return false
		}
		if estA, okA := estimateMinutes(a.TimeEstimateValue, a.TimeEstimateUnit); okA {
			if estB, okB := estimateMinutes(b.TimeEstimateValue, b.TimeEstimateUnit); okB {
				if estA != estB {
					return estA < estB
				}
			} else {
				return true
			}
		} else if _, okB := estimateMinutes(b.TimeEstimateValue, b.TimeEstimateUnit); okB {
			return false
		}
		return strings.ToLower(a.Title) < strings.ToLower(b.Title)
	})
}

func rankPriority(priority string) int {
	switch priority {
	case PriorityHigh:
		return 0
	case PriorityMedium:
		return 1
	case PriorityLow:
		return 2
	case "":
		return 3
	default:
		return 4
	}
}

func parseDeadline(deadline string) (time.Time, bool) {
	if strings.TrimSpace(deadline) == "" {
		return time.Time{}, false
	}
	if t, err := time.Parse("2006-01-02", deadline); err == nil {
		return t, true
	}
	if t, err := time.Parse(time.RFC3339, deadline); err == nil {
		return t, true
	}
	return time.Time{}, false
}

func estimateMinutes(value int, unit string) (int, bool) {
	if value <= 0 {
		return 0, false
	}
	switch unit {
	case EstimateMinutes:
		return value, true
	case EstimateHours:
		return value * 60, true
	case EstimateDays:
		return value * 60 * 24, true
	default:
		return 0, false
	}
}
