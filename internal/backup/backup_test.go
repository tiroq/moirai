package backup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBackupProfileCreatesBackup(t *testing.T) {
	dir := t.TempDir()
	profileName := "alpha"
	profileFile := filepath.Join(dir, profilePrefix+profileName)

	if err := os.WriteFile(profileFile, []byte("data"), 0o600); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	backupPath, err := BackupProfile(dir, profileName)
	if err != nil {
		t.Fatalf("BackupProfile: %v", err)
	}

	base := filepath.Base(backupPath)
	if !strings.Contains(base, backupMarker) {
		t.Fatalf("expected backup marker in %s", base)
	}

	expectedPrefix := profilePrefix + profileName + backupMarker
	if !strings.HasPrefix(base, expectedPrefix) {
		t.Fatalf("expected prefix %s in %s", expectedPrefix, base)
	}
}

func TestBackupProfileCopiesContent(t *testing.T) {
	dir := t.TempDir()
	profileName := "beta"
	profileFile := filepath.Join(dir, profilePrefix+profileName)
	content := []byte("hello backup")

	if err := os.WriteFile(profileFile, content, 0o600); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	backupPath, err := BackupProfile(dir, profileName)
	if err != nil {
		t.Fatalf("BackupProfile: %v", err)
	}

	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}

	if string(backupContent) != string(content) {
		t.Fatalf("backup content mismatch: %q", string(backupContent))
	}
}

func TestListProfileBackupsNewestFirst(t *testing.T) {
	dir := t.TempDir()
	profileName := "gamma"

	backups := []string{
		profilePrefix + profileName + backupMarker + "20240101-000000",
		profilePrefix + profileName + backupMarker + "20240103-000000",
		profilePrefix + profileName + backupMarker + "20240102-000000",
	}
	for _, name := range backups {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600); err != nil {
			t.Fatalf("write backup %s: %v", name, err)
		}
	}

	listed, err := ListProfileBackups(dir, profileName)
	if err != nil {
		t.Fatalf("ListProfileBackups: %v", err)
	}

	if len(listed) != 3 {
		t.Fatalf("expected 3 backups, got %d", len(listed))
	}

	expected := []string{
		profilePrefix + profileName + backupMarker + "20240103-000000",
		profilePrefix + profileName + backupMarker + "20240102-000000",
		profilePrefix + profileName + backupMarker + "20240101-000000",
	}
	for i, name := range expected {
		if listed[i] != name {
			t.Fatalf("expected %s at %d, got %s", name, i, listed[i])
		}
	}
}

func TestListProfileBackupsIgnoresOtherProfiles(t *testing.T) {
	dir := t.TempDir()
	profileName := "delta"

	files := []string{
		profilePrefix + profileName + backupMarker + "20240101-000000",
		profilePrefix + "other" + backupMarker + "20240102-000000",
		"notes.txt",
	}
	for _, name := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600); err != nil {
			t.Fatalf("write file %s: %v", name, err)
		}
	}

	listed, err := ListProfileBackups(dir, profileName)
	if err != nil {
		t.Fatalf("ListProfileBackups: %v", err)
	}

	if len(listed) != 1 {
		t.Fatalf("expected 1 backup, got %d", len(listed))
	}
	if listed[0] != profilePrefix+profileName+backupMarker+"20240101-000000" {
		t.Fatalf("unexpected backup %s", listed[0])
	}
}

func TestBackupProfileRequiresName(t *testing.T) {
	if _, err := BackupProfile(t.TempDir(), ""); err == nil {
		t.Fatal("expected error for empty profile name")
	}
}

func TestListProfileBackupsRequiresName(t *testing.T) {
	if _, err := ListProfileBackups(t.TempDir(), ""); err == nil {
		t.Fatal("expected error for empty profile name")
	}
}

func TestBackupActiveCopiesContent(t *testing.T) {
	dir := t.TempDir()
	activePath := filepath.Join(dir, activeFileName)
	content := []byte("active data")

	if err := os.WriteFile(activePath, content, 0o600); err != nil {
		t.Fatalf("write active: %v", err)
	}

	backupPath, err := BackupActive(dir)
	if err != nil {
		t.Fatalf("BackupActive: %v", err)
	}

	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}

	if string(backupContent) != string(content) {
		t.Fatalf("backup content mismatch: %q", string(backupContent))
	}
}
