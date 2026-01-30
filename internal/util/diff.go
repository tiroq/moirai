package util

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
)

// GitBin allows tests to override the git binary name.
var GitBin = "git"

// ErrGitNotAvailable indicates git was not found in PATH.
var ErrGitNotAvailable = errors.New("git not available")

// GitDiffNoIndex returns the colored git diff between two files.
func GitDiffNoIndex(oldPath, newPath string) (string, error) {
	if _, err := exec.LookPath(GitBin); err != nil {
		return "", fmt.Errorf("%w: %v", ErrGitNotAvailable, err)
	}

	args := []string{"--no-pager", "diff", "--no-index", "--color=always", oldPath, newPath}
	stdout, stderr, err := RunCommand(context.Background(), GitBin, args...)
	if err == nil {
		return stdout, nil
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ProcessState != nil && exitErr.ProcessState.ExitCode() == 1 {
			return stdout, nil
		}
	}

	if stderr != "" {
		return "", fmt.Errorf("git diff failed: %s", stderr)
	}
	return "", fmt.Errorf("git diff failed: %w", err)
}
