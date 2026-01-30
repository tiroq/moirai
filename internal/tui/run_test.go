package tui

import (
	"os"
	"path/filepath"
	"testing"

	"moirai/internal/app"

	tea "github.com/charmbracelet/bubbletea"
)

func TestLoadModelUsesProfilesAndActive(t *testing.T) {
	dir := t.TempDir()
	alphaPath := filepath.Join(dir, "oh-my-opencode.json.alpha")
	betaPath := filepath.Join(dir, "oh-my-opencode.json.beta")
	if err := os.WriteFile(alphaPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("write alpha: %v", err)
	}
	if err := os.WriteFile(betaPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("write beta: %v", err)
	}
	activePath := filepath.Join(dir, "oh-my-opencode.json")
	if err := os.Symlink(betaPath, activePath); err != nil {
		t.Fatalf("symlink active: %v", err)
	}

	m, err := loadModel(app.AppConfig{ConfigDir: dir})
	if err != nil {
		t.Fatalf("loadModel: %v", err)
	}
	if !m.hasActive || m.activeName != "beta" {
		t.Fatalf("expected active beta, got %v %q", m.hasActive, m.activeName)
	}
	if len(m.profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(m.profiles))
	}
}

func TestRunInvokesProgramRunner(t *testing.T) {
	origRunner := programRunner
	t.Cleanup(func() {
		programRunner = origRunner
	})

	called := false
	programRunner = func(m tea.Model) error {
		called = true
		return nil
	}

	dir := t.TempDir()
	if err := Run(app.AppConfig{ConfigDir: dir}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !called {
		t.Fatalf("expected programRunner to be called")
	}
}
