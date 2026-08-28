package data

import (
	"sync"
	"time"
)

// cursorSaveInterval bounds both how stale the cursor on disk can get and how
// often state.json is rewritten while the user navigates.
const cursorSaveInterval = 500 * time.Millisecond

// CursorSaver persists cursor positions in the background. The TUI records the
// focused row as it moves, which is what makes the cursor survive a force quit —
// SIGKILL, a closed terminal, a crash — none of which reach the save on exit.
// Writing straight through would put an fsync on every keystroke, so moves
// coalesce: only the latest cursor per project is kept, and one write covers a
// whole burst of navigation.
//
// Writes go through StateRepository, which reloads and merges before writing,
// so they stay safe against `phasionary serve` touching the same file.
type CursorSaver struct {
	repo     StateRepository
	interval time.Duration

	mu      sync.Mutex
	pending map[string]Cursor

	quit chan struct{}
	done chan struct{}
}

// NewCursorSaver starts the background writer. Call Close to stop it and flush.
func NewCursorSaver(repo StateRepository) *CursorSaver {
	return newCursorSaver(repo, cursorSaveInterval)
}

func newCursorSaver(repo StateRepository, interval time.Duration) *CursorSaver {
	s := &CursorSaver{
		repo:     repo,
		interval: interval,
		pending:  make(map[string]Cursor),
		quit:     make(chan struct{}),
		done:     make(chan struct{}),
	}
	go s.run()
	return s
}

func (s *CursorSaver) run() {
	defer close(s.done)
	tick := time.NewTicker(s.interval)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			s.flush()
		case <-s.quit:
			s.flush()
			return
		}
	}
}

// Set queues a cursor for a project, replacing any cursor not yet written. A
// zero Cursor is queued like any other: it clears the stored entry.
func (s *CursorSaver) Set(projectID string, cursor Cursor) {
	if projectID == "" {
		return
	}
	s.mu.Lock()
	s.pending[projectID] = cursor
	s.mu.Unlock()
}

// Drop discards a project's queued cursor. Call it once the stored entry has
// been settled another way, so the queued value cannot land on top of it —
// deleting a project would otherwise resurrect its cursor.
func (s *CursorSaver) Drop(projectID string) {
	s.mu.Lock()
	delete(s.pending, projectID)
	s.mu.Unlock()
}

func (s *CursorSaver) flush() {
	s.mu.Lock()
	pending := s.pending
	s.pending = make(map[string]Cursor)
	s.mu.Unlock()

	for id, cursor := range pending {
		// A failed cursor write is not worth reporting: the row is a
		// convenience, and there is nowhere here to surface it. The next session
		// starts on the first task.
		_ = s.repo.SetCursor(id, cursor)
	}
}

// Close stops the writer and flushes what is still queued, so a clean exit never
// drops the last move.
func (s *CursorSaver) Close() {
	close(s.quit)
	<-s.done
}
