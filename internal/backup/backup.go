package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"moirai/internal/profile"
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

// RestoreProfileFromBackup restores the profile file from a backup in dir.
func RestoreProfileFromBackup(dir, profileName string, from string) (string, error) {
	if profileName == "" {
		return "", fmt.Errorf("profile name is required")
	}
	if from == "" {
		return "", fmt.Errorf("backup path is required")
	}

	profileFile := profilePrefix + profileName
	profilePath := filepath.Join(dir, profileFile)
	if _, err := os.Stat(profilePath); err != nil {
		return "", err
	}

	backupPath, err := resolveBackupPath(dir, from)
	if err != nil {
		return "", err
	}

	inDir, err := isInConfigDir(backupPath, dir)
	if err != nil {
		return "", err
	}
	if !inDir {
		return "", fmt.Errorf("backup must be in config dir")
	}

	base := filepath.Base(backupPath)
	prefix := profileFile + backupMarker
	if !strings.HasPrefix(base, prefix) || len(base) == len(prefix) {
		return "", fmt.Errorf("backup does not match profile")
	}

	preBackupPath, err := BackupProfile(dir, profileName)
	if err != nil {
		return "", err
	}

	backupInfo, err := os.Stat(backupPath)
	if err != nil {
		return "", err
	}
	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		return "", err
	}
	if err := profile.SaveProfileAtomic(profilePath, backupData, backupInfo.Mode().Perm()); err != nil {
		return "", err
	}

	return preBackupPath, nil
}

func resolveBackupPath(dir, from string) (string, error) {
	if _, err := os.Stat(from); err == nil {
		return from, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}

	candidate := filepath.Join(dir, from)
	if _, err := os.Stat(candidate); err != nil {
		return "", err
	}
	return candidate, nil
}

func isInConfigDir(path, dir string) (bool, error) {
	pathAbs, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}
	dirAbs, err := filepath.Abs(dir)
	if err != nil {
		return false, err
	}

	pathResolved := pathAbs
	if resolved, err := filepath.EvalSymlinks(pathAbs); err == nil {
		pathResolved = resolved
	}
	dirResolved := dirAbs
	if resolved, err := filepath.EvalSymlinks(dirAbs); err == nil {
		dirResolved = resolved
	}

	return filepath.Dir(pathResolved) == dirResolved, nil
}
