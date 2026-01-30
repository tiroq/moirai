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

func TestRestoreProfileFromBackupWritesContent(t *testing.T) {
	dir := t.TempDir()
	profileName := "restore-write"
	profilePath := filepath.Join(dir, profilePrefix+profileName)
	original := []byte("original")
	restored := []byte("restored")

	if err := os.WriteFile(profilePath, original, 0o600); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	backupName := profilePrefix + profileName + backupMarker + "20240101-000000"
	backupPath := filepath.Join(dir, backupName)
	if err := os.WriteFile(backupPath, restored, 0o600); err != nil {
		t.Fatalf("write backup: %v", err)
	}

	if _, err := RestoreProfileFromBackup(dir, profileName, backupName); err != nil {
		t.Fatalf("RestoreProfileFromBackup: %v", err)
	}

	updated, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatalf("read profile: %v", err)
	}
	if string(updated) != string(restored) {
		t.Fatalf("profile content mismatch: %q", string(updated))
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".tmp-") {
			t.Fatalf("unexpected temp file %s", entry.Name())
		}
	}
}

func TestRestoreProfileFromBackupCreatesPreBackup(t *testing.T) {
	dir := t.TempDir()
	profileName := "restore-prebackup"
	profilePath := filepath.Join(dir, profilePrefix+profileName)
	original := []byte("original")
	restored := []byte("restored")

	if err := os.WriteFile(profilePath, original, 0o600); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	backupName := profilePrefix + profileName + backupMarker + "20240102-000000"
	backupPath := filepath.Join(dir, backupName)
	if err := os.WriteFile(backupPath, restored, 0o600); err != nil {
		t.Fatalf("write backup: %v", err)
	}

	preBackupPath, err := RestoreProfileFromBackup(dir, profileName, backupName)
	if err != nil {
		t.Fatalf("RestoreProfileFromBackup: %v", err)
	}

	base := filepath.Base(preBackupPath)
	expectedPrefix := profilePrefix + profileName + backupMarker
	if !strings.HasPrefix(base, expectedPrefix) {
		t.Fatalf("expected prefix %s in %s", expectedPrefix, base)
	}

	preBackupContent, err := os.ReadFile(preBackupPath)
	if err != nil {
		t.Fatalf("read pre-backup: %v", err)
	}
	if string(preBackupContent) != string(original) {
		t.Fatalf("pre-backup content mismatch: %q", string(preBackupContent))
	}
}

func TestRestoreProfileFromBackupRejectsWrongPrefix(t *testing.T) {
	dir := t.TempDir()
	profileName := "restore-prefix"
	profilePath := filepath.Join(dir, profilePrefix+profileName)

	if err := os.WriteFile(profilePath, []byte("profile"), 0o600); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	backupName := profilePrefix + "other" + backupMarker + "20240103-000000"
	if err := os.WriteFile(filepath.Join(dir, backupName), []byte("backup"), 0o600); err != nil {
		t.Fatalf("write backup: %v", err)
	}

	if _, err := RestoreProfileFromBackup(dir, profileName, backupName); err == nil {
		t.Fatal("expected error for mismatched backup prefix")
	}
}

func TestRestoreProfileFromBackupRequiresFrom(t *testing.T) {
	dir := t.TempDir()
	profileName := "restore-missing-from"
	profilePath := filepath.Join(dir, profilePrefix+profileName)

	if err := os.WriteFile(profilePath, []byte("profile"), 0o600); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	if _, err := RestoreProfileFromBackup(dir, profileName, ""); err == nil {
		t.Fatal("expected error for empty backup path")
	}
}

func TestRestoreProfileFromBackupRejectsMissingBackup(t *testing.T) {
	dir := t.TempDir()
	profileName := "restore-missing-backup"
	profilePath := filepath.Join(dir, profilePrefix+profileName)

	if err := os.WriteFile(profilePath, []byte("profile"), 0o600); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	if _, err := RestoreProfileFromBackup(dir, profileName, "missing.bak"); err == nil {
		t.Fatal("expected error for missing backup")
	}
}

func TestRestoreProfileFromBackupRejectsBackupOutsideConfigDir(t *testing.T) {
	dir := t.TempDir()
	otherDir := t.TempDir()
	profileName := "restore-outside"
	profilePath := filepath.Join(dir, profilePrefix+profileName)

	if err := os.WriteFile(profilePath, []byte("profile"), 0o600); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	backupName := profilePrefix + profileName + backupMarker + "20240104-000000"
	backupPath := filepath.Join(otherDir, backupName)
	if err := os.WriteFile(backupPath, []byte("backup"), 0o600); err != nil {
		t.Fatalf("write backup: %v", err)
	}

	if _, err := RestoreProfileFromBackup(dir, profileName, backupPath); err == nil {
		t.Fatal("expected error for backup outside config dir")
	}
}

func TestRestoreProfileFromBackupWithAbsolutePath(t *testing.T) {
	dir := t.TempDir()
	profileName := "restore-absolute"
	profilePath := filepath.Join(dir, profilePrefix+profileName)
	original := []byte("original")
	restored := []byte("restored")

	if err := os.WriteFile(profilePath, original, 0o600); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	backupName := profilePrefix + profileName + backupMarker + "20240105-000000"
	backupPath := filepath.Join(dir, backupName)
	if err := os.WriteFile(backupPath, restored, 0o600); err != nil {
		t.Fatalf("write backup: %v", err)
	}

	if _, err := RestoreProfileFromBackup(dir, profileName, backupPath); err != nil {
		t.Fatalf("RestoreProfileFromBackup: %v", err)
	}

	updated, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatalf("read profile: %v", err)
	}
	if string(updated) != string(restored) {
		t.Fatalf("profile content mismatch: %q", string(updated))
	}
}
