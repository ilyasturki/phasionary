package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"phasionary/internal/domain"
)

func TestComputeVisibility_ScrollOffsetZero_ProjectVisible(t *testing.T) {
	// Create a simple project with categories and tasks
	project := domain.Project{
		Name: "Test Project",
		Categories: []domain.Category{
			{
				Name: "Category 1",
				Tasks: []domain.Task{
					{Title: "Task 1", Status: domain.StatusTodo},
					{Title: "Task 2", Status: domain.StatusTodo},
				},
			},
		},
	}

	// Build positions (mimicking what the app does)
	positions := []focusPosition{
		{Kind: focusProject, CategoryIndex: -1, TaskIndex: -1},
		{Kind: focusCategory, CategoryIndex: 0, TaskIndex: -1},
		{Kind: focusTask, CategoryIndex: 0, TaskIndex: 0},
		{Kind: focusTask, CategoryIndex: 0, TaskIndex: 1},
	}

	// Build layout
	builder := NewLayoutBuilder(DefaultLayoutConfig(), 80, "icons", nil, nil)
	layout := builder.Build(project, positions)

	// Create viewport with reasonable screen height
	viewport := NewViewport(&layout, 20, DefaultLayoutConfig())

	// Compute visibility with scrollOffset = 0
	viewport.ComputeVisibility(0)

	// Verify the first item is a project
	require.True(t, len(layout.Items) > 0, "Layout should have items")
	assert.Equal(t, LayoutProject, layout.Items[0].Kind, "First item should be project")
	assert.Equal(t, 0, layout.Items[0].PositionIndex, "Project should have PositionIndex=0")

	// Key assertion: VisibleStart should be 0 when scrollOffset is 0
	assert.Equal(t, 0, viewport.VisibleStart, "VisibleStart should be 0 when scrollOffset is 0")
	assert.False(t, viewport.HasMoreAbove, "HasMoreAbove should be false when scrollOffset is 0")
}

func TestComputeVisibility_ZeroHeight_ProjectStillVisible(t *testing.T) {
	// This simulates the startup condition where Height is 0
	project := domain.Project{
		Name: "Test Project",
		Categories: []domain.Category{
			{
				Name: "Category 1",
				Tasks: []domain.Task{
					{Title: "Task 1", Status: domain.StatusTodo},
				},
			},
		},
	}

	positions := []focusPosition{
		{Kind: focusProject, CategoryIndex: -1, TaskIndex: -1},
		{Kind: focusCategory, CategoryIndex: 0, TaskIndex: -1},
		{Kind: focusTask, CategoryIndex: 0, TaskIndex: 0},
	}

	builder := NewLayoutBuilder(DefaultLayoutConfig(), 80, "icons", nil, nil)
	layout := builder.Build(project, positions)

	// Create viewport with 0 height (startup condition)
	viewport := NewViewport(&layout, 0, DefaultLayoutConfig())
	viewport.ComputeVisibility(0)

	// Even with zero height, VisibleStart should be 0
	assert.Equal(t, 0, viewport.VisibleStart, "VisibleStart should be 0 even with zero height")

	// VisibleEnd should be at least 1 (at least the project should be visible)
	assert.GreaterOrEqual(t, viewport.VisibleEnd, 1, "VisibleEnd should be at least 1")
}

func TestComputeVisibility_ScrollOffsetNonZero(t *testing.T) {
	project := domain.Project{
		Name: "Test Project",
		Categories: []domain.Category{
			{
				Name: "Category 1",
				Tasks: []domain.Task{
					{Title: "Task 1", Status: domain.StatusTodo},
					{Title: "Task 2", Status: domain.StatusTodo},
				},
			},
		},
	}

	positions := []focusPosition{
		{Kind: focusProject, CategoryIndex: -1, TaskIndex: -1},
		{Kind: focusCategory, CategoryIndex: 0, TaskIndex: -1},
		{Kind: focusTask, CategoryIndex: 0, TaskIndex: 0},
		{Kind: focusTask, CategoryIndex: 0, TaskIndex: 1},
	}

	builder := NewLayoutBuilder(DefaultLayoutConfig(), 80, "icons", nil, nil)
	layout := builder.Build(project, positions)
	viewport := NewViewport(&layout, 20, DefaultLayoutConfig())

	// Compute visibility with scrollOffset = 1 (skip project)
	viewport.ComputeVisibility(1)

	// The visible start should NOT be 0 (the project should be scrolled past)
	assert.True(t, viewport.HasMoreAbove, "HasMoreAbove should be true when scrollOffset > 0")
	// VisibleStart should point to the first item that has PositionIndex >= scrollOffset
	assert.Greater(t, viewport.VisibleStart, 0, "VisibleStart should be > 0 when scrollOffset is 1")
}
