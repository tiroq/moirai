package link

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func requireSymlink(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	targetPath := filepath.Join(dir, "target")
	if err := os.WriteFile(targetPath, []byte("{}"), 0o600); err != nil {
		t.Fatalf("write target: %v", err)
	}
	linkPath := filepath.Join(dir, "link")
	if err := os.Symlink("target", linkPath); err != nil {
		t.Skipf("symlink not supported: %v", err)
	}
}

func TestApplyProfileCreatesSymlinkWhenMissing(t *testing.T) {
	requireSymlink(t)
	dir := t.TempDir()

	profileName := "alpha"
	profileFile := "oh-my-opencode.json." + profileName
	if err := os.WriteFile(filepath.Join(dir, profileFile), []byte("{}"), 0o600); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	if err := ApplyProfile(dir, profileName); err != nil {
		t.Fatalf("ApplyProfile: %v", err)
	}

	activePath := filepath.Join(dir, "oh-my-opencode.json")
	info, err := os.Lstat(activePath)
	if err != nil {
		t.Fatalf("lstat active: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected active to be symlink")
	}
	target, err := os.Readlink(activePath)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	if target != profileFile {
		t.Fatalf("expected link to %q, got %q", profileFile, target)
	}
}

func TestApplyProfileReplacesSymlink(t *testing.T) {
	requireSymlink(t)
	dir := t.TempDir()

	profileOld := "old"
	profileNew := "new"
	profileOldFile := "oh-my-opencode.json." + profileOld
	profileNewFile := "oh-my-opencode.json." + profileNew

	if err := os.WriteFile(filepath.Join(dir, profileOldFile), []byte("{}"), 0o600); err != nil {
		t.Fatalf("write old profile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, profileNewFile), []byte("{}"), 0o600); err != nil {
		t.Fatalf("write new profile: %v", err)
	}

	activePath := filepath.Join(dir, "oh-my-opencode.json")
	if err := os.Symlink(profileOldFile, activePath); err != nil {
		t.Skipf("symlink not supported: %v", err)
	}

	if err := ApplyProfile(dir, profileNew); err != nil {
		t.Fatalf("ApplyProfile: %v", err)
	}

	target, err := os.Readlink(activePath)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	if target != profileNewFile {
		t.Fatalf("expected link to %q, got %q", profileNewFile, target)
	}
}

func TestApplyProfileBacksUpRegularFile(t *testing.T) {
	requireSymlink(t)
	dir := t.TempDir()

	profileName := "backup"
	profileFile := "oh-my-opencode.json." + profileName
	if err := os.WriteFile(filepath.Join(dir, profileFile), []byte("{}"), 0o600); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	activePath := filepath.Join(dir, "oh-my-opencode.json")
	original := []byte("{\"k\":\"v\"}")
	if err := os.WriteFile(activePath, original, 0o600); err != nil {
		t.Fatalf("write active: %v", err)
	}

	if err := ApplyProfile(dir, profileName); err != nil {
		t.Fatalf("ApplyProfile: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	var backupName string
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "oh-my-opencode.json.bak.") {
			backupName = entry.Name()
			break
		}
	}
	if backupName == "" {
		t.Fatalf("expected backup file")
	}

	backupPath := filepath.Join(dir, backupName)
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(backupContent) != string(original) {
		t.Fatalf("backup content mismatch")
	}

	info, err := os.Lstat(activePath)
	if err != nil {
		t.Fatalf("lstat active: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected active to be symlink")
	}
	target, err := os.Readlink(activePath)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	if target != profileFile {
		t.Fatalf("expected link to %q, got %q", profileFile, target)
	}
}

func TestApplyProfileMissingProfile(t *testing.T) {
	dir := t.TempDir()

	err := ApplyProfile(dir, "missing")
	if err == nil {
		t.Fatalf("expected error for missing profile")
	}
}

func TestApplyProfileActiveNotRegular(t *testing.T) {
	requireSymlink(t)
	dir := t.TempDir()

	profileName := "diractive"
	profileFile := "oh-my-opencode.json." + profileName
	if err := os.WriteFile(filepath.Join(dir, profileFile), []byte("{}"), 0o600); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	activePath := filepath.Join(dir, "oh-my-opencode.json")
	if err := os.Mkdir(activePath, 0o700); err != nil {
		t.Fatalf("mkdir active: %v", err)
	}

	err := ApplyProfile(dir, profileName)
	if err == nil {
		t.Fatalf("expected error for non-regular active path")
	}
}
