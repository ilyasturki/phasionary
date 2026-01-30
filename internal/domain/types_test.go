package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTask_IncreasePriority(t *testing.T) {
	t.Run("increases from low to medium", func(t *testing.T) {
		task := Task{Priority: PriorityLow}
		changed := task.IncreasePriority()
		assert.True(t, changed)
		assert.Equal(t, PriorityMedium, task.Priority)
	})

	t.Run("increases from medium to high", func(t *testing.T) {
		task := Task{Priority: PriorityMedium}
		changed := task.IncreasePriority()
		assert.True(t, changed)
		assert.Equal(t, PriorityHigh, task.Priority)
	})

	t.Run("cannot increase from high", func(t *testing.T) {
		task := Task{Priority: PriorityHigh}
		changed := task.IncreasePriority()
		assert.False(t, changed)
		assert.Equal(t, PriorityHigh, task.Priority)
	})

	t.Run("empty priority defaults to medium", func(t *testing.T) {
		task := Task{Priority: ""}
		changed := task.IncreasePriority()
		assert.True(t, changed)
		assert.Equal(t, PriorityMedium, task.Priority)
	})
}

func TestTask_DecreasePriority(t *testing.T) {
	t.Run("decreases from high to medium", func(t *testing.T) {
		task := Task{Priority: PriorityHigh}
		changed := task.DecreasePriority()
		assert.True(t, changed)
		assert.Equal(t, PriorityMedium, task.Priority)
	})

	t.Run("decreases from medium to low", func(t *testing.T) {
		task := Task{Priority: PriorityMedium}
		changed := task.DecreasePriority()
		assert.True(t, changed)
		assert.Equal(t, PriorityLow, task.Priority)
	})

	t.Run("cannot decrease from low", func(t *testing.T) {
		task := Task{Priority: PriorityLow}
		changed := task.DecreasePriority()
		assert.False(t, changed)
		assert.Equal(t, PriorityLow, task.Priority)
	})

	t.Run("empty priority defaults to medium", func(t *testing.T) {
		task := Task{Priority: ""}
		changed := task.DecreasePriority()
		assert.True(t, changed)
		assert.Equal(t, PriorityMedium, task.Priority)
	})
}

func TestTask_CycleStatus(t *testing.T) {
	t.Run("cycles from todo to in_progress", func(t *testing.T) {
		task := Task{Status: StatusTodo}
		changed := task.CycleStatus()
		assert.True(t, changed)
		assert.Equal(t, StatusInProgress, task.Status)
	})

	t.Run("cycles from in_progress to completed", func(t *testing.T) {
		task := Task{Status: StatusInProgress}
		changed := task.CycleStatus()
		assert.True(t, changed)
		assert.Equal(t, StatusCompleted, task.Status)
	})

	t.Run("cycles from completed to todo", func(t *testing.T) {
		task := Task{Status: StatusCompleted}
		changed := task.CycleStatus()
		assert.True(t, changed)
		assert.Equal(t, StatusTodo, task.Status)
	})

	t.Run("cycles from cancelled to todo", func(t *testing.T) {
		task := Task{Status: StatusCancelled}
		changed := task.CycleStatus()
		assert.True(t, changed)
		assert.Equal(t, StatusTodo, task.Status)
	})
}

func TestCategory_AddTask(t *testing.T) {
	t.Run("adds task to category", func(t *testing.T) {
		cat := Category{Tasks: []Task{}}
		task := Task{ID: "test-1", Title: "Test Task"}
		cat.AddTask(task)
		assert.Len(t, cat.Tasks, 1)
		assert.Equal(t, "test-1", cat.Tasks[0].ID)
	})
}

func TestCategory_RemoveTask(t *testing.T) {
	t.Run("removes task at valid index", func(t *testing.T) {
		cat := Category{Tasks: []Task{
			{ID: "1"}, {ID: "2"}, {ID: "3"},
		}}
		err := cat.RemoveTask(1)
		require.NoError(t, err)
		assert.Len(t, cat.Tasks, 2)
		assert.Equal(t, "1", cat.Tasks[0].ID)
		assert.Equal(t, "3", cat.Tasks[1].ID)
	})

	t.Run("returns error for negative index", func(t *testing.T) {
		cat := Category{Tasks: []Task{{ID: "1"}}}
		err := cat.RemoveTask(-1)
		assert.Error(t, err)
	})

	t.Run("returns error for index out of range", func(t *testing.T) {
		cat := Category{Tasks: []Task{{ID: "1"}}}
		err := cat.RemoveTask(5)
		assert.Error(t, err)
	})
}

