package syncclient

import (
	"context"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"phasionary/internal/data"
	"phasionary/internal/domain"
	"phasionary/internal/journal"
	"phasionary/internal/server"
)

type testServer struct {
	db  *server.DB
	url string
}

func newTestServer(t *testing.T) *testServer {
	t.Helper()
	db, err := server.OpenDB(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	srv := httptest.NewServer(server.New(db, "").Handler())
	t.Cleanup(srv.Close)
	return &testServer{db: db, url: srv.URL}
}

func (s *testServer) code(t *testing.T) string {
	t.Helper()
	code, err := s.db.NewEnrollCode(context.Background())
	require.NoError(t, err)
	return code
}

type node struct {
	stateDir string
	store    *data.Store
}

func newNode(t *testing.T) *node {
	t.Helper()
	stateDir := t.TempDir()
	store := data.NewStore(filepath.Join(t.TempDir(), "projects"))
	require.NoError(t, store.Ensure())
	store.SetRecorder(journal.NewRecorder(stateDir))
	return &node{stateDir: stateDir, store: store}
}

func (n *node) login(t *testing.T, s *testServer) Summary {
	t.Helper()
	_, sum, err := Login(context.Background(), n.stateDir, n.store, s.url, s.code(t), "node")
	require.NoError(t, err)
	return sum
}

func (n *node) run(t *testing.T) Summary {
	t.Helper()
	sum, err := Run(context.Background(), n.stateDir, n.store)
	require.NoError(t, err)
	return sum
}

func (n *node) pending(t *testing.T) int {
	t.Helper()
	ops, err := journal.Open(n.stateDir).Entries()
	require.NoError(t, err)
	return len(ops)
}

func (n *node) project(t *testing.T, id string) domain.Project {
	t.Helper()
	p, err := n.store.LoadProjectByID(id)
	require.NoError(t, err)
	return p
}

func (n *node) edit(t *testing.T, id string, fn func(*domain.Project)) {
	t.Helper()
	_, err := n.store.WithProjectLocked(id, func(p *domain.Project) error {
		fn(p)
		return nil
	})
	require.NoError(t, err)
}

func firstTask(p *domain.Project) *domain.Task {
	return &p.Categories[0].Tasks[0]
}

func first(p domain.Project) domain.Task {
	return p.Categories[0].Tasks[0]
}

func seed(t *testing.T, n *node) domain.Project {
	t.Helper()
	p, err := n.store.CreateProject("Plan")
	require.NoError(t, err)
	return p
}

func assertSame(t *testing.T, want, got domain.Project) {
	t.Helper()
	assert.Equal(t, want.Name, got.Name)
	assert.Empty(t, journal.DiffProjects(&want, got))
}

func TestLoginUploadsAndSecondDeviceReceives(t *testing.T) {
	s := newTestServer(t)
	a, b := newNode(t), newNode(t)
	p := seed(t, a)
	assert.Equal(t, 0, a.pending(t), "nothing journaled before enrollment")

	sum := a.login(t, s)
	assert.Greater(t, sum.Pushed, 0, "the initial upload is the creation cascade")
	assert.Equal(t, 1, sum.Unchanged, "the echo of its own upload is identical")
	assert.Equal(t, 0, a.pending(t), "acked ops are pruned")
	dev, ok, err := journal.LoadDevice(a.stateDir)
	require.NoError(t, err)
	require.True(t, ok)
	assert.NotZero(t, dev.Cursor)
	assert.NotEmpty(t, dev.Token)

	sum = b.login(t, s)
	assert.Equal(t, 1, sum.Written)
	assertSame(t, a.project(t, p.ID), b.project(t, p.ID))
}

func TestEditsFlowBothWays(t *testing.T) {
	s := newTestServer(t)
	a, b := newNode(t), newNode(t)
	p := seed(t, a)
	a.login(t, s)
	b.login(t, s)

	b.edit(t, p.ID, func(p *domain.Project) { firstTask(p).Title = "edited on B" })
	assert.Greater(t, b.pending(t), 0, "the edit was journaled by the attached recorder")
	sum := b.run(t)
	assert.Greater(t, sum.Pushed, 0)
	assert.Equal(t, 0, sum.Written)
	assert.Equal(t, 1, sum.Unchanged)

	sum = a.run(t)
	assert.Equal(t, 0, sum.Pushed)
	assert.Equal(t, 1, sum.Written)
	assert.Equal(t, "edited on B", first(a.project(t, p.ID)).Title)

	a.edit(t, p.ID, func(p *domain.Project) { _ = firstTask(p).SetStatus(domain.StatusCompleted) })
	a.run(t)
	b.run(t)
	assertSame(t, a.project(t, p.ID), b.project(t, p.ID))

	sum = b.run(t)
	assert.Equal(t, Summary{}, sum, "a quiet round does nothing")
}

func TestConcurrentLocalEditIsSkippedThenMerged(t *testing.T) {
	s := newTestServer(t)
	a, b := newNode(t), newNode(t)
	p := seed(t, a)
	a.login(t, s)
	b.login(t, s)

	b.edit(t, p.ID, func(p *domain.Project) { firstTask(p).Title = "from B" })
	b.run(t)

	afterPush = func() {
		a.edit(t, p.ID, func(p *domain.Project) { _ = firstTask(p).SetPriority(domain.PriorityTrivial) })
	}
	t.Cleanup(func() { afterPush = func() {} })
	sum := a.run(t)
	afterPush = func() {}
	assert.Equal(t, 1, sum.Skipped)
	assert.Equal(t, 0, sum.Written)
	local := a.project(t, p.ID)
	assert.NotEqual(t, "from B", first(local).Title, "the stale snapshot must not clobber the local edit")
	assert.Equal(t, domain.PriorityTrivial, first(local).Priority)

	sum = a.run(t)
	assert.Greater(t, sum.Pushed, 0)
	assert.Equal(t, 1, sum.Written)
	merged := first(a.project(t, p.ID))
	assert.Equal(t, "from B", merged.Title)
	assert.Equal(t, domain.PriorityTrivial, merged.Priority)

	b.run(t)
	assertSame(t, a.project(t, p.ID), b.project(t, p.ID))
}

func TestRemoteDeleteRemovesFile(t *testing.T) {
	s := newTestServer(t)
	a, b := newNode(t), newNode(t)
	p := seed(t, a)
	a.login(t, s)
	b.login(t, s)

	require.NoError(t, a.store.DeleteProject(p.ID))
	a.run(t)
	sum := b.run(t)
	assert.Equal(t, 1, sum.Removed)
	_, err := b.store.LoadProjectByID(p.ID)
	require.ErrorIs(t, err, data.ErrProjectNotFound)
	assert.Equal(t, 0, b.pending(t), "the removal is not journaled as a local delete")
}

func TestEnrollmentErrors(t *testing.T) {
	s := newTestServer(t)
	a := newNode(t)
	_, err := Run(context.Background(), a.stateDir, a.store)
	require.ErrorIs(t, err, ErrNotEnrolled)

	_, _, err = Login(context.Background(), a.stateDir, a.store, s.url, "WRONG-CODE", "a")
	require.Error(t, err)
	var httpErr *HTTPError
	require.ErrorAs(t, err, &httpErr)
	_, ok, err := journal.LoadDevice(a.stateDir)
	require.NoError(t, err)
	assert.False(t, ok, "a refused enrollment leaves no identity behind")

	a.login(t, s)
	_, _, err = Login(context.Background(), a.stateDir, a.store, s.url, s.code(t), "a")
	require.ErrorIs(t, err, ErrAlreadyEnrolled)
}
