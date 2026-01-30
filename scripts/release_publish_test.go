package scripts_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestReleasePublish_PushesTagToOrigin(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	tmp := t.TempDir()
	home := filepath.Join(tmp, "home")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatalf("mkdir home: %v", err)
	}

	originDir := filepath.Join(tmp, "origin.git")
	repoDir := filepath.Join(tmp, "repo")

	env := append(os.Environ(),
		"HOME="+home,
		"GIT_TERMINAL_PROMPT=0",
	)

	runGit(t, env, "", "init", "--bare", originDir)
	runGit(t, env, "", "init", repoDir)
	runGit(t, env, repoDir, "config", "user.email", "test@example.com")
	runGit(t, env, repoDir, "config", "user.name", "Test User")
	runGit(t, env, repoDir, "remote", "add", "origin", originDir)

	if err := os.WriteFile(filepath.Join(repoDir, "file.txt"), []byte("content\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	runGit(t, env, repoDir, "add", "file.txt")
	runGit(t, env, repoDir, "commit", "-m", "init")
	runGit(t, env, repoDir, "tag", "-a", "v0.1.0", "-m", "release v0.1.0")

	scriptPath := releasePublishScriptPath(t)
	if fi, err := os.Stat(scriptPath); err != nil {
		t.Fatalf("stat script: %v", err)
	} else if fi.Mode()&0o111 == 0 {
		t.Fatalf("script is not executable: %s", scriptPath)
	}

	cmd := exec.Command(scriptPath, "v0.1.0")
	cmd.Dir = repoDir
	cmd.Env = env

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("script failed: %v\noutput:\n%s", err, out)
	}

	if !strings.Contains(string(out), "Release v0.1.0 published. GitHub Actions will build assets.") {
		t.Fatalf("missing confirmation line in output:\n%s", out)
	}

	showRef := runGitOutput(t, env, originDir, "show-ref", "--tags", "v0.1.0")
	if !strings.Contains(showRef, "refs/tags/v0.1.0") {
		t.Fatalf("tag not found in origin:\n%s", showRef)
	}
}

func TestReleasePublish_FailsWhenTagMissing(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	tmp := t.TempDir()
	home := filepath.Join(tmp, "home")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatalf("mkdir home: %v", err)
	}

	originDir := filepath.Join(tmp, "origin.git")
	repoDir := filepath.Join(tmp, "repo")

	env := append(os.Environ(),
		"HOME="+home,
		"GIT_TERMINAL_PROMPT=0",
	)

	runGit(t, env, "", "init", "--bare", originDir)
	runGit(t, env, "", "init", repoDir)
	runGit(t, env, repoDir, "remote", "add", "origin", originDir)

	scriptPath := releasePublishScriptPath(t)

	cmd := exec.Command(scriptPath, "v0.1.0")
	cmd.Dir = repoDir
	cmd.Env = env

	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected failure, got success:\n%s", out)
	}
	if !strings.Contains(string(out), "tag does not exist locally: v0.1.0") {
		t.Fatalf("unexpected output:\n%s", out)
	}
}

func TestReleasePublish_RejectsInvalidVersionFormat(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	tmp := t.TempDir()
	home := filepath.Join(tmp, "home")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatalf("mkdir home: %v", err)
	}

	repoDir := filepath.Join(tmp, "repo")
	env := append(os.Environ(),
		"HOME="+home,
		"GIT_TERMINAL_PROMPT=0",
	)

	runGit(t, env, "", "init", repoDir)

	scriptPath := releasePublishScriptPath(t)

	cmd := exec.Command(scriptPath, "0.1.0")
	cmd.Dir = repoDir
	cmd.Env = env

	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected failure, got success:\n%s", out)
	}
	if !strings.Contains(string(out), "invalid version format: 0.1.0") {
		t.Fatalf("unexpected output:\n%s", out)
	}
}

func releasePublishScriptPath(t *testing.T) string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	return filepath.Join(filepath.Dir(thisFile), "release-publish.sh")
}

func runGit(t *testing.T, env []string, dir string, args ...string) {
	t.Helper()
	if _, err := runGitCmd(env, dir, args...); err != nil {
		t.Fatalf("git %s failed: %v", strings.Join(args, " "), err)
	}
}

func runGitOutput(t *testing.T, env []string, dir string, args ...string) string {
	t.Helper()
	out, err := runGitCmd(env, dir, args...)
	if err != nil {
		t.Fatalf("git %s failed: %v", strings.Join(args, " "), err)
	}
	return out
}

func runGitCmd(env []string, dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = env

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	return buf.String(), err
}
