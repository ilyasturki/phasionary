package journal

import "phasionary/internal/domain"

type Recorder struct {
	device  Device
	journal *Journal
}

// (nil, nil) means not enrolled; an unreadable device file is an error, never
// a silent fallback to local-only.
func OpenIfConfigured(stateDir string) (*Recorder, error) {
	d, ok, err := LoadDevice(stateDir)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return &Recorder{device: d, journal: Open(stateDir)}, nil
}

func (r *Recorder) RecordSave(old *domain.Project, updated domain.Project) error {
	return r.journal.Append(r.device.DeviceID, DiffProjects(old, updated))
}

// No child tombstones: the server cascades a project delete.
func (r *Recorder) RecordDelete(projectID string) error {
	return r.journal.Append(r.device.DeviceID, []Draft{{
		Kind:      KindProjectDelete,
		ProjectID: projectID,
	}})
}
