package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"phasionary/internal/domain"
	"phasionary/internal/journal"
	"phasionary/internal/syncproto"
)

func init() { enrollFailureDelay = 0 }

type testEnv struct {
	t   *testing.T
	db  *DB
	srv *httptest.Server
}

func newEnv(t *testing.T) *testEnv {
	t.Helper()
	db, err := OpenDB(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	srv := httptest.NewServer(New(db, "").Handler())
	t.Cleanup(srv.Close)
	return &testEnv{t: t, db: db, srv: srv}
}

func (e *testEnv) post(path, token string, body any) (int, []byte) {
	e.t.Helper()
	data, err := json.Marshal(body)
	require.NoError(e.t, err)
	req, err := http.NewRequest(http.MethodPost, e.srv.URL+path, bytes.NewReader(data))
	require.NoError(e.t, err)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(e.t, err)
	defer resp.Body.Close()
	out, err := io.ReadAll(resp.Body)
	require.NoError(e.t, err)
	return resp.StatusCode, out
}

type testDevice struct {
	id, token string
	seq       uint64
	cursor    uint64
}

func (e *testEnv) enroll(name string) *testDevice {
	e.t.Helper()
	code, err := e.db.NewEnrollCode(context.Background())
	require.NoError(e.t, err)
	id, err := domain.NewID()
	require.NoError(e.t, err)
	status, body := e.post(syncproto.EnrollPath, "", syncproto.EnrollRequest{Code: code, DeviceID: id, Name: name})
	require.Equal(e.t, http.StatusOK, status, string(body))
	var resp syncproto.EnrollResponse
	require.NoError(e.t, json.Unmarshal(body, &resp))
	return &testDevice{id: id, token: resp.Token}
}

func (d *testDevice) ops(ts string, drafts []journal.Draft) []journal.Op {
	var out []journal.Op
	for _, dr := range drafts {
		d.seq++
		id, _ := domain.NewID()
		out = append(out, journal.Op{OpID: id, DeviceID: d.id, Seq: d.seq, Timestamp: ts, Draft: dr})
	}
	return out
}

func (e *testEnv) sync(d *testDevice, ops []journal.Op) syncproto.SyncResponse {
	e.t.Helper()
	status, body := e.post(syncproto.SyncPath, d.token,
		syncproto.SyncRequest{DeviceID: d.id, Cursor: d.cursor, Ops: ops})
	require.Equal(e.t, http.StatusOK, status, string(body))
	var resp syncproto.SyncResponse
	require.NoError(e.t, json.Unmarshal(body, &resp))
	d.cursor = resp.ServerCursor
	return resp
}

func findSnapshot(t *testing.T, resp syncproto.SyncResponse, id string) syncproto.ProjectSnapshot {
	t.Helper()
	for _, s := range resp.Projects {
		if s.ID == id {
			return s
		}
	}
	t.Fatalf("project %s not in response", id)
	return syncproto.ProjectSnapshot{}
}

func baseProject(t *testing.T) domain.Project {
	t.Helper()
	p, err := domain.NewProject("Base")
	require.NoError(t, err)
	p.CreatedAt = "2030-01-01T00:00:00Z"
	for _, name := range []string{"Inbox", "Later"} {
		c, err := domain.NewCategory(name)
		require.NoError(t, err)
		c.CreatedAt = p.CreatedAt
		p.Categories = append(p.Categories, c)
	}
	for _, title := range []string{"one", "two"} {
		task, err := domain.NewTask(title)
		require.NoError(t, err)
		task.CreatedAt, task.UpdatedAt = p.CreatedAt, p.CreatedAt
		p.Categories[0].Tasks = append(p.Categories[0].Tasks, task)
	}
	return p
}

func clone(t *testing.T, p domain.Project) domain.Project {
	t.Helper()
	data, err := json.Marshal(p)
	require.NoError(t, err)
	var out domain.Project
	require.NoError(t, json.Unmarshal(data, &out))
	return out
}

func assertSameProject(t *testing.T, want domain.Project, got *domain.Project) {
	t.Helper()
	require.NotNil(t, got)
	assert.Equal(t, want.Name, got.Name)
	assert.Empty(t, journal.DiffProjects(&want, *got), "snapshot differs from expected project")
}

const (
	t1 = "2030-06-01T10:00:00Z"
	t2 = "2030-06-01T10:00:01Z"
	t3 = "2030-06-01T10:00:02Z"
)

func TestEnrollCodeIsSingleUseAndExpires(t *testing.T) {
	e := newEnv(t)
	code, err := e.db.NewEnrollCode(context.Background())
	require.NoError(t, err)
	id, _ := domain.NewID()

	loose := []byte(code)
	loose = bytes.ToLower(bytes.ReplaceAll(loose, []byte("-"), nil))
	status, _ := e.post(syncproto.EnrollPath, "", syncproto.EnrollRequest{Code: string(loose), DeviceID: id})
	assert.Equal(t, http.StatusOK, status)

	id2, _ := domain.NewID()
	status, _ = e.post(syncproto.EnrollPath, "", syncproto.EnrollRequest{Code: code, DeviceID: id2})
	assert.Equal(t, http.StatusUnauthorized, status, "a code admits one device")

	status, _ = e.post(syncproto.EnrollPath, "", syncproto.EnrollRequest{Code: "ZZZZ-ZZZZ", DeviceID: id2})
	assert.Equal(t, http.StatusUnauthorized, status)

	_, err = e.db.sql.Exec(`INSERT INTO enroll_codes (code, expires_at) VALUES ('AAAA-BBBB', '2000-01-01T00:00:00Z')`)
	require.NoError(t, err)
	status, _ = e.post(syncproto.EnrollPath, "", syncproto.EnrollRequest{Code: "AAAA-BBBB", DeviceID: id2})
	assert.Equal(t, http.StatusUnauthorized, status, "expired codes are refused")

	status, _ = e.post(syncproto.EnrollPath, "", syncproto.EnrollRequest{Code: code, DeviceID: "../evil"})
	assert.Equal(t, http.StatusBadRequest, status)
}

func TestSyncRequiresMatchingDevice(t *testing.T) {
	e := newEnv(t)
	a := e.enroll("a")
	b := e.enroll("b")

	status, _ := e.post(syncproto.SyncPath, "", syncproto.SyncRequest{DeviceID: a.id})
	assert.Equal(t, http.StatusUnauthorized, status)
	status, _ = e.post(syncproto.SyncPath, "nope", syncproto.SyncRequest{DeviceID: a.id})
	assert.Equal(t, http.StatusUnauthorized, status)
	status, _ = e.post(syncproto.SyncPath, a.token, syncproto.SyncRequest{DeviceID: b.id})
	assert.Equal(t, http.StatusForbidden, status)
}

func TestPushIsIdempotentOnRetry(t *testing.T) {
	e := newEnv(t)
	a := e.enroll("a")
	p := baseProject(t)
	ops := a.ops(t1, journal.DiffProjects(nil, p))

	first := e.sync(a, ops)
	assert.Equal(t, uint64(len(ops)), first.AckedThroughSeq)
	assertSameProject(t, p, findSnapshot(t, first, p.ID).Project)

	a.cursor = 0
	again := e.sync(a, ops)
	assert.Equal(t, uint64(len(ops)), again.AckedThroughSeq)
	assertSameProject(t, p, findSnapshot(t, again, p.ID).Project)
	assert.Equal(t, first.ServerCursor, again.ServerCursor, "a replayed push must not advance the cursor")
}

func TestTwoDevicesConverge(t *testing.T) {
	e := newEnv(t)
	a, b := e.enroll("a"), e.enroll("b")
	p := baseProject(t)
	e.sync(a, a.ops(t1, journal.DiffProjects(nil, p)))

	pulled := e.sync(b, nil)
	onB := findSnapshot(t, pulled, p.ID).Project
	assertSameProject(t, p, onB)

	edited := clone(t, *onB)
	require.NoError(t, edited.Categories[0].Tasks[0].SetStatus(domain.StatusCompleted))
	edited.Categories[0].Tasks[0].UpdatedAt = t2
	e.sync(b, b.ops(t2, journal.DiffProjects(onB, edited)))

	onA := findSnapshot(t, e.sync(a, nil), p.ID).Project
	assertSameProject(t, edited, onA)

	assert.Empty(t, e.sync(a, nil).Projects, "nothing changed since A's cursor")
	assert.Empty(t, e.sync(b, nil).Projects)
}

func TestSameFieldConflictIsLastWriterWins(t *testing.T) {
	e := newEnv(t)
	a, b := e.enroll("a"), e.enroll("b")
	p := baseProject(t)
	e.sync(a, a.ops(t1, journal.DiffProjects(nil, p)))
	e.sync(b, nil)

	fromA := clone(t, p)
	fromA.Categories[0].Tasks[0].Title = "A's title"
	fromB := clone(t, p)
	fromB.Categories[0].Tasks[0].Title = "B's title"

	// B's edit is later but reaches the server first.
	e.sync(b, b.ops(t3, journal.DiffProjects(&p, fromB)))
	afterA := e.sync(a, a.ops(t2, journal.DiffProjects(&p, fromA)))
	assert.Equal(t, "B's title", findSnapshot(t, afterA, p.ID).Project.Categories[0].Tasks[0].Title)
	afterB := e.sync(b, nil)
	assert.Equal(t, "B's title", findSnapshot(t, afterB, p.ID).Project.Categories[0].Tasks[0].Title)
}

func TestTombstoneBeatsLaterUpdate(t *testing.T) {
	e := newEnv(t)
	a, b := e.enroll("a"), e.enroll("b")
	p := baseProject(t)
	e.sync(a, a.ops(t1, journal.DiffProjects(nil, p)))
	e.sync(b, nil)
	victim := p.Categories[0].Tasks[0].ID

	deleted := clone(t, p)
	deleted.Categories[0].Tasks = deleted.Categories[0].Tasks[1:]
	e.sync(a, a.ops(t2, journal.DiffProjects(&p, deleted)))

	edited := clone(t, p)
	edited.Categories[0].Tasks[0].Title = "edited after deletion"
	resp := e.sync(b, b.ops(t3, journal.DiffProjects(&p, edited)))

	got := findSnapshot(t, resp, p.ID).Project
	for _, task := range got.Categories[0].Tasks {
		assert.NotEqual(t, victim, task.ID, "a deleted task must not come back")
	}
	assert.Len(t, got.Categories[0].Tasks, 1)
}

func TestConcurrentAddsBothSurvive(t *testing.T) {
	e := newEnv(t)
	a, b := e.enroll("a"), e.enroll("b")
	p := baseProject(t)
	e.sync(a, a.ops(t1, journal.DiffProjects(nil, p)))
	e.sync(b, nil)

	fromA := clone(t, p)
	ta, _ := domain.NewTask("from A")
	ta.CreatedAt, ta.UpdatedAt = t2, t2
	fromA.Categories[0].Tasks = append(fromA.Categories[0].Tasks, ta)
	fromB := clone(t, p)
	tb, _ := domain.NewTask("from B")
	tb.CreatedAt, tb.UpdatedAt = t3, t3
	fromB.Categories[0].Tasks = append(fromB.Categories[0].Tasks, tb)

	e.sync(a, a.ops(t2, journal.DiffProjects(&p, fromA)))
	resp := e.sync(b, b.ops(t3, journal.DiffProjects(&p, fromB)))

	tasks := findSnapshot(t, resp, p.ID).Project.Categories[0].Tasks
	require.Len(t, tasks, 4)
	assert.Equal(t, tb.ID, tasks[2].ID, "B's later order wins")
	assert.Equal(t, ta.ID, tasks[3].ID, "A's task is appended, not lost")
}

func TestMoveRenameAndReorderRoundTrip(t *testing.T) {
	e := newEnv(t)
	a := e.enroll("a")
	p := baseProject(t)
	e.sync(a, a.ops(t1, journal.DiffProjects(nil, p)))

	moved := clone(t, p)
	task := moved.Categories[0].Tasks[0]
	moved.Categories[0].Tasks = moved.Categories[0].Tasks[1:]
	moved.Categories[1].Tasks = []domain.Task{task}
	resp := e.sync(a, a.ops(t2, journal.DiffProjects(&p, moved)))
	assertSameProject(t, moved, findSnapshot(t, resp, p.ID).Project)

	changed := clone(t, moved)
	changed.Categories[0].Name = "Now"
	changed.Categories[0], changed.Categories[1] = changed.Categories[1], changed.Categories[0]
	changed.Name = "Renamed"
	resp = e.sync(a, a.ops(t3, journal.DiffProjects(&moved, changed)))
	assertSameProject(t, changed, findSnapshot(t, resp, p.ID).Project)
}

func TestProjectDeletePullsAsDeleted(t *testing.T) {
	e := newEnv(t)
	a, b := e.enroll("a"), e.enroll("b")
	p := baseProject(t)
	e.sync(a, a.ops(t1, journal.DiffProjects(nil, p)))
	e.sync(b, nil)

	e.sync(a, a.ops(t2, []journal.Draft{{Kind: journal.KindProjectDelete, ProjectID: p.ID}}))
	snap := findSnapshot(t, e.sync(b, nil), p.ID)
	assert.True(t, snap.Deleted)
	assert.Nil(t, snap.Project)

	resp := e.sync(b, b.ops(t3, journal.DiffProjects(nil, p)))
	assert.True(t, findSnapshot(t, resp, p.ID).Deleted)
}

func TestInvalidOpsRejected(t *testing.T) {
	e := newEnv(t)
	a := e.enroll("a")
	bad := a.ops(t1, []journal.Draft{{Kind: journal.KindProjectCreate, ProjectID: "../escape"}})
	status, _ := e.post(syncproto.SyncPath, a.token, syncproto.SyncRequest{DeviceID: a.id, Ops: bad})
	assert.Equal(t, http.StatusBadRequest, status)

	unknown := a.ops(t1, []journal.Draft{{Kind: "task.explode", ProjectID: "p1", EntityID: "t1"}})
	status, _ = e.post(syncproto.SyncPath, a.token, syncproto.SyncRequest{DeviceID: a.id, Ops: unknown})
	assert.Equal(t, http.StatusBadRequest, status)

	other := e.enroll("b")
	foreign := other.ops(t1, []journal.Draft{{Kind: journal.KindProjectCreate, ProjectID: "p1"}})
	status, _ = e.post(syncproto.SyncPath, a.token, syncproto.SyncRequest{DeviceID: a.id, Ops: foreign})
	assert.Equal(t, http.StatusBadRequest, status, "ops journaled by another device")
}

func TestDatabaseSurvivesReopen(t *testing.T) {
	dir := t.TempDir()
	db, err := OpenDB(dir)
	require.NoError(t, err)
	code, err := db.NewEnrollCode(context.Background())
	require.NoError(t, err)
	require.NoError(t, db.Close())

	db, err = OpenDB(dir)
	require.NoError(t, err)
	defer db.Close()
	tx, err := db.begin(context.Background())
	require.NoError(t, err)
	defer tx.Rollback()
	ok, err := consumeCode(tx, code)
	require.NoError(t, err)
	assert.True(t, ok)
}
