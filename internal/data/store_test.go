package data

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
