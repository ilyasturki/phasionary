package config

import (
	"os"
	"path/filepath"
)

const (
	EnvDataPath = "PHASIONARY_DATA_PATH"
)

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
