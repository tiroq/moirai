package profile

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"moirai/internal/util"
)

func TestDiffProfiles_ReturnsColoredDiff(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	writeProfileFile(t, dir, "a", `{"a":1}`+"\n")
	writeProfileFile(t, dir, "b", `{"a":2}`+"\n")

	out, err := DiffProfiles(dir, "a", "b")
	if err != nil {
		t.Fatalf("DiffProfiles: %v", err)
	}
	if out == "" || !strings.Contains(out, "diff --git") {
		t.Fatalf("unexpected diff output:\n%s", out)
	}
}

func TestDiffProfileAgainstFile_ReturnsColoredDiff(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	writeProfileFile(t, dir, "p", `{"x":1}`+"\n")
	otherRel := "other.json"
	if err := os.WriteFile(filepath.Join(dir, otherRel), []byte(`{"x":2}`+"\n"), 0o644); err != nil {
		t.Fatalf("write other: %v", err)
	}

	out, err := DiffProfileAgainstFile(dir, "p", otherRel)
	if err != nil {
		t.Fatalf("DiffProfileAgainstFile: %v", err)
	}
	if out == "" || !strings.Contains(out, "diff --git") {
		t.Fatalf("unexpected diff output:\n%s", out)
	}
}

func TestDiffProfiles_ReportsMissingGit(t *testing.T) {
	dir := t.TempDir()
	writeProfileFile(t, dir, "a", `{"a":1}`+"\n")
	writeProfileFile(t, dir, "b", `{"a":2}`+"\n")

	prev := util.GitBin
	util.GitBin = "definitely-not-git"
	t.Cleanup(func() { util.GitBin = prev })

	_, err := DiffProfiles(dir, "a", "b")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, util.ErrGitNotAvailable) {
		t.Fatalf("expected ErrGitNotAvailable, got: %v", err)
	}
	if !strings.Contains(err.Error(), "git is required for diff") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func writeProfileFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, profilePrefix+name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write profile %q: %v", name, err)
	}
}
