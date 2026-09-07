package journal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"phasionary/internal/domain"
	"phasionary/internal/fsutil"
)

const deviceFileName = "device.json"

// 0700: the state dir holds the sync credential (a bearer token once enrolled).
const stateDirMode = 0o700

type Device struct {
	DeviceID  string `json:"device_id"`
	ServerURL string `json:"server_url,omitempty"`
	Token     string `json:"token,omitempty"`
	// Advanced only after a pull is fully applied, so an interrupted one is
	// redone.
	Cursor uint64 `json:"cursor,omitempty"`
}

func devicePath(stateDir string) string {
	return filepath.Join(stateDir, deviceFileName)
}

func LoadDevice(stateDir string) (Device, bool, error) {
	data, err := os.ReadFile(devicePath(stateDir))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Device{}, false, nil
		}
		return Device{}, false, err
	}
	var d Device
	if err := json.Unmarshal(data, &d); err != nil {
		return Device{}, false, fmt.Errorf("parsing %s: %w", devicePath(stateDir), err)
	}
	if d.DeviceID == "" {
		return Device{}, false, fmt.Errorf("%s: missing device_id", devicePath(stateDir))
	}
	return d, true, nil
}

func SaveDevice(stateDir string, d Device) error {
	if err := os.MkdirAll(stateDir, stateDirMode); err != nil {
		return err
	}
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	return fsutil.WriteAtomic(devicePath(stateDir), data, 0o600)
}

// Discards pending journal entries too; callers check for those first.
func Unenroll(stateDir string) error {
	for _, name := range []string{deviceFileName, journalFileName, headFileName, lockFileName} {
		if err := os.Remove(filepath.Join(stateDir, name)); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}
	return nil
}

// Refuses to overwrite: an identity swap orphans the server's cursor for the
// old ID.
func MintDevice(stateDir, serverURL, token string) (Device, error) {
	if _, ok, err := LoadDevice(stateDir); err != nil {
		return Device{}, err
	} else if ok {
		return Device{}, errors.New("device identity already exists")
	}
	id, err := domain.NewID()
	if err != nil {
		return Device{}, err
	}
	d := Device{DeviceID: id, ServerURL: serverURL, Token: token}
	if err := SaveDevice(stateDir, d); err != nil {
		return Device{}, err
	}
	return d, nil
}
