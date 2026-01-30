package main

import (
	"os"
	"path/filepath"
	"testing"

	"moirai/internal/app"
)

func TestCLIFlagOverridesConfigFile(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "moirai.json")
	if err := os.WriteFile(configPath, []byte(`{"enableAutofill": false}`), 0o600); err != nil {
		t.Fatalf("expected to write config file, got %v", err)
	}

	remaining, flags, err := parseGlobalFlags([]string{"list", "--enable-autofill"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(remaining) != 1 || remaining[0] != "list" {
		t.Fatalf("expected remaining args to be [list], got %v", remaining)
	}

	var override *bool
	if flags.EnableAutofillSet {
		override = &flags.EnableAutofill
	}

	config, err := app.LoadConfig(configDir, override)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !config.EnableAutofill {
		t.Fatalf("expected EnableAutofill to be true after CLI override")
	}
}