func TestProject_AddCategory(t *testing.T) {
	t.Run("adds category to project", func(t *testing.T) {
		proj := Project{Categories: []Category{}}
		cat := Category{ID: "cat-1", Name: "Test"}
		proj.AddCategory(cat)
		assert.Len(t, proj.Categories, 1)
		assert.Equal(t, "cat-1", proj.Categories[0].ID)
	})
}

func TestProject_InsertCategory(t *testing.T) {
	t.Run("inserts category at specified index", func(t *testing.T) {
		proj := Project{Categories: []Category{
			{ID: "1"}, {ID: "3"},
		}}
		cat := Category{ID: "2"}
		proj.InsertCategory(1, cat)
		assert.Len(t, proj.Categories, 3)
		assert.Equal(t, "1", proj.Categories[0].ID)
		assert.Equal(t, "2", proj.Categories[1].ID)
		assert.Equal(t, "3", proj.Categories[2].ID)
	})

	t.Run("appends at end for out of range index", func(t *testing.T) {
		proj := Project{Categories: []Category{{ID: "1"}}}
		cat := Category{ID: "2"}
		proj.InsertCategory(100, cat)
		assert.Len(t, proj.Categories, 2)
		assert.Equal(t, "2", proj.Categories[1].ID)
	})

	t.Run("inserts at end for negative index (clamped)", func(t *testing.T) {
		proj := Project{Categories: []Category{{ID: "1"}}}
		cat := Category{ID: "2"}
		proj.InsertCategory(-1, cat)
		assert.Len(t, proj.Categories, 2)
		assert.Equal(t, "2", proj.Categories[1].ID)
	})
}

func TestProject_RemoveCategory(t *testing.T) {
	t.Run("removes category at valid index", func(t *testing.T) {
		proj := Project{Categories: []Category{
			{ID: "1"}, {ID: "2"}, {ID: "3"},
		}}
		err := proj.RemoveCategory(1)
		require.NoError(t, err)
		assert.Len(t, proj.Categories, 2)
		assert.Equal(t, "1", proj.Categories[0].ID)
		assert.Equal(t, "3", proj.Categories[1].ID)
	})

	t.Run("returns error for negative index", func(t *testing.T) {
		proj := Project{Categories: []Category{{ID: "1"}}}
		err := proj.RemoveCategory(-1)
		assert.Error(t, err)
	})

	t.Run("returns error for index out of range", func(t *testing.T) {
		proj := Project{Categories: []Category{{ID: "1"}}}
		err := proj.RemoveCategory(5)
		assert.Error(t, err)
	})
}

func TestTask_SetStatus(t *testing.T) {
	t.Run("sets status and updates timestamp", func(t *testing.T) {
		task := Task{Status: StatusTodo}
		err := task.SetStatus(StatusInProgress)
		require.NoError(t, err)
		assert.Equal(t, StatusInProgress, task.Status)
		assert.NotEmpty(t, task.UpdatedAt)
	})

	t.Run("sets completion date when completed", func(t *testing.T) {
		task := Task{Status: StatusTodo}
		err := task.SetStatus(StatusCompleted)
		require.NoError(t, err)
		assert.NotEmpty(t, task.CompletionDate)
	})

	t.Run("clears completion date when not completed", func(t *testing.T) {
		task := Task{Status: StatusCompleted, CompletionDate: "2024-01-01"}
		err := task.SetStatus(StatusTodo)
		require.NoError(t, err)
		assert.Empty(t, task.CompletionDate)
	})

	t.Run("returns error for invalid status", func(t *testing.T) {
		task := Task{Status: StatusTodo}
		err := task.SetStatus("invalid")
		assert.Error(t, err)
	})
}

func TestTask_SetPriority(t *testing.T) {
	t.Run("sets priority and updates timestamp", func(t *testing.T) {
		task := Task{Priority: ""}
		err := task.SetPriority(PriorityHigh)
		require.NoError(t, err)
		assert.Equal(t, PriorityHigh, task.Priority)
		assert.NotEmpty(t, task.UpdatedAt)
	})

	t.Run("returns error for invalid priority", func(t *testing.T) {
		task := Task{Priority: ""}
		err := task.SetPriority("invalid")
		assert.Error(t, err)
	})
}
