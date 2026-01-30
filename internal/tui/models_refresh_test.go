package tui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"moirai/internal/opencode"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModelPickerUsesCacheAndSkipsRefreshWhenTTLValid(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)

	cachePath := filepath.Join(configHome, "opencode", "moirai", "models.txt")
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(cachePath, []byte("cached-a\ncached-b\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	now := time.Now()
	if err := os.Chtimes(cachePath, now, now); err != nil {
		t.Fatalf("Chtimes: %v", err)
	}

	m := model{
		screen:        screenAgents,
		agentsEntries: []agentEntry{{Name: "sisyphus"}},
		agentsSelected: 0,
		actions:       normalizeActions(stubActions()),
	}
	m.actions.loadModels = loadModelList

	updated, cmd := m.openModelPicker()
	got := updated.(model)
	if got.screen != screenModels {
		t.Fatalf("expected models screen, got %v", got.screen)
	}
	if cmd != nil {
		t.Fatalf("expected no refresh cmd when cache is fresh")
	}
	if joined := strings.Join(got.modelAll, ","); joined != "cached-a,cached-b" {
		t.Fatalf("expected cached models, got %q", joined)
	}
}

func TestModelPickerFallsBackToDefaultWhenCacheMissing(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)

	m := model{
		screen:        screenAgents,
		agentsEntries: []agentEntry{{Name: "sisyphus"}},
		agentsSelected: 0,
		actions:       normalizeActions(stubActions()),
	}
	m.actions.loadModels = loadModelList

	updated, _ := m.openModelPicker()
	got := updated.(model)
	if joined := strings.Join(got.modelAll, ","); joined != strings.Join(defaultModelList(), ",") {
		t.Fatalf("expected default models, got %q", joined)
	}
}

func TestModelPickerSchedulesRefreshWhenCacheOldAndAppliesUpdate(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)

	cachePath := filepath.Join(configHome, "opencode", "moirai", "models.txt")
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(cachePath, []byte("old-model\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	then := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(cachePath, then, then); err != nil {
		t.Fatalf("Chtimes: %v", err)
	}

	restore := opencode.SetRunnerForTest(func(_ context.Context, _ string, _ ...string) ([]byte, []byte, error) {
		return []byte("new-a\nnew-b\n"), nil, nil
	})
	defer restore()

	m := model{
		screen:        screenAgents,
		agentsEntries: []agentEntry{{Name: "sisyphus"}},
		agentsSelected: 0,
		actions:       normalizeActions(stubActions()),
	}
	m.actions.loadModels = loadModelList

	updated, cmd := m.openModelPicker()
	if cmd == nil {
		t.Fatalf("expected refresh cmd when cache is old")
	}
	msg := cmd()
	if msg == nil {
		t.Fatalf("expected refresh message")
	}
	updated2, _ := updated.(model).Update(msg)
	got := updated2.(model)
	if joined := strings.Join(got.modelAll, ","); joined != "new-a,new-b" {
		t.Fatalf("expected refreshed models, got %q", joined)
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "new-a\nnew-b\n" {
		t.Fatalf("expected cache to be updated, got %q", string(data))
	}
}

func TestModelPickerManualRefreshKeyIsNonBlocking(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)

	cachePath := filepath.Join(configHome, "opencode", "moirai", "models.txt")
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(cachePath, []byte("cached\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	now := time.Now()
	if err := os.Chtimes(cachePath, now, now); err != nil {
		t.Fatalf("Chtimes: %v", err)
	}

	called := false
	restore := opencode.SetRunnerForTest(func(_ context.Context, _ string, _ ...string) ([]byte, []byte, error) {
		called = true
		return []byte("updated\n"), nil, nil
	})
	defer restore()

	m := model{
		screen:          screenModels,
		modelAll:        []string{"cached"},
		modelFiltered:   []string{"cached"},
		modelSelected:   0,
		modelTargetAgent: "sisyphus",
		actions:         normalizeActions(stubActions()),
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("R")})
	got := updated.(model)
	if got.modelSearch != "" {
		t.Fatalf("expected search unchanged, got %q", got.modelSearch)
	}
	if cmd == nil {
		t.Fatalf("expected refresh cmd")
	}
	if called {
		t.Fatalf("expected refresh to be async (runner called only when cmd executes)")
	}
	_ = cmd()
	if !called {
		t.Fatalf("expected runner to be called when cmd executes")
	}
}
