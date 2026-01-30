package link

import (
	"fmt"
	"os"
	"path/filepath"

	"moirai/internal/backup"
)

// ApplyProfile switches the active config symlink to the selected profile.
func ApplyProfile(dir, profileName string) error {
	if profileName == "" {
		return fmt.Errorf("profile name is required")
	}

	targetName := profilePrefix + profileName
	targetPath := filepath.Join(dir, targetName)
	targetInfo, err := os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("profile %q not found", profileName)
		}
		return err
	}
	if targetInfo.IsDir() {
		return fmt.Errorf("profile %q is a directory", profileName)
	}

	activePath := filepath.Join(dir, activeFileName)
	info, err := os.Lstat(activePath)
	if err != nil {
		if os.IsNotExist(err) {
			return os.Symlink(targetName, activePath)
		}
		return err
	}

	if info.Mode()&os.ModeSymlink != 0 {
		if err := os.Remove(activePath); err != nil {
			return err
		}
		return os.Symlink(targetName, activePath)
	}

	if !info.Mode().IsRegular() {
		return fmt.Errorf("active config is not a regular file")
	}

	if _, err := backup.BackupActive(dir); err != nil {
		return fmt.Errorf("backup active config: %w", err)
	}
	if err := os.Remove(activePath); err != nil {
		return err
	}
	return os.Symlink(targetName, activePath)
}
