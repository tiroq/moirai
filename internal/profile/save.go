package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SaveProfileAtomic writes a config to path using a temp file and rename.
func SaveProfileAtomic(path string, cfg *RootConfig) error {
	if cfg == nil {
		return fmt.Errorf("config is required")
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	return SaveProfileDataAtomic(path, data, info.Mode().Perm())
}

// SaveProfileDataAtomic writes raw data to path using a temp file and rename.
func SaveProfileDataAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tempFile, err := os.CreateTemp(dir, ".tmp-")
	if err != nil {
		return err
	}
	tempName := tempFile.Name()
	defer func() {
		_ = os.Remove(tempName)
	}()

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
