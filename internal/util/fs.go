package util

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandUser expands a leading "~" to the current user's home directory.
func ExpandUser(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if path == "~" {
			return home, nil
		}
		rest := strings.TrimPrefix(path, "~/")
		return filepath.Join(home, rest), nil
	}
	return path, nil
}

// FileExists reports whether a file system entry exists at path.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ListDir returns the directory entries for dir.
func ListDir(dir string) ([]os.DirEntry, error) {
	return os.ReadDir(dir)
}
