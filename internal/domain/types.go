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

var ErrDuplicateCategoryName = errors.New("category name already exists")

// Project is stored as a single JSON file.
type Project struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	CreatedAt  string     `json:"created_at"`
	UpdatedAt  string     `json:"updated_at"`
	Categories []Category `json:"categories"`
}

type Category struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at,omitempty"`
	EstimateMinutes int    `json:"estimate_minutes,omitempty"`
	Tasks           []Task `json:"tasks"`
}

type Task struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Status          string `json:"status"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
	Priority        string `json:"priority,omitempty"`
	CompletionDate  string `json:"completion_date,omitempty"`
	EstimateMinutes int    `json:"estimate_minutes,omitempty"`
	Description     string `json:"description,omitempty"`
}

var EstimatePresets = []int{0, 15, 30, 60, 120, 240, 480, 960, 1440, 2400}

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

func NewProject(name string) (Project, error) {
	id, err := NewID()
	if err != nil {
		return Project{}, err
	}
	now := NowTimestamp()
	project := Project{
		ID:         id,
		Name:       name,
		CreatedAt:  now,
		UpdatedAt:  now,
		Categories: []Category{},
	}
	return project, nil
}

func NewCategory(name string) (Category, error) {
	id, err := NewID()
	if err != nil {
		return Category{}, err
	}
	now := NowTimestamp()
	return Category{
		ID:        id,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
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

func (t *Task) IncreasePriority() bool {
	var newPriority string
	switch t.Priority {
	case PriorityLow:
		newPriority = PriorityMedium
	case PriorityMedium, "":
		newPriority = PriorityHigh
	case PriorityHigh:
		return false
	default:
		newPriority = PriorityMedium
	}
	t.Priority = newPriority
	t.UpdatedAt = NowTimestamp()
	return true
}

func (t *Task) DecreasePriority() bool {
	var newPriority string
	switch t.Priority {
	case PriorityHigh:
		newPriority = PriorityMedium
	case PriorityMedium, "":
		newPriority = PriorityLow
	case PriorityLow:
		return false
	default:
		newPriority = PriorityMedium
	}
	t.Priority = newPriority
	t.UpdatedAt = NowTimestamp()
	return true
}

func (t *Task) SetEstimate(minutes int) {
	t.EstimateMinutes = minutes
	t.UpdatedAt = NowTimestamp()
}

func (t *Task) CycleStatus() bool {
	var nextStatus string
	switch t.Status {
	case StatusTodo:
		nextStatus = StatusInProgress
	case StatusInProgress:
		nextStatus = StatusCompleted
	case StatusCompleted:
		nextStatus = StatusCancelled
	case StatusCancelled:
		nextStatus = StatusTodo
	default:
		nextStatus = StatusTodo
	}
	if nextStatus == t.Status {
		return false
	}
	_ = t.SetStatus(nextStatus)
	return true
}

type StatusCounts struct {
	Todo       int
	InProgress int
	Completed  int
	Cancelled  int
}

func (s StatusCounts) Total() int {
	return s.Todo + s.InProgress + s.Completed + s.Cancelled
}

func (c *Category) StatusCounts() StatusCounts {
	var counts StatusCounts
	for _, task := range c.Tasks {
		switch task.Status {
		case StatusTodo:
			counts.Todo++
		case StatusInProgress:
			counts.InProgress++
		case StatusCompleted:
			counts.Completed++
		case StatusCancelled:
			counts.Cancelled++
		}
	}
	return counts
}

func (p *Project) StatusCounts() StatusCounts {
	var counts StatusCounts
	for _, cat := range p.Categories {
		c := cat.StatusCounts()
		counts.Todo += c.Todo
		counts.InProgress += c.InProgress
		counts.Completed += c.Completed
		counts.Cancelled += c.Cancelled
	}
	return counts
}

func (c *Category) AggregateStatus() string {
	if len(c.Tasks) == 0 {
		return ""
	}
	allFinished := true
	hasInProgress := false
	hasFinished := false
	for _, t := range c.Tasks {
		if t.Status == StatusInProgress {
			hasInProgress = true
		}
		if t.Status == StatusCompleted || t.Status == StatusCancelled {
			hasFinished = true
		} else {
			allFinished = false
		}
	}
	if hasInProgress {
		return StatusInProgress
	}
	if allFinished {
		return StatusCompleted
	}
	if hasFinished {
		return StatusInProgress
	}
	return StatusTodo
}

func (c *Category) SetEstimate(minutes int) {
	c.EstimateMinutes = minutes
	c.UpdatedAt = NowTimestamp()
}

func (c *Category) AddTask(task Task) {
	c.Tasks = append(c.Tasks, task)
	c.UpdatedAt = NowTimestamp()
}

func (c *Category) InsertTask(index int, task Task) {
	if index < 0 || index > len(c.Tasks) {
		index = len(c.Tasks)
	}
	c.Tasks = append(c.Tasks, Task{})
	copy(c.Tasks[index+1:], c.Tasks[index:])
	c.Tasks[index] = task
	c.UpdatedAt = NowTimestamp()
}

func (c *Category) RemoveTask(index int) error {
	if index < 0 || index >= len(c.Tasks) {
		return errors.New("task index out of range")
	}
	c.Tasks = append(c.Tasks[:index], c.Tasks[index+1:]...)
	c.UpdatedAt = NowTimestamp()
	return nil
}

func (p *Project) AddCategory(cat Category) {
	p.Categories = append(p.Categories, cat)
	p.UpdatedAt = NowTimestamp()
}

func (p *Project) InsertCategory(index int, cat Category) {
	if index < 0 || index > len(p.Categories) {
		index = len(p.Categories)
	}
	p.Categories = append(p.Categories, Category{})
	copy(p.Categories[index+1:], p.Categories[index:])
	p.Categories[index] = cat
	p.UpdatedAt = NowTimestamp()
}

func (p *Project) RemoveCategory(index int) error {
	if index < 0 || index >= len(p.Categories) {
		return errors.New("category index out of range")
	}
	p.Categories = append(p.Categories[:index], p.Categories[index+1:]...)
	p.UpdatedAt = NowTimestamp()
	return nil
}

func (p *Project) RenameCategory(index int, name string) error {
	if index < 0 || index >= len(p.Categories) {
		return errors.New("category index out of range")
	}
	normalized := NormalizeName(name)
	for i, cat := range p.Categories {
		if i != index && NormalizeName(cat.Name) == normalized {
			return ErrDuplicateCategoryName
		}
	}
	p.Categories[index].Name = name
	now := NowTimestamp()
	p.Categories[index].UpdatedAt = now
	p.UpdatedAt = now
	return nil
}

func (p *Project) AddCategoryNamed(name string) (*Category, error) {
	normalized := NormalizeName(name)
	for _, cat := range p.Categories {
		if NormalizeName(cat.Name) == normalized {
			return nil, ErrDuplicateCategoryName
		}
	}
	cat, err := NewCategory(name)
	if err != nil {
		return nil, err
	}
	p.AddCategory(cat)
	return &p.Categories[len(p.Categories)-1], nil
}

func (p *Project) MoveTask(srcCatIdx, taskIdx, dstCatIdx int) error {
	if srcCatIdx < 0 || srcCatIdx >= len(p.Categories) {
		return errors.New("source category index out of range")
	}
	if dstCatIdx < 0 || dstCatIdx >= len(p.Categories) {
		return errors.New("destination category index out of range")
	}
	if srcCatIdx == dstCatIdx {
		return errors.New("source and destination categories are the same")
	}
	srcTasks := p.Categories[srcCatIdx].Tasks
	if taskIdx < 0 || taskIdx >= len(srcTasks) {
		return errors.New("task index out of range")
	}
	task := srcTasks[taskIdx]
	if err := p.Categories[srcCatIdx].RemoveTask(taskIdx); err != nil {
		return err
	}
	p.Categories[dstCatIdx].AddTask(task)
	p.UpdatedAt = NowTimestamp()
	return nil
}
