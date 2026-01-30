package main

import (
	"os"
	"path/filepath"
	"testing"

	"moirai/internal/app"
	"moirai/internal/backup"
)

func TestAutofillRefusedWhenDisabled(t *testing.T) {
	configDir := t.TempDir()
	profileName := "alpha"
	profilePath := filepath.Join(configDir, "oh-my-opencode.json."+profileName)
	if err := os.WriteFile(profilePath, []byte(`{"agents":{}}`), 0o600); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	config := app.AppConfig{ConfigDir: configDir, EnableAutofill: false}
	exitCode, err := runAutofill(config, profileName, "openai")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if exitCode != 3 {
		t.Fatalf("expected exit code 3, got %d", exitCode)
	}
}

func TestAutofillCreatesBackup(t *testing.T) {
	configDir := t.TempDir()
	profileName := "alpha"
	profilePath := filepath.Join(configDir, "oh-my-opencode.json."+profileName)
	original := []byte(`{"agents":{"sisyphus":{"model":""}}}`)
	if err := os.WriteFile(profilePath, original, 0o600); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	config := app.AppConfig{ConfigDir: configDir, EnableAutofill: true}
	exitCode, err := runAutofill(config, profileName, "openai")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	backups, err := backup.ListProfileBackups(configDir, profileName)
	if err != nil {
		t.Fatalf("ListProfileBackups: %v", err)
	}
	if len(backups) != 1 {
		t.Fatalf("expected 1 backup, got %d", len(backups))
	}
	backupPath := filepath.Join(configDir, backups[0])
	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(backupData) != string(original) {
		t.Fatalf("backup does not match original content")
	}
}
