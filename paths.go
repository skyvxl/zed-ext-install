package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// ZedPaths holds all relevant paths for the Zed extensions directory.
type ZedPaths struct {
	Base      string // e.g. ~/Library/Application Support/Zed/extensions/
	Installed string // Base/installed/
	Index     string // Base/index.json
}

func GetZedPaths() (*ZedPaths, error) {
	base, err := getExtensionsBase()
	if err != nil {
		return nil, err
	}
	return &ZedPaths{
		Base:      base,
		Installed: filepath.Join(base, "installed"),
		Index:     filepath.Join(base, "index.json"),
	}, nil
}

func getExtensionsBase() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		return filepath.Join(home, "Library", "Application Support", "Zed", "extensions"), nil
	case "linux":
		if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
			return filepath.Join(dir, "zed", "extensions"), nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		return filepath.Join(home, ".local", "share", "zed", "extensions"), nil
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}
