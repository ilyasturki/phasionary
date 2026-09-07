package journal

import (
	"os"

	"phasionary/internal/domain"
)

type Recorder struct {
	stateDir string
}

func NewRecorder(stateDir string) *Recorder { return &Recorder{stateDir: stateDir} }

func (r *Recorder) Active() bool {
	_, err := os.Stat(devicePath(r.stateDir))
	return err == nil
}

// Missing device file → no-op; unreadable → error, never a silent fallback
// to local-only.
func (r *Recorder) append(drafts []Draft) error {
	d, ok, err := LoadDevice(r.stateDir)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	return Open(r.stateDir).Append(d.DeviceID, drafts)
}

func (r *Recorder) RecordSave(old *domain.Project, updated domain.Project) error {
	return r.append(DiffProjects(old, updated))
}

// No child tombstones: the server cascades a project delete.
func (r *Recorder) RecordDelete(projectID string) error {
	return r.append([]Draft{{Kind: KindProjectDelete, ProjectID: projectID}})
}
