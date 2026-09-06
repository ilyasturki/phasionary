package journal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppendAssignsMonotonicSeq(t *testing.T) {
	j := Open(t.TempDir())

	require.NoError(t, j.Append("dev-1", []Draft{
		{Kind: KindTaskCreate, ProjectID: "p1", EntityID: "t1"},
		{Kind: KindTaskUpdate, ProjectID: "p1", EntityID: "t1"},
	}))
	require.NoError(t, j.Append("dev-1", []Draft{
		{Kind: KindTaskDelete, ProjectID: "p1", EntityID: "t1"},
	}))

	ops, err := j.Entries()
	require.NoError(t, err)
	require.Len(t, ops, 3)
	for i, op := range ops {
		assert.Equal(t, uint64(i+1), op.Seq)
		assert.Equal(t, "dev-1", op.DeviceID)
		assert.NotEmpty(t, op.OpID)
		assert.NotEmpty(t, op.Timestamp)
	}
	assert.Equal(t, KindTaskDelete, ops[2].Kind)
}

func TestPruneKeepsUnackedAndSeqMonotonic(t *testing.T) {
	j := Open(t.TempDir())

	require.NoError(t, j.Append("dev-1", []Draft{
		{Kind: KindTaskCreate, ProjectID: "p1", EntityID: "t1"},
		{Kind: KindTaskCreate, ProjectID: "p1", EntityID: "t2"},
		{Kind: KindTaskCreate, ProjectID: "p1", EntityID: "t3"},
	}))
	require.NoError(t, j.PruneThrough(1))

	ops, err := j.Entries()
	require.NoError(t, err)
	require.Len(t, ops, 2)
	assert.Equal(t, uint64(2), ops[0].Seq)
	assert.Equal(t, uint64(3), ops[1].Seq)

	require.NoError(t, j.PruneThrough(3))
	ops, err = j.Entries()
	require.NoError(t, err)
	require.Empty(t, ops, "acked entries must be gone")

	require.NoError(t, j.Append("dev-1", []Draft{
		{Kind: KindTaskCreate, ProjectID: "p1", EntityID: "t4"},
	}))
	ops, err = j.Entries()
	require.NoError(t, err)
	require.Len(t, ops, 1)
	assert.Equal(t, uint64(4), ops[0].Seq, "seq never reuses a pruned value")
}

func TestTornTailHealed(t *testing.T) {
	dir := t.TempDir()
	j := Open(dir)

	require.NoError(t, j.Append("dev-1", []Draft{
		{Kind: KindTaskCreate, ProjectID: "p1", EntityID: "t1"},
	}))
	f, err := os.OpenFile(filepath.Join(dir, "journal.jsonl"), os.O_WRONLY|os.O_APPEND, 0o600)
	require.NoError(t, err)
	_, err = f.WriteString(`{"op_id":"torn`)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	require.NoError(t, j.Append("dev-1", []Draft{
		{Kind: KindTaskUpdate, ProjectID: "p1", EntityID: "t1"},
	}))

	ops, err := j.Entries()
	require.NoError(t, err)
	require.Len(t, ops, 2, "torn fragment is skipped, both real ops survive")
	assert.Equal(t, uint64(1), ops[0].Seq)
	assert.Equal(t, uint64(2), ops[1].Seq)
}

func TestEmptyJournal(t *testing.T) {
	j := Open(t.TempDir())
	ops, err := j.Entries()
	require.NoError(t, err)
	assert.Empty(t, ops)
	assert.NoError(t, j.Append("dev-1", nil), "empty batch is a no-op")
}
