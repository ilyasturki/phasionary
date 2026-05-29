package data

import (
	"errors"
	"io/fs"
	"os"
	"syscall"

	"phasionary/internal/domain"
)

// SaveProjectLocked saves a project while holding an exclusive flock on a
// per-project lock file. Use this when multiple processes may write to the
// same data directory (e.g. TUI + `phasionary serve` at the same time).
//
// The lock file lives next to the project JSON at `{id}.json.lock` and is
// kept on disk between calls; only the flock state matters. Closing the fd
// releases the advisory lock, so an explicit LOCK_UN isn't needed.
func (s *Store) SaveProjectLocked(project domain.Project) error {
	f, err := s.acquireProjectLock(project.ID)
	if err != nil {
		return err
	}
	defer f.Close()
	return s.saveProjectAtomic(project)
}

// WithProjectLocked acquires the project flock, loads the project, runs fn,
// then saves — all under the same lock so concurrent processes cannot lose
// each other's updates between read and write.
//
// Returns ErrProjectNotFound without creating a .lock file when the project
// does not exist; this avoids unauthenticated 404 spam littering the data dir
// with empty lock files.
func (s *Store) WithProjectLocked(id string, fn func(*domain.Project) error) (domain.Project, error) {
	if _, err := os.Stat(s.projectPath(id)); errors.Is(err, fs.ErrNotExist) {
		return domain.Project{}, ErrProjectNotFound
	}
	f, err := s.acquireProjectLock(id)
	if err != nil {
		return domain.Project{}, err
	}
	defer f.Close()
	// Re-check after locking: another process may have deleted the project
	// between the existence check and the flock acquisition.
	project, err := s.LoadProjectByID(id)
	if err != nil {
		return domain.Project{}, err
	}
	if err := fn(&project); err != nil {
		return project, err
	}
	if err := s.saveProjectAtomic(project); err != nil {
		return project, err
	}
	return project, nil
}

func (s *Store) acquireProjectLock(id string) (*os.File, error) {
	if err := os.MkdirAll(s.Dir, 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(s.lockPath(id), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		_ = f.Close()
		return nil, err
	}
	return f, nil
}
