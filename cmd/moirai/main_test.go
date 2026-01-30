package main

import (
	"bytes"
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

func TestParseGlobalFlagsVersion(t *testing.T) {
	remaining, flags, err := parseGlobalFlags([]string{"--version"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(remaining) != 0 {
		t.Fatalf("expected no remaining args, got %v", remaining)
	}
	if !flags.ShowVersion {
		t.Fatalf("expected ShowVersion to be true")
	}
}

func TestShouldPrintVersion(t *testing.T) {
	if !shouldPrintVersion([]string{"version"}, globalFlags{}) {
		t.Fatalf("expected version command to print version")
	}
	if !shouldPrintVersion([]string{"list"}, globalFlags{ShowVersion: true}) {
		t.Fatalf("expected --version flag to print version")
	}
}

func TestPrintVersion(t *testing.T) {
	original := app.Version
	app.Version = "test-version"
	defer func() {
		app.Version = original
	}()

	var buf bytes.Buffer
	printVersion(&buf)
	if got := buf.String(); got != "test-version\n" {
		t.Fatalf("expected version output to be %q, got %q", "test-version\n", got)
	}
}
