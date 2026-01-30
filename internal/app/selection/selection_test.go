package selection

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	positions := []Position{
		{Kind: FocusProject, CategoryIndex: -1, TaskIndex: -1},
		{Kind: FocusCategory, CategoryIndex: 0, TaskIndex: -1},
		{Kind: FocusTask, CategoryIndex: 0, TaskIndex: 0},
	}

	t.Run("initializes with valid selection", func(t *testing.T) {
		m := NewManager(positions, 1)
		assert.Equal(t, 1, m.Selected())
		assert.Equal(t, 3, m.Count())
	})

	t.Run("clamps negative selection to zero", func(t *testing.T) {
		m := NewManager(positions, -5)
		assert.Equal(t, 0, m.Selected())
	})

	t.Run("clamps overflow selection to last", func(t *testing.T) {
		m := NewManager(positions, 100)
		assert.Equal(t, 2, m.Selected())
	})

	t.Run("handles empty positions", func(t *testing.T) {
		m := NewManager(nil, 0)
		assert.Equal(t, -1, m.Selected())
		assert.True(t, m.IsEmpty())
	})
}

func TestManager_MoveBy(t *testing.T) {
	positions := []Position{
		{Kind: FocusProject},
		{Kind: FocusCategory},
		{Kind: FocusTask},
		{Kind: FocusTask},
	}

	t.Run("moves down", func(t *testing.T) {
		m := NewManager(positions, 0)
		moved := m.MoveBy(1)
		assert.True(t, moved)
		assert.Equal(t, 1, m.Selected())
	})

	t.Run("moves up", func(t *testing.T) {
		m := NewManager(positions, 2)
		moved := m.MoveBy(-1)
		assert.True(t, moved)
		assert.Equal(t, 1, m.Selected())
	})

	t.Run("clamps at top", func(t *testing.T) {
		m := NewManager(positions, 0)
		moved := m.MoveBy(-10)
		assert.False(t, moved)
		assert.Equal(t, 0, m.Selected())
	})

	t.Run("clamps at bottom", func(t *testing.T) {
		m := NewManager(positions, 3)
		moved := m.MoveBy(10)
		assert.False(t, moved)
		assert.Equal(t, 3, m.Selected())
	})

	t.Run("returns false on empty", func(t *testing.T) {
		m := NewManager(nil, 0)
		moved := m.MoveBy(1)
		assert.False(t, moved)
	})
}

func TestManager_MoveTo(t *testing.T) {
	positions := []Position{
		{Kind: FocusProject},
		{Kind: FocusCategory},
		{Kind: FocusTask},
	}

	t.Run("moves to valid index", func(t *testing.T) {
		m := NewManager(positions, 0)
		moved := m.MoveTo(2)
		assert.True(t, moved)
		assert.Equal(t, 2, m.Selected())
	})

	t.Run("clamps invalid index", func(t *testing.T) {
		m := NewManager(positions, 0)
		moved := m.MoveTo(100)
		assert.True(t, moved)
		assert.Equal(t, 2, m.Selected())
	})
}

func TestManager_JumpToFirst(t *testing.T) {
	positions := []Position{
		{Kind: FocusProject},
		{Kind: FocusCategory},
		{Kind: FocusTask},
	}

	t.Run("jumps to first", func(t *testing.T) {
		m := NewManager(positions, 2)
		moved := m.JumpToFirst()
		assert.True(t, moved)
		assert.Equal(t, 0, m.Selected())
	})

	t.Run("returns false when already at first", func(t *testing.T) {
		m := NewManager(positions, 0)
		moved := m.JumpToFirst()
		assert.False(t, moved)
	})
}

func TestManager_JumpToLast(t *testing.T) {
	positions := []Position{
		{Kind: FocusProject},
		{Kind: FocusCategory},
		{Kind: FocusTask},
	}

	t.Run("jumps to last", func(t *testing.T) {
		m := NewManager(positions, 0)
		moved := m.JumpToLast()
		assert.True(t, moved)
		assert.Equal(t, 2, m.Selected())
	})

	t.Run("returns false when already at last", func(t *testing.T) {
		m := NewManager(positions, 2)
		moved := m.JumpToLast()
		assert.False(t, moved)
	})
}

func TestManager_SelectedPosition(t *testing.T) {
	positions := []Position{
		{Kind: FocusProject, CategoryIndex: -1, TaskIndex: -1},
		{Kind: FocusTask, CategoryIndex: 0, TaskIndex: 2},
	}

	t.Run("returns selected position", func(t *testing.T) {
		m := NewManager(positions, 1)
		pos, ok := m.SelectedPosition()
		require.True(t, ok)
		assert.Equal(t, FocusTask, pos.Kind)
		assert.Equal(t, 0, pos.CategoryIndex)
		assert.Equal(t, 2, pos.TaskIndex)
	})

	t.Run("returns false when empty", func(t *testing.T) {
		m := NewManager(nil, 0)
		_, ok := m.SelectedPosition()
		assert.False(t, ok)
	})
}

func TestManager_FindPositionIndex(t *testing.T) {
	positions := []Position{
		{Kind: FocusProject},
		{Kind: FocusCategory, CategoryIndex: 0},
		{Kind: FocusTask, CategoryIndex: 0, TaskIndex: 0},
		{Kind: FocusTask, CategoryIndex: 0, TaskIndex: 1},
		{Kind: FocusCategory, CategoryIndex: 1},
	}

	t.Run("finds matching position", func(t *testing.T) {
		m := NewManager(positions, 0)
		idx := m.FindPositionIndex(func(p Position) bool {
			return p.Kind == FocusTask && p.TaskIndex == 1
		})
		assert.Equal(t, 3, idx)
	})

	t.Run("returns -1 when not found", func(t *testing.T) {
		m := NewManager(positions, 0)
		idx := m.FindPositionIndex(func(p Position) bool {
			return p.Kind == FocusTask && p.TaskIndex == 99
		})
		assert.Equal(t, -1, idx)
	})
}

func TestManager_SelectByPredicate(t *testing.T) {
	positions := []Position{
		{Kind: FocusProject},
		{Kind: FocusCategory, CategoryIndex: 1},
		{Kind: FocusTask, CategoryIndex: 1, TaskIndex: 0},
	}

	t.Run("selects matching position", func(t *testing.T) {
		m := NewManager(positions, 0)
		ok := m.SelectByPredicate(func(p Position) bool {
			return p.Kind == FocusCategory && p.CategoryIndex == 1
		})
		assert.True(t, ok)
		assert.Equal(t, 1, m.Selected())
	})

	t.Run("returns false when not found", func(t *testing.T) {
		m := NewManager(positions, 0)
		ok := m.SelectByPredicate(func(p Position) bool {
			return p.Kind == FocusCategory && p.CategoryIndex == 99
		})
		assert.False(t, ok)
		assert.Equal(t, 0, m.Selected())
	})
}

func TestManager_SetPositions(t *testing.T) {
	positions := []Position{
		{Kind: FocusProject},
		{Kind: FocusTask},
	}

	t.Run("updates positions and clamps selection", func(t *testing.T) {
		m := NewManager(positions, 1)
		newPositions := []Position{{Kind: FocusProject}}
		m.SetPositions(newPositions)
		assert.Equal(t, 1, m.Count())
		assert.Equal(t, 0, m.Selected())
	})
}
