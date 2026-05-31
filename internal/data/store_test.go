package data

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"phasionary/internal/domain"
)

func TestDeleteProject(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)
	require.NoError(t, store.Ensure())

	project, err := store.CreateProject("Test Project")
	require.NoError(t, err)

	projectPath := filepath.Join(tmpDir, project.ID+".json")
	_, err = os.Stat(projectPath)
	require.NoError(t, err, "project file should exist")

	err = store.DeleteProject(project.ID)
	require.NoError(t, err)

	_, err = os.Stat(projectPath)
	assert.True(t, os.IsNotExist(err), "project file should be deleted")

	projects, err := store.ListProjects()
	require.NoError(t, err)
	assert.Empty(t, projects)
}

func TestDeleteProject_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)
	require.NoError(t, store.Ensure())

	err := store.DeleteProject("nonexistent-id")
	assert.ErrorIs(t, err, ErrProjectNotFound)
}

func TestRenameProject_RejectsDuplicateName(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)
	require.NoError(t, store.Ensure())

	a, err := store.CreateProject("Alpha")
	require.NoError(t, err)
	_, err = store.CreateProject("Beta")
	require.NoError(t, err)

	_, err = store.RenameProject(a.ID, "Beta")
	assert.ErrorIs(t, err, ErrDuplicateProjectName)

	got, err := store.LoadProjectByID(a.ID)
	require.NoError(t, err)
	assert.Equal(t, "Alpha", got.Name, "rename must not partially apply when rejected")
}

func TestRenameProject_AllowsSameName(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)
	require.NoError(t, store.Ensure())

	p, err := store.CreateProject("Alpha")
	require.NoError(t, err)

	updated, err := store.RenameProject(p.ID, "Alpha")
	require.NoError(t, err, "renaming a project to its current name must not trip the duplicate check")
	assert.Equal(t, "Alpha", updated.Name)
}

func TestRenameProject_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)
	require.NoError(t, store.Ensure())

	_, err := store.RenameProject("nonexistent-id", "Anything")
	assert.ErrorIs(t, err, ErrProjectNotFound)
}

func TestSaveProjectLocked_RejectsMissing(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)
	require.NoError(t, store.Ensure())

	p, err := store.CreateProject("Ghost")
	require.NoError(t, err)
	require.NoError(t, store.DeleteProject(p.ID))

	// Simulates a TUI/CLI process holding a stale in-memory snapshot and
	// trying to save after another process deleted the project. The save
	// must refuse rather than resurrect the deleted file.
	err = store.SaveProjectLocked(p)
	assert.ErrorIs(t, err, ErrProjectNotFound)

	_, statErr := os.Stat(filepath.Join(tmpDir, p.ID+".json"))
	assert.True(t, os.IsNotExist(statErr), "deleted project must not be resurrected by SaveProjectLocked")
}

func TestImportProject_RejectsDuplicateName(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)
	require.NoError(t, store.Ensure())

	original, err := store.CreateProject("Roadmap")
	require.NoError(t, err)

	// Re-importing a project that still exists must not silently overwrite it;
	// the duplicate-name check rejects it (a different ID doesn't help).
	_, err = store.ImportProject(domain.Project{ID: "some-other-id", Name: "Roadmap"})
	assert.ErrorIs(t, err, ErrDuplicateProjectName)

	reloaded, err := store.LoadProjectByID(original.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, reloaded.Categories, "the original project must be left intact")
}

func TestImportProject_CollidingIDGetsFreshID(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)
	require.NoError(t, store.Ensure())

	original, err := store.CreateProject("Alpha")
	require.NoError(t, err)

	// An import whose ID collides with a live project (but whose name is
	// unique) must be saved under a fresh ID rather than clobbering the
	// existing file.
	imported, err := store.ImportProject(domain.Project{ID: original.ID, Name: "Beta"})
	require.NoError(t, err)
	assert.NotEqual(t, original.ID, imported.ID, "colliding import ID must be replaced")

	reloaded, err := store.LoadProjectByID(original.ID)
	require.NoError(t, err)
	assert.Equal(t, "Alpha", reloaded.Name, "the original project must survive the import")
}

func TestImportProject_FillsBlankIDAndTimestamp(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)
	require.NoError(t, store.Ensure())

	imported, err := store.ImportProject(domain.Project{Name: "Hand-written"})
	require.NoError(t, err)
	assert.NotEmpty(t, imported.ID, "a blank ID must be filled in")
	assert.NotEmpty(t, imported.CreatedAt, "a blank CreatedAt must be filled in")

	_, err = store.LoadProjectByID(imported.ID)
	require.NoError(t, err)
}

func TestLoadProjectByID_RejectsBlankIDFile(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)
	require.NoError(t, store.Ensure())

	// A file with empty "id" would otherwise load as a zero-ID project and
	// later be saved back to "<dir>/.json", splitting state.
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "broken.json"), []byte(`{"id":"","name":"x"}`), 0o644))

	_, err := store.LoadProjectByID("broken")
	assert.ErrorIs(t, err, ErrProjectNotFound)
}
