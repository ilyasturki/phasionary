package config

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
)

const (
	EnvDataPath   = "PHASIONARY_DATA_PATH"
	EnvConfigPath = "PHASIONARY_CONFIG_PATH"

	StatusDisplayText  = "text"
	StatusDisplayIcons = "icons"

	PriorityColorFull = "full"
	PriorityColorIcon = "icon"
	PriorityColorNone = "none"
)

// Config holds user preferences.
type Config struct {
	StatusDisplay               string `json:"status_display,omitempty"`
	PriorityColor               string `json:"priority_color,omitempty"`
	ShowShortcutBar             bool   `json:"show_shortcut_bar"`
	ExpandDescriptionsByDefault bool   `json:"expand_descriptions_by_default"`

	// ServeToken is the bearer token `phasionary serve` requires. It is
	// generated on first serve rather than defaulted, so an empty value here
	// means "not yet generated" and never "authentication disabled".
	//
	// This is the only secret the file holds, and it is why the whole file is
	// written 0600: a full-access credential in a world-readable file would be
	// readable by every other account on the machine.
	ServeToken string `json:"serve_token,omitempty"`
}

// NewServeToken returns a fresh 256-bit bearer token, base64url-encoded without
// padding so it survives a URL, a shell argument and a phone's text field
// unescaped.
func NewServeToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() Config {
	return Config{
		StatusDisplay:               StatusDisplayIcons,
		PriorityColor:               PriorityColorFull,
		ShowShortcutBar:             true,
		ExpandDescriptionsByDefault: false,
	}
}

func ResolveDataDir(input string) (string, error) {
	if input != "" {
		return filepath.Join(input, "projects"), nil
	}
	if env := os.Getenv(EnvDataPath); env != "" {
		return filepath.Join(env, "projects"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "phasionary", "projects"), nil
}

// configDirFromPath accepts either a directory path or a path pointing at a
// `*.json` file (in which case the parent directory is used).
func configDirFromPath(p string) string {
	if strings.HasSuffix(p, ".json") {
		return filepath.Dir(p)
	}
	return p
}

// ResolveConfigDir returns the config directory path.
// Priority: input > PHASIONARY_CONFIG_PATH > XDG_CONFIG_HOME > ~/.config/phasionary
func ResolveConfigDir(input string) (string, error) {
	if input != "" {
		return configDirFromPath(input), nil
	}
	if env := os.Getenv(EnvConfigPath); env != "" {
		return configDirFromPath(env), nil
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "phasionary"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "phasionary"), nil
}

// ResolveConfigPath returns the full path to config.json.
func ResolveConfigPath(input string) (string, error) {
	dir, err := ResolveConfigDir(input)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}
