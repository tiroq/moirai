package profile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"moirai/internal/util"
)

// DiffProfiles returns the colored diff between two profiles.
func DiffProfiles(dir, profileA, profileB string) (string, error) {
	if profileA == "" || profileB == "" {
		return "", fmt.Errorf("profile name is required")
	}

	pathA := filepath.Join(dir, profilePrefix+profileA)
	pathB := filepath.Join(dir, profilePrefix+profileB)

	if _, err := os.Stat(pathA); err != nil {
		return "", err
	}
	if _, err := os.Stat(pathB); err != nil {
		return "", err
	}

	diff, err := util.GitDiffNoIndex(pathA, pathB)
	if err != nil {
		if errors.Is(err, util.ErrGitNotAvailable) {
			return "", fmt.Errorf("git is required for diff: %w", err)
		}
		return "", err
	}
	return diff, nil
}

// DiffProfileAgainstFile returns the colored diff between a profile and a file.
func DiffProfileAgainstFile(dir, profileName, other string) (string, error) {
	if profileName == "" || other == "" {
		return "", fmt.Errorf("profile name is required")
	}

	profilePath := filepath.Join(dir, profilePrefix+profileName)
	if _, err := os.Stat(profilePath); err != nil {
		return "", err
	}

	otherPath := other
	if !filepath.IsAbs(otherPath) {
		otherPath = filepath.Join(dir, other)
	}
	if _, err := os.Stat(otherPath); err != nil {
		return "", err
	}

	diff, err := util.GitDiffNoIndex(otherPath, profilePath)
	if err != nil {
		if errors.Is(err, util.ErrGitNotAvailable) {
			return "", fmt.Errorf("git is required for diff: %w", err)
		}
		return "", err
	}
	return diff, nil
}
