package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"phasionary/internal/domain"
)

func twoStores(t *testing.T) (*Store, *Store, domain.Project) {
	t.Helper()
	dir := t.TempDir()
	a, b := NewStore(dir), NewStore(dir)
	p, err := a.CreateProject("Shared")
	require.NoError(t, err)
	return a, b, p
}

func TestSnapshotSaveRefusedAfterExternalWrite(t *testing.T) {
	a, b, p := twoStores(t)
	mine, err := a.LoadProjectByID(p.ID)
	require.NoError(t, err)

	_, err = b.WithProjectLocked(p.ID, func(p *domain.Project) error {
		p.Name = "Renamed elsewhere"
		return nil
	})
	require.NoError(t, err)

	// The picker lists every project while the open one holds unsaved edits:
	// listing must not move the baseline or this save would go through.
	_, err = a.ListProjects()
	require.NoError(t, err)

	mine.Name = "Renamed here"
	err = a.SaveProjectLocked(mine)
	require.ErrorIs(t, err, ErrStaleProject)

	onDisk, err := b.LoadProjectByID(p.ID)
	require.NoError(t, err)
	assert.Equal(t, "Renamed elsewhere", onDisk.Name, "the refused save must not touch the file")

	fresh, err := a.LoadProjectByID(p.ID)
	require.NoError(t, err)
	fresh.Name = "Renamed here"
	require.NoError(t, a.SaveProjectLocked(fresh))
}

func TestOwnSavesStayFresh(t *testing.T) {
	a, _, p := twoStores(t)
	mine, err := a.LoadProjectByID(p.ID)
	require.NoError(t, err)
	require.NoError(t, a.SaveProjectLocked(mine))
	require.NoError(t, a.SaveProjectLocked(mine), "a write adopts the version it wrote")
}

func TestWithProjectLockedUnaffectedByExternalWrite(t *testing.T) {
	a, b, p := twoStores(t)
	_, err := a.LoadProjectByID(p.ID)
	require.NoError(t, err)
	_, err = b.WithProjectLocked(p.ID, func(p *domain.Project) error {
		p.Name = "B"
		return nil
	})
	require.NoError(t, err)
	got, err := a.WithProjectLocked(p.ID, func(p *domain.Project) error {
		require.Equal(t, "B", p.Name)
		p.Name = "A after B"
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, "A after B", got.Name)
}

func TestUnknownProjectSavesWithoutBaseline(t *testing.T) {
	_, b, p := twoStores(t)
	c := NewStore(b.Dir)
	p.Name = "blind"
	require.NoError(t, c.SaveProjectLocked(p))
}

func TestReplaceProjectBypassesRecorderAndBaseline(t *testing.T) {
	a, b, p := twoStores(t)
	rec := &fakeRecorder{}
	a.SetRecorder(rec)
	mine, err := a.LoadProjectByID(p.ID)
	require.NoError(t, err)

	incoming := mine
	incoming.Name = "from server"
	written, err := a.ReplaceProject(p.ID, func(current *domain.Project) (*domain.Project, error) {
		require.NotNil(t, current)
		return &incoming, nil
	})
	require.NoError(t, err)
	assert.True(t, written)
	assert.Empty(t, rec.saves, "sync writes must not be journaled")

	got, err := b.LoadProjectByID(p.ID)
	require.NoError(t, err)
	assert.Equal(t, "from server", got.Name)

	skipped, err := a.ReplaceProject(p.ID, func(*domain.Project) (*domain.Project, error) { return nil, nil })
	require.NoError(t, err)
	assert.False(t, skipped)

	fresh, err := domain.NewProject("New from server")
	require.NoError(t, err)
	written, err = a.ReplaceProject(fresh.ID, func(current *domain.Project) (*domain.Project, error) {
		require.Nil(t, current)
		return &fresh, nil
	})
	require.NoError(t, err)
	assert.True(t, written, "a project this device lacks is created")
}

func TestRemoveProjectRecordsNothing(t *testing.T) {
	a, _, p := twoStores(t)
	rec := &fakeRecorder{}
	a.SetRecorder(rec)
	removed, err := a.RemoveProject(p.ID)
	require.NoError(t, err)
	assert.True(t, removed)
	assert.Empty(t, rec.deletes)
	removed, err = a.RemoveProject(p.ID)
	require.NoError(t, err)
	assert.False(t, removed, "a second removal is a no-op")
	_, err = a.LoadProjectByID(p.ID)
	require.ErrorIs(t, err, ErrProjectNotFound)
}
