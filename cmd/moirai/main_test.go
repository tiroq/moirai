package main

import (
	"bytes"
	"path/filepath"
	"testing"

	"moirai/internal/app"
)

func TestRunLaunchesTUIWhenNoArgs(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	called := false
	stub := func(config app.AppConfig) error {
		called = true
		expected := filepath.Join(home, ".config", "opencode")
		if config.ConfigDir != expected {
			t.Fatalf("expected config dir %q, got %q", expected, config.ConfigDir)
		}
		return nil
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := run([]string{"moirai"}, stub, stdout, stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if !called {
		t.Fatalf("expected TUI runner to be called")
	}
}

func TestRunSkipsTUIWhenArgsExist(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	called := false
	stub := func(config app.AppConfig) error {
		called = true
		return nil
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := run([]string{"moirai", "help"}, stub, stdout, stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if called {
		t.Fatalf("expected TUI runner not to be called")
	}
	if stdout.Len() == 0 {
		t.Fatalf("expected help output")
	}
}
