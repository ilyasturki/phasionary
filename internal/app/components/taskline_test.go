package components

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"phasionary/internal/domain"
)

func TestNewTaskLineRenderer(t *testing.T) {
	renderer := NewTaskLineRenderer(80, "text")
	assert.NotNil(t, renderer)
	assert.Equal(t, 80, renderer.width)
	assert.Equal(t, "text", renderer.statusDisplay)
}

func TestTaskLineRenderer_Render(t *testing.T) {
	t.Run("renders unselected task", func(t *testing.T) {
		renderer := NewTaskLineRenderer(0, "text")
		task := domain.Task{
			Title:    "Test task",
			Status:   domain.StatusTodo,
			Priority: domain.PriorityMedium,
		}
		result := renderer.Render(task, false)
		assert.Contains(t, result, "Test task")
		assert.Contains(t, result, "todo")
		assert.True(t, strings.HasPrefix(result, "  "))
	})

	t.Run("renders selected task with cursor prefix", func(t *testing.T) {
		renderer := NewTaskLineRenderer(0, "text")
		task := domain.Task{
			Title:    "Test task",
			Status:   domain.StatusInProgress,
			Priority: domain.PriorityHigh,
		}
		result := renderer.Render(task, true)
		assert.Contains(t, result, "Test task")
		assert.Contains(t, result, "progress")
	})

	t.Run("wraps long task titles", func(t *testing.T) {
		renderer := NewTaskLineRenderer(40, "text")
		task := domain.Task{
			Title:  "This is a very long task title that should wrap to multiple lines",
			Status: domain.StatusTodo,
		}
		result := renderer.Render(task, false)
		lines := strings.Split(result, "\n")
		assert.True(t, len(lines) > 1)
	})

	t.Run("renders with icon status display", func(t *testing.T) {
		renderer := NewTaskLineRenderer(0, "icons")
		task := domain.Task{
			Title:  "Test task",
			Status: domain.StatusCompleted,
		}
		result := renderer.Render(task, false)
		assert.Contains(t, result, "[")
		assert.Contains(t, result, "]")
	})
}

func TestTaskLineRenderer_StatusLabel(t *testing.T) {
	t.Run("returns text labels when not icons mode", func(t *testing.T) {
		renderer := NewTaskLineRenderer(0, "text")
		assert.Contains(t, renderer.statusLabel(domain.StatusTodo), "todo")
		assert.Contains(t, renderer.statusLabel(domain.StatusInProgress), "progress")
		assert.Contains(t, renderer.statusLabel(domain.StatusCompleted), "completed")
		assert.Contains(t, renderer.statusLabel(domain.StatusCancelled), "cancelled")
	})

	t.Run("returns icons when icons mode", func(t *testing.T) {
		renderer := NewTaskLineRenderer(0, "icons")
		assert.Equal(t, " ", renderer.statusLabel(domain.StatusTodo))
		assert.Equal(t, "/", renderer.statusLabel(domain.StatusInProgress))
		assert.Equal(t, "x", renderer.statusLabel(domain.StatusCompleted))
		assert.Equal(t, "-", renderer.statusLabel(domain.StatusCancelled))
	})
}

func TestSafeWidth(t *testing.T) {
	t.Run("returns available space", func(t *testing.T) {
		assert.Equal(t, 70, safeWidth(80, 10))
	})

	t.Run("returns minimum of 1 for negative result", func(t *testing.T) {
		assert.Equal(t, 1, safeWidth(10, 20))
	})
}
