package journal

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"phasionary/internal/domain"
	"phasionary/internal/fsutil"
)

const (
	journalFileName = "journal.jsonl"
	lockFileName    = "journal.lock"
	headFileName    = "journal.head"
)

// Outbox, not history: pruned to empty once the server acks.
type Journal struct {
	dir string
}

func Open(stateDir string) *Journal {
	return &Journal{dir: stateDir}
}

func (j *Journal) path() string     { return filepath.Join(j.dir, journalFileName) }
func (j *Journal) lockPath() string { return filepath.Join(j.dir, lockFileName) }
func (j *Journal) headPath() string { return filepath.Join(j.dir, headFileName) }

func (j *Journal) lock() (*os.File, error) {
	if err := os.MkdirAll(j.dir, stateDirMode); err != nil {
		return nil, err
	}
	return fsutil.Lock(j.lockPath(), 0o600)
}

// Unparseable lines are skipped, not fatal: they only arise from a crash-torn
// append, and failing here would block every future save.
func (j *Journal) readEntries() ([]Op, error) {
	data, err := os.ReadFile(j.path())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var ops []Op
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		var op Op
		if err := json.Unmarshal(line, &op); err != nil {
			continue
		}
		ops = append(ops, op)
	}
	return ops, sc.Err()
}

// Max of the head file and surviving entries: seq must never regress across a
// prune, or the server drops fresh ops as already seen.
func (j *Journal) highWater(entries []Op) uint64 {
	var hw uint64
	if data, err := os.ReadFile(j.headPath()); err == nil {
		if v, err := strconv.ParseUint(string(bytes.TrimSpace(data)), 10, 64); err == nil {
			hw = v
		}
	}
	for _, op := range entries {
		hw = max(hw, op.Seq)
	}
	return hw
}

func (j *Journal) Append(deviceID string, drafts []Draft) error {
	if len(drafts) == 0 {
		return nil
	}
	l, err := j.lock()
	if err != nil {
		return err
	}
	defer l.Close()

	entries, err := j.readEntries()
	if err != nil {
		return err
	}
	seq := j.highWater(entries)

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	now := domain.NowTimestamp()
	for _, d := range drafts {
		id, err := domain.NewID()
		if err != nil {
			return err
		}
		seq++
		op := Op{OpID: id, DeviceID: deviceID, Seq: seq, Timestamp: now, Draft: d}
		if err := enc.Encode(op); err != nil {
			return err
		}
	}

	// O_RDWR, not O_WRONLY: the torn-tail probe below reads the last byte.
	f, err := os.OpenFile(j.path(), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	// Heal a torn tail before appending: without the newline the previous
	// crash's fragment would glue onto our first op and corrupt both lines.
	if info, err := f.Stat(); err == nil && info.Size() > 0 {
		tail := make([]byte, 1)
		if _, err := f.ReadAt(tail, info.Size()-1); err == nil && tail[0] != '\n' {
			if _, err := f.Write([]byte{'\n'}); err != nil {
				_ = f.Close()
				return err
			}
		}
	}
	if _, err := f.Write(buf.Bytes()); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

func (j *Journal) Entries() ([]Op, error) {
	l, err := j.lock()
	if err != nil {
		return nil, err
	}
	defer l.Close()
	return j.readEntries()
}

func (j *Journal) PruneThrough(seq uint64) error {
	l, err := j.lock()
	if err != nil {
		return err
	}
	defer l.Close()

	entries, err := j.readEntries()
	if err != nil {
		return err
	}
	// Head first: a crash after it leaves only a seq gap, while the reverse
	// order could reuse seqs the server has already seen.
	seq = min(seq, j.highWater(entries))
	if err := fsutil.WriteAtomic(j.headPath(), strconv.AppendUint(nil, seq, 10), 0o600); err != nil {
		return err
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for _, op := range entries {
		if op.Seq <= seq {
			continue
		}
		if err := enc.Encode(op); err != nil {
			return err
		}
	}
	return fsutil.WriteAtomic(j.path(), buf.Bytes(), 0o600)
}
