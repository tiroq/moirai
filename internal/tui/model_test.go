package tui

import (
	"strings"
	"testing"

	"moirai/internal/profile"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModelSelectsActive(t *testing.T) {
	profiles := []profile.ProfileInfo{
		{Name: "alpha"},
		{Name: "beta"},
	}
	m := newModel("/config", false, profiles, "beta", true)
	if m.selected != 1 {
		t.Fatalf("expected selected 1, got %d", m.selected)
	}
}

func TestModelUpdateMovement(t *testing.T) {
	profiles := []profile.ProfileInfo{
		{Name: "alpha"},
		{Name: "beta"},
	}
	m := newModelWithActions("/config", false, profiles, "", false, stubActions())

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = updated.(model)
	if m.selected != 1 {
		t.Fatalf("expected selected 1 after down, got %d", m.selected)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m = updated.(model)
	if m.selected != 0 {
		t.Fatalf("expected selected 0 after up, got %d", m.selected)
	}
}

func TestProfilesApplyTriggersApply(t *testing.T) {
	profiles := []profile.ProfileInfo{
		{Name: "alpha"},
		{Name: "beta"},
	}

	called := false
	actions := stubActions()
	actions.applyProfile = func(dir, profileName string) error {
		called = true
		if dir != "/config" {
			t.Fatalf("expected dir /config, got %q", dir)
		}
		if profileName != "beta" {
			t.Fatalf("expected profile beta, got %q", profileName)
		}
		return nil
	}
	actions.activeProfile = func(dir string) (string, bool, error) {
		return "beta", true, nil
	}

	m := newModelWithActions("/config", false, profiles, "beta", true, actions)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("expected apply command")
	}
	msg := cmd()
	updated, _ = updated.(model).Update(msg)
	m = updated.(model)
	if !called {
		t.Fatalf("expected apply to be called")
	}
	if m.status == "" {
		t.Fatalf("expected status after apply")
	}
}

func TestApplyRefreshesActiveProfile(t *testing.T) {
	profiles := []profile.ProfileInfo{
		{Name: "alpha"},
	}

	refreshed := false
	actions := stubActions()
	actions.applyProfile = func(dir, profileName string) error {
		return nil
	}
	actions.activeProfile = func(dir string) (string, bool, error) {
		refreshed = true
		return "alpha", true, nil
	}

	m := newModelWithActions("/config", false, profiles, "", false, actions)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	msg := cmd()
	updated, _ = updated.(model).Update(msg)
	m = updated.(model)

	if !refreshed {
		t.Fatalf("expected active profile refresh")
	}
	if !m.hasActive || m.activeName != "alpha" {
		t.Fatalf("expected active profile alpha, got %v %q", m.hasActive, m.activeName)
	}
}

func TestBackupsScreenUsesListAndHandlesEmpty(t *testing.T) {
	profiles := []profile.ProfileInfo{
		{Name: "alpha"},
	}

	called := false
	actions := stubActions()
	actions.listProfileBackups = func(dir, profileName string) ([]string, error) {
		called = true
		if profileName != "alpha" {
			t.Fatalf("expected profile alpha, got %q", profileName)
		}
		return nil, nil
	}

	m := newModelWithActions("/config", false, profiles, "", false, actions)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	msg := cmd()
	updated, _ = updated.(model).Update(msg)
	m = updated.(model)

	if !called {
		t.Fatalf("expected backups list to be called")
	}
	if m.screen != screenBackups {
		t.Fatalf("expected backups screen, got %v", m.screen)
	}
	if len(m.backups) != 0 {
		t.Fatalf("expected no backups, got %d", len(m.backups))
	}
	if !strings.Contains(m.View(), "(none)") {
		t.Fatalf("expected empty backups message")
	}
}

func TestDiffScreenNoBackupsMessage(t *testing.T) {
	profiles := []profile.ProfileInfo{
		{Name: "alpha"},
	}

	actions := stubActions()
	actions.diffAgainstLastBackup = func(dir, profileName string) (string, bool, error) {
		return "", false, nil
	}

	m := newModelWithActions("/config", false, profiles, "", false, actions)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	msg := cmd()
	updated, _ = updated.(model).Update(msg)
	m = updated.(model)

	if m.screen != screenDiff {
		t.Fatalf("expected diff screen, got %v", m.screen)
	}
	if !strings.Contains(m.diffMessage, "No backups") {
		t.Fatalf("expected no backups message, got %q", m.diffMessage)
	}
}

func TestModelUpdateQuit(t *testing.T) {
	m := newModelWithActions("/config", false, nil, "", false, stubActions())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatalf("expected quit command")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("expected quit message, got %T", msg)
	}
}

func stubActions() modelActions {
	return modelActions{
		applyProfile: func(dir, profileName string) error {
			return nil
		},
		listProfileBackups: func(dir, profileName string) ([]string, error) {
			return nil, nil
		},
		activeProfile: func(dir string) (string, bool, error) {
			return "", false, nil
		},
		diffAgainstLastBackup: func(dir, profileName string) (string, bool, error) {
			return "", true, nil
		},
		diffBetweenProfiles: func(dir, profileA, profileB string) (string, error) {
			return "", nil
		},
		loadProfile: func(path string) (*profile.RootConfig, error) {
			return &profile.RootConfig{}, nil
		},
		saveProfile: func(path string, cfg *profile.RootConfig) error {
			return nil
		},
		backupProfile: func(dir, profileName string) (string, error) {
			return "", nil
		},
		applyAutofill: func(cfg *profile.RootConfig, knownAgents []string, preset profile.Preset) bool {
			return profile.ApplyAutofill(cfg, knownAgents, preset)
		},
		loadModels: func() []string {
			return []string{"gpt-4o-mini"}
		},
	}
}
