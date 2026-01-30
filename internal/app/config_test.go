package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigMissingFileDefaults(t *testing.T) {
	configDir := t.TempDir()

	config, err := LoadConfig(configDir, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if config.EnableAutofill {
		t.Fatalf("expected EnableAutofill to default to false")
	}
	if config.ConfigDir != filepath.Clean(configDir) {
		t.Fatalf("expected ConfigDir to be %q, got %q", filepath.Clean(configDir), config.ConfigDir)
	}
}

func TestLoadConfigValidFile(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "moirai.json")
	if err := os.WriteFile(configPath, []byte(`{"enableAutofill": true}`), 0o600); err != nil {
		t.Fatalf("expected to write config file, got %v", err)
	}

	config, err := LoadConfig(configDir, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !config.EnableAutofill {
		t.Fatalf("expected EnableAutofill to be true")
	}
}
