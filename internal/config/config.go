package config

import (
	"os"
	"path/filepath"
)

const (
	EnvDataPath   = "PHASIONARY_DATA_PATH"
	EnvConfigPath = "PHASIONARY_CONFIG_PATH"
)

// Config holds user preferences. Fields will be added as needed.
type Config struct{}

// DefaultConfig returns a Config with default values.
func DefaultConfig() Config {
	return Config{}
}

func ResolveDataDir(input string) (string, error) {
	if input != "" {
		return input, nil
	}
	if env := os.Getenv(EnvDataPath); env != "" {
		return env, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "phasionary"), nil
}

// ResolveConfigDir returns the config directory path.
// Priority: input > PHASIONARY_CONFIG_PATH > XDG_CONFIG_HOME > ~/.config/phasionary
func ResolveConfigDir(input string) (string, error) {
	if input != "" {
		return input, nil
	}
	if env := os.Getenv(EnvConfigPath); env != "" {
		return env, nil
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
