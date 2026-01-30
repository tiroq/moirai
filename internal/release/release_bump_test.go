package release

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func findRepoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("repo root not found from %q", dir)
		}
		dir = parent
	}
}

func run(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()

	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s failed: %v\n%s", name, strings.Join(args, " "), err, string(out))
	}
	return string(out)
}

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	return run(t, dir, "git", args...)
}

func initGitRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	git(t, dir, "init")
	git(t, dir, "config", "user.name", "Moirai Test")
	git(t, dir, "config", "user.email", "moirai-test@example.invalid")

	if err := os.WriteFile(filepath.Join(dir, "README.txt"), []byte("test\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	git(t, dir, "add", "README.txt")
	git(t, dir, "commit", "-m", "init")

	return dir
}

func runReleaseBump(t *testing.T, repoDir string, arg string) (string, error) {
	t.Helper()

	script := filepath.Join(findRepoRoot(t), "scripts", "release-bump.sh")
	cmd := exec.Command("bash", script, arg)
	cmd.Dir = repoDir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func tagType(t *testing.T, repoDir string, tag string) string {
	t.Helper()
	out := git(t, repoDir, "cat-file", "-t", tag)
	return strings.TrimSpace(out)
}

func tagMessage(t *testing.T, repoDir string, tag string) string {
	t.Helper()
	out := git(t, repoDir, "for-each-ref", "refs/tags/"+tag, "--format=%(contents)")
	return strings.TrimSpace(out)
}

func hasTag(t *testing.T, repoDir string, tag string) bool {
	t.Helper()
	out := git(t, repoDir, "tag", "--list", tag)
	return strings.TrimSpace(out) == tag
}

func isClean(t *testing.T, repoDir string) bool {
	t.Helper()
	out := git(t, repoDir, "status", "--porcelain")
	return strings.TrimSpace(out) == ""
}

func TestReleaseBump_NoTags_PatchCreatesAnnotatedTag(t *testing.T) {
	repoDir := initGitRepo(t)

	out, err := runReleaseBump(t, repoDir, "patch")
	if err != nil {
		t.Fatalf("expected success, got error:\n%s", out)
	}

	if !hasTag(t, repoDir, "v0.0.1") {
		t.Fatalf("expected tag v0.0.1 to exist")
	}
	if got := tagType(t, repoDir, "v0.0.1"); got != "tag" {
		t.Fatalf("expected annotated tag type 'tag', got %q", got)
	}
	if got := tagMessage(t, repoDir, "v0.0.1"); got != "release v0.0.1" {
		t.Fatalf("expected tag message %q, got %q", "release v0.0.1", got)
	}
	if !strings.Contains(out, "next: git push origin v0.0.1") {
		t.Fatalf("expected next steps in output, got:\n%s", out)
	}
	if !isClean(t, repoDir) {
		t.Fatalf("expected repo to remain clean after tagging")
	}
}

func TestReleaseBump_PatchUsesHighestVersionTag(t *testing.T) {
	repoDir := initGitRepo(t)

	git(t, repoDir, "tag", "-a", "v0.0.9", "-m", "release v0.0.9")
	git(t, repoDir, "tag", "-a", "v0.0.10", "-m", "release v0.0.10")

	out, err := runReleaseBump(t, repoDir, "patch")
	if err != nil {
		t.Fatalf("expected success, got error:\n%s", out)
	}
	if !hasTag(t, repoDir, "v0.0.11") {
		t.Fatalf("expected tag v0.0.11 to exist")
	}
}

func TestReleaseBump_ExplicitVersionCreatesTag(t *testing.T) {
	repoDir := initGitRepo(t)

	out, err := runReleaseBump(t, repoDir, "v0.1.0")
	if err != nil {
		t.Fatalf("expected success, got error:\n%s", out)
	}

	if !hasTag(t, repoDir, "v0.1.0") {
		t.Fatalf("expected tag v0.1.0 to exist")
	}
	if got := tagMessage(t, repoDir, "v0.1.0"); got != "release v0.1.0" {
		t.Fatalf("expected tag message %q, got %q", "release v0.1.0", got)
	}
}

func TestReleaseBump_AllowsUntrackedFiles(t *testing.T) {
	repoDir := initGitRepo(t)

	if err := os.WriteFile(filepath.Join(repoDir, "UNTRACKED.txt"), []byte("untracked\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	out, err := runReleaseBump(t, repoDir, "patch")
	if err != nil {
		t.Fatalf("expected success, got error:\n%s", out)
	}
	if !hasTag(t, repoDir, "v0.0.1") {
		t.Fatalf("expected tag v0.0.1 to exist")
	}
}

func TestReleaseBump_RefusesDirtyWorktree(t *testing.T) {
	repoDir := initGitRepo(t)

	if err := os.WriteFile(filepath.Join(repoDir, "README.txt"), []byte("dirty\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	out, err := runReleaseBump(t, repoDir, "patch")
	if err == nil {
		t.Fatalf("expected failure, got success:\n%s", out)
	}
	if !strings.Contains(out, "working tree is dirty") {
		t.Fatalf("expected dirty worktree error, got:\n%s", out)
	}
	if hasTag(t, repoDir, "v0.0.1") {
		t.Fatalf("expected no tag to be created on failure")
	}
}

func TestReleaseBump_RefusesInvalidVersion(t *testing.T) {
	repoDir := initGitRepo(t)

	out, err := runReleaseBump(t, repoDir, "v0.1")
	if err == nil {
		t.Fatalf("expected failure, got success:\n%s", out)
	}
	if !strings.Contains(out, "invalid version format") {
		t.Fatalf("expected invalid version error, got:\n%s", out)
	}
}

func TestReleaseBump_RefusesExistingTag(t *testing.T) {
	repoDir := initGitRepo(t)

	git(t, repoDir, "tag", "-a", "v0.1.0", "-m", "release v0.1.0")

	out, err := runReleaseBump(t, repoDir, "v0.1.0")
	if err == nil {
		t.Fatalf("expected failure, got success:\n%s", out)
	}
	if !strings.Contains(out, "tag already exists") {
		t.Fatalf("expected existing tag error, got:\n%s", out)
	}
}
