package backup

import (
	"path/filepath"

	"moirai/internal/util"
)

const activeFileName = "oh-my-opencode.json"

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
