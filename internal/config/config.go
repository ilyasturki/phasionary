package config

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	EnvDataPath   = "PHASIONARY_DATA_PATH"
	EnvConfigPath = "PHASIONARY_CONFIG_PATH"
	EnvStatePath  = "PHASIONARY_STATE_PATH"

	EnvServerDataPath = "PHASIONARY_SERVER_DATA_PATH"

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

	// HelpExpanded remembers whether the help dialog opens on the full
	// shortcut reference rather than the essentials card.
	HelpExpanded bool `json:"help_expanded"`
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() Config {
	return Config{
		StatusDisplay:               StatusDisplayIcons,
		PriorityColor:               PriorityColorFull,
		ShowShortcutBar:             true,
		ExpandDescriptionsByDefault: false,
		HelpExpanded:                false,
	}
}

// input > $env > $xdg/sub > ~/home...; an empty xdg skips that branch.
func resolveDir(input, env, xdg, sub string, home ...string) (string, error) {
	if input != "" {
		return input, nil
	}
	if v := os.Getenv(env); v != "" {
		return v, nil
	}
	if xdg != "" {
		if v := os.Getenv(xdg); v != "" {
			return filepath.Join(v, sub), nil
		}
	}
	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(append([]string{h}, home...)...), nil
}

func ResolveDataDir(input string) (string, error) {
	dir, err := resolveDir(input, EnvDataPath, "", "", ".local", "share", "phasionary")
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "projects"), nil
}

// The per-device state directory must never be carried between machines by a
// file syncer: it holds the device identity.
func ResolveStateDir() (string, error) {
	return resolveDir("", EnvStatePath, "XDG_STATE_HOME", "phasionary", ".local", "state", "phasionary")
}

func ResolveServerDataDir(input string) (string, error) {
	return resolveDir(input, EnvServerDataPath, "XDG_DATA_HOME", "phasionary-server", ".local", "share", "phasionary-server")
}

// Accepts either a directory path or a path pointing at a `*.json` file.
func configDirFromPath(p string) string {
	if strings.HasSuffix(p, ".json") {
		return filepath.Dir(p)
	}
	return p
}

func ResolveConfigDir(input string) (string, error) {
	dir, err := resolveDir(input, EnvConfigPath, "XDG_CONFIG_HOME", "phasionary", ".config", "phasionary")
	if err != nil {
		return "", err
	}
	return configDirFromPath(dir), nil
}

// ResolveConfigPath returns the full path to config.json.
func ResolveConfigPath(input string) (string, error) {
	dir, err := ResolveConfigDir(input)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}
