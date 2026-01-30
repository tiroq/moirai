package util

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGitDiffNoIndexMissingGit(t *testing.T) {
	original := GitBin
	GitBin = "definitely-not-a-real-git-binary"
	t.Cleanup(func() {
		GitBin = original
	})

	if _, err := GitDiffNoIndex("old", "new"); !errors.Is(err, ErrGitNotAvailable) {
		t.Fatalf("expected ErrGitNotAvailable, got %v", err)
	}
}

func TestGitDiffNoIndexWithDiff(t *testing.T) {
	if _, err := exec.LookPath(GitBin); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	oldPath := filepath.Join(dir, "old.txt")
	newPath := filepath.Join(dir, "new.txt")
	if err := os.WriteFile(oldPath, []byte("old"), 0o600); err != nil {
		t.Fatalf("write old: %v", err)
	}
	if err := os.WriteFile(newPath, []byte("new"), 0o600); err != nil {
		t.Fatalf("write new: %v", err)
	}

	diff, err := GitDiffNoIndex(oldPath, newPath)
	if err != nil {
		t.Fatalf("GitDiffNoIndex: %v", err)
	}
	if diff == "" {
		t.Fatal("expected diff output")
	}
}

func TestGitDiffNoIndexNoDiff(t *testing.T) {
	if _, err := exec.LookPath(GitBin); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	pathA := filepath.Join(dir, "same-a.txt")
	pathB := filepath.Join(dir, "same-b.txt")
	if err := os.WriteFile(pathA, []byte("same"), 0o600); err != nil {
		t.Fatalf("write a: %v", err)
	}
	if err := os.WriteFile(pathB, []byte("same"), 0o600); err != nil {
		t.Fatalf("write b: %v", err)
	}

	diff, err := GitDiffNoIndex(pathA, pathB)
	if err != nil {
		t.Fatalf("GitDiffNoIndex: %v", err)
	}
	if diff != "" {
		t.Fatalf("expected empty diff, got %q", diff)
	}
}
