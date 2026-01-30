package models

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const metaSchemaVersion = 1

type metaFile struct {
	UpdatedAt string `json:"updatedAt"`
	Source    string `json:"source"`
	Version   int    `json:"version"`
}

func cacheDir(configHome string) string {
	return filepath.Join(configHome, "opencode", "moirai")
}

func modelsPath(configHome string) string {
	return filepath.Join(cacheDir(configHome), "models.txt")
}

func metaPath(configHome string) string {
	return filepath.Join(cacheDir(configHome), "models.meta.json")
}

// LoadCachedModels reads the cached model list from disk.
// It returns ok=false when the cache is missing or empty.
func LoadCachedModels(configHome string) ([]string, bool, error) {
	data, err := os.ReadFile(modelsPath(configHome))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, false, nil
		}
		return nil, false, err
	}
	models := parseLines(data)
	if len(models) == 0 {
		return nil, false, nil
	}
	return models, true, nil
}

// SaveCachedModelsAtomic writes the model list to the cache using a temp file + rename.
func SaveCachedModelsAtomic(configHome string, models []string) error {
	dir := cacheDir(configHome)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	content := strings.Join(models, "\n")
	if content != "" {
		content += "\n"
	}
	if err := writeFileAtomic(modelsPath(configHome), []byte(content), 0o644); err != nil {
		return err
	}

	meta := metaFile{
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		Source:    "opencode models",
		Version:   metaSchemaVersion,
	}
	metaData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	metaData = append(metaData, '\n')
	if err := writeFileAtomic(metaPath(configHome), metaData, 0o644); err != nil {
		return err
	}
	return nil
}

// CacheAge returns the age of the cached model list (based on models.txt modtime).
// It returns ok=false when the cache does not exist.
func CacheAge(configHome string) (time.Duration, bool, error) {
	info, err := os.Stat(modelsPath(configHome))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return time.Since(info.ModTime()), true, nil
}

func parseLines(data []byte) []string {
	lines := strings.Split(string(data), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		out = append(out, line)
	}
	return out
}

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tempFile, err := os.CreateTemp(dir, ".tmp-")
	if err != nil {
		return err
	}
	tempName := tempFile.Name()
	defer func() { _ = os.Remove(tempName) }()

	if _, err := tempFile.Write(data); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err := tempFile.Chmod(perm); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempName, path); err != nil {
		return err
	}
	return nil
}

