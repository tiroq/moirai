package tui

import (
	"moirai/internal/backup"
	"moirai/internal/link"
	"moirai/internal/profile"
)

type modelActions struct {
	applyProfile          func(dir, profileName string) error
	listProfileBackups    func(dir, profileName string) ([]string, error)
	activeProfile         func(dir string) (string, bool, error)
	diffAgainstLastBackup func(dir, profileName string) (string, bool, error)
	diffBetweenProfiles   func(dir, profileA, profileB string) (string, error)
}

func defaultActions() modelActions {
	return modelActions{
		applyProfile:          link.ApplyProfile,
		listProfileBackups:    backup.ListProfileBackups,
		activeProfile:         link.ActiveProfile,
		diffAgainstLastBackup: diffAgainstLastBackup,
		diffBetweenProfiles:   profile.DiffProfiles,
	}
}

func diffAgainstLastBackup(dir, profileName string) (string, bool, error) {
	backupName, ok, err := backup.LatestProfileBackup(dir, profileName)
	if err != nil {
		return "", false, err
	}
	if !ok {
		return "", false, nil
	}
	diff, err := profile.DiffProfileAgainstFile(dir, profileName, backupName)
	if err != nil {
		return "", true, err
	}
	return diff, true, nil
}
