package syncclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"phasionary/internal/data"
	"phasionary/internal/domain"
	"phasionary/internal/journal"
	"phasionary/internal/syncproto"
)

var (
	ErrNotEnrolled     = errors.New("this device is not enrolled with a sync server")
	ErrAlreadyEnrolled = errors.New("this device is already enrolled; run `phasionary sync logout` first")

	httpClient = &http.Client{Timeout: 2 * time.Minute}
)

type HTTPError struct {
	Status  int
	Message string
}

func (e *HTTPError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("server answered %d", e.Status)
	}
	return fmt.Sprintf("server answered %d: %s", e.Status, e.Message)
}

func post(ctx context.Context, baseURL, token, path string, body, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode/100 != 2 {
		var msg syncproto.ErrorResponse
		_ = json.Unmarshal(raw, &msg)
		return &HTTPError{Status: resp.StatusCode, Message: msg.Error}
	}
	return json.Unmarshal(raw, out)
}

type Summary struct {
	Pushed    int
	Written   int
	Unchanged int
	// Skipped: changed locally mid-round; the next round reconciles them.
	Skipped int
	Removed int
}

func Login(ctx context.Context, stateDir string, store *data.Store, serverURL, code, name string) (journal.Device, Summary, error) {
	if _, ok, err := journal.LoadDevice(stateDir); err != nil {
		return journal.Device{}, Summary{}, err
	} else if ok {
		return journal.Device{}, Summary{}, ErrAlreadyEnrolled
	}
	serverURL = strings.TrimRight(serverURL, "/")
	id, err := domain.NewID()
	if err != nil {
		return journal.Device{}, Summary{}, err
	}
	var resp syncproto.EnrollResponse
	req := syncproto.EnrollRequest{Code: code, DeviceID: id, Name: name}
	if err := post(ctx, serverURL, "", syncproto.EnrollPath, req, &resp); err != nil {
		return journal.Device{}, Summary{}, fmt.Errorf("enrolling: %w", err)
	}
	dev := journal.Device{DeviceID: id, ServerURL: serverURL, Token: resp.Token}
	if err := journal.SaveDevice(stateDir, dev); err != nil {
		return journal.Device{}, Summary{}, err
	}

	projects, err := store.ListProjects()
	if err != nil {
		return dev, Summary{}, err
	}
	j := journal.Open(stateDir)
	for _, p := range projects {
		// ReplaceProject only for its flock: a racing save lands before or
		// after the cascade, never inside it.
		_, err := store.ReplaceProject(p.ID, func(current *domain.Project) (*domain.Project, error) {
			if current == nil {
				return nil, nil
			}
			return nil, j.Append(dev.DeviceID, journal.DiffProjects(nil, *current))
		})
		if err != nil {
			return dev, Summary{}, err
		}
	}
	sum, err := Run(ctx, stateDir, store)
	return dev, sum, err
}

// Test seam: lands a concurrent local edit between the push and the apply.
var afterPush = func() {}

func Run(ctx context.Context, stateDir string, store *data.Store) (Summary, error) {
	dev, ok, err := journal.LoadDevice(stateDir)
	if err != nil {
		return Summary{}, err
	}
	if !ok {
		return Summary{}, ErrNotEnrolled
	}
	j := journal.Open(stateDir)
	ops, err := j.Entries()
	if err != nil {
		return Summary{}, err
	}
	var resp syncproto.SyncResponse
	req := syncproto.SyncRequest{DeviceID: dev.DeviceID, Cursor: dev.Cursor, Ops: ops}
	if err := post(ctx, dev.ServerURL, dev.Token, syncproto.SyncPath, req, &resp); err != nil {
		return Summary{}, err
	}
	sum := Summary{Pushed: len(ops)}
	if err := j.PruneThrough(resp.AckedThroughSeq); err != nil {
		return sum, err
	}
	afterPush()

	for _, snap := range resp.Projects {
		if snap.Deleted {
			removed, err := store.RemoveProject(snap.ID)
			if err != nil {
				return sum, fmt.Errorf("removing project %s: %w", snap.ID, err)
			}
			if removed {
				sum.Removed++
			}
			continue
		}
		if snap.Project == nil {
			continue
		}
		written, err := store.ReplaceProject(snap.ID, func(current *domain.Project) (*domain.Project, error) {
			pending, err := j.Entries()
			if err != nil {
				return nil, err
			}
			// Still journaled ⇒ written after the push; the snapshot predates
			// it, so leave it for the next round.
			if slices.ContainsFunc(pending, func(op journal.Op) bool { return op.ProjectID == snap.ID }) {
				sum.Skipped++
				return nil, nil
			}
			// Rewriting an identical snapshot would make an open TUI's next
			// save stale for nothing.
			if current != nil && len(journal.DiffProjects(current, *snap.Project)) == 0 {
				sum.Unchanged++
				return nil, nil
			}
			return snap.Project, nil
		})
		if err != nil {
			return sum, fmt.Errorf("writing project %s: %w", snap.ID, err)
		}
		if written {
			sum.Written++
		}
	}

	dev.Cursor = resp.ServerCursor
	if err := journal.SaveDevice(stateDir, dev); err != nil {
		return sum, err
	}
	return sum, nil
}
