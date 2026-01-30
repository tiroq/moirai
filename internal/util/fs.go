package util

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
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

// Timestamp returns a local time string formatted as YYYYMMDD-HHMMSS.
func Timestamp() string {
	return time.Now().Format("20060102-150405")
}

// CopyFileAtomic copies src to dst using a temp file in dst's directory.
func CopyFileAtomic(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dir := filepath.Dir(dst)
	tempFile, err := os.CreateTemp(dir, ".tmp-")
	if err != nil {
		return err
	}
	tempName := tempFile.Name()
	defer func() {
		_ = os.Remove(tempName)
	}()

	if _, err := io.Copy(tempFile, srcFile); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err := tempFile.Chmod(info.Mode().Perm()); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempName, dst); err != nil {
		return err
	}
	return nil
}
