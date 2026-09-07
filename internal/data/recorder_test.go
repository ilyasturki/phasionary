package data

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"phasionary/internal/domain"
	"phasionary/internal/journal"
)

type fakeRecorder struct {
	saves   []bool
	deletes []string
	fail    error
}

func (f *fakeRecorder) Active() bool { return true }

func (f *fakeRecorder) RecordSave(old *domain.Project, updated domain.Project) error {
	if f.fail != nil {
		return f.fail
	}
	f.saves = append(f.saves, old != nil)
	return nil
}

func (f *fakeRecorder) RecordDelete(projectID string) error {
	if f.fail != nil {
		return f.fail
	}
	f.deletes = append(f.deletes, projectID)
	return nil
}

func TestRecorderSeesEveryWritePath(t *testing.T) {
	store := NewStore(t.TempDir())
	require.NoError(t, store.Ensure())
	rec := &fakeRecorder{}
	store.SetRecorder(rec)

	project, err := store.CreateProject("Journaled")
	require.NoError(t, err)
	require.Len(t, rec.saves, 1)

	_, err = store.WithProjectLocked(project.ID, func(p *domain.Project) error {
		p.Name = "Journaled" // no-op edit: the diff, not the store, decides what's an op
		return nil
	})
	require.NoError(t, err)
	require.Len(t, rec.saves, 2)

	require.NoError(t, store.SaveProjectLocked(project))
	require.Len(t, rec.saves, 3, "the async-saver path (WriteProjectLocked) records too")
	assert.Equal(t, []bool{false, true, true}, rec.saves, "only the create has no prior state")

	require.NoError(t, store.DeleteProject(project.ID))
	assert.Equal(t, []string{project.ID}, rec.deletes)
}

func TestRecorderFailureAbortsSave(t *testing.T) {
	store := NewStore(t.TempDir())
	require.NoError(t, store.Ensure())

	project, err := store.CreateProject("Precious")
	require.NoError(t, err)

	store.SetRecorder(&fakeRecorder{fail: errors.New("journal disk gone")})
	_, err = store.RenameProject(project.ID, "Renamed")
	require.Error(t, err)

	reloaded, err2 := store.LoadProjectByID(project.ID)
	require.NoError(t, err2)
	assert.Equal(t, "Precious", reloaded.Name, "the file must not change when journaling failed")
}

func TestStoreWritesReachJournal(t *testing.T) {
	stateDir := t.TempDir()
	device, err := journal.MintDevice(stateDir, "", "")
	require.NoError(t, err)
	store := NewStore(t.TempDir())
	require.NoError(t, store.Ensure())
	store.SetRecorder(journal.NewRecorder(stateDir))

	project, err := store.CreateProject("Synced")
	require.NoError(t, err)
	_, err = store.WithProjectLocked(project.ID, func(p *domain.Project) error {
		_, err := p.AddCategoryNamed("Extra")
		return err
	})
	require.NoError(t, err)
	require.NoError(t, store.DeleteProject(project.ID))

	ops, err := journal.Open(stateDir).Entries()
	require.NoError(t, err)
	require.NotEmpty(t, ops)
	assert.Equal(t, journal.KindProjectCreate, ops[0].Kind)
	assert.Equal(t, journal.KindProjectDelete, ops[len(ops)-1].Kind)
	for i, op := range ops {
		assert.Equal(t, device.DeviceID, op.DeviceID)
		assert.Equal(t, uint64(i+1), op.Seq, "seqs are gapless across store operations")
		assert.Equal(t, project.ID, op.ProjectID)
	}
}
