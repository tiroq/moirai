package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"moirai/internal/util"
)

const activeFileName = "oh-my-opencode.json"
const profilePrefix = activeFileName + "."
const backupMarker = ".bak."

// BackupActive creates a backup of the active config file in dir.
func BackupActive(dir string) (string, error) {
	backupName := activeFileName + ".bak." + util.Timestamp()
	backupPath := filepath.Join(dir, backupName)
	activePath := filepath.Join(dir, activeFileName)

	if err := util.CopyFileAtomic(activePath, backupPath); err != nil {
		return "", err
	}
	return backupPath, nil
}

// BackupProfile creates a backup of the named profile in dir.
func BackupProfile(dir, profileName string) (string, error) {
	if profileName == "" {
		return "", fmt.Errorf("profile name is required")
	}
	profileFile := profilePrefix + profileName
	profilePath := filepath.Join(dir, profileFile)
	if _, err := os.Stat(profilePath); err != nil {
		return "", err
	}

	backupName := profileFile + backupMarker + util.Timestamp()
	backupPath := filepath.Join(dir, backupName)
	if err := util.CopyFileAtomic(profilePath, backupPath); err != nil {
		return "", err
	}
	return backupPath, nil
}

// ListProfileBackups returns the profile backups in dir, newest first.
func ListProfileBackups(dir, profileName string) ([]string, error) {
	if profileName == "" {
		return nil, fmt.Errorf("profile name is required")
	}
	entries, err := util.ListDir(dir)
	if err != nil {
		return nil, err
	}

	prefix := profilePrefix + profileName + backupMarker
	backups := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		if len(name) == len(prefix) {
			continue
		}
		backups = append(backups, name)
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i] > backups[j]
	})
	return backups, nil
}
