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
	actions.activeProfile = func(_ string) (string, bool, error) {
		return "beta", true, nil
	}

	m := newModelWithActions("/config", false, profiles, "beta", true, actions)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatalf("expected no command before confirmation")
	}
	m = updated.(model)
	if !m.confirm.Open {
		t.Fatalf("expected confirm to open")
	}

	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if cmd == nil {
		t.Fatalf("expected apply command after confirmation")
	}
	msg := cmd()
	updated, _ = updated.(model).Update(msg)
	m = updated.(model)
	if !called {
		t.Fatalf("expected apply to be called")
	}
	if m.status.Message == "" {
		t.Fatalf("expected status after apply")
	}
}

func TestApplyRefreshesActiveProfile(t *testing.T) {
	profiles := []profile.ProfileInfo{
		{Name: "alpha"},
	}

	refreshed := false
	actions := stubActions()
	actions.applyProfile = func(_, _ string) error {
		return nil
	}
	actions.activeProfile = func(_ string) (string, bool, error) {
		refreshed = true
		return "alpha", true, nil
	}

	m := newModelWithActions("/config", false, profiles, "", false, actions)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatalf("expected no command before confirmation")
	}
	updated, cmd = updated.(model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if cmd == nil {
		t.Fatalf("expected apply command after confirmation")
	}
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
	actions.listProfileBackups = func(_ string, profileName string) ([]string, error) {
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
	actions.diffAgainstLastBackup = func(_, _ string) (string, bool, error) {
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

func TestHelpOverlayTogglesAndBlocksMovement(t *testing.T) {
	profiles := []profile.ProfileInfo{
		{Name: "alpha"},
		{Name: "beta"},
	}
	m := newModelWithActions("/config", false, profiles, "", false, stubActions())

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	m = updated.(model)
	if !m.helpOpen {
		t.Fatalf("expected help overlay open")
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = updated.(model)
	if m.selected != 0 {
		t.Fatalf("expected selection unchanged while help open, got %d", m.selected)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	m = updated.(model)
	if m.helpOpen {
		t.Fatalf("expected help overlay closed")
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = updated.(model)
	if m.selected != 1 {
		t.Fatalf("expected selected 1 after down, got %d", m.selected)
	}
}

func TestProfilesFilterReducesVisibleAndEnterExitsFilterMode(t *testing.T) {
	profiles := []profile.ProfileInfo{
		{Name: "alpha"},
		{Name: "beta"},
		{Name: "gamma"},
	}
	m := newModelWithActions("/config", false, profiles, "", false, stubActions())

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	m = updated.(model)
	if !m.profileFilterMode {
		t.Fatalf("expected filter mode on")
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ga")})
	m = updated.(model)
	if len(m.profilesVisible) != 1 || m.profilesVisible[0].Name != "gamma" {
		t.Fatalf("expected only gamma visible, got %v", m.profilesVisible)
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatalf("expected no command when leaving filter mode")
	}
	m = updated.(model)
	if m.profileFilterMode {
		t.Fatalf("expected filter mode off after enter")
	}
	if m.confirm.Open {
		t.Fatalf("expected no confirm when leaving filter mode")
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if !m.confirm.Open {
		t.Fatalf("expected confirm to open when pressing enter outside filter mode")
	}
}

func TestStatusClearsOnNextAction(t *testing.T) {
	profiles := []profile.ProfileInfo{{Name: "alpha"}}
	actions := stubActions()
	actions.activeProfile = func(_ string) (string, bool, error) { return "alpha", true, nil }
	m := newModelWithActions("/config", false, profiles, "", false, actions)

	updated, _ := m.Update(applyResultMsg{profile: "alpha"})
	m = updated.(model)
	if m.status.Kind != statusKindSuccess || m.status.Message == "" {
		t.Fatalf("expected success status, got kind=%v msg=%q", m.status.Kind, m.status.Message)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = updated.(model)
	if m.status.Message != "" {
		t.Fatalf("expected status cleared on next key, got %q", m.status.Message)
	}
}

func stubActions() modelActions {
	return modelActions{
		applyProfile: func(_, _ string) error {
			return nil
		},
		listProfileBackups: func(_, _ string) ([]string, error) {
			return nil, nil
		},
		activeProfile: func(_ string) (string, bool, error) {
			return "", false, nil
		},
		diffAgainstLastBackup: func(_, _ string) (string, bool, error) {
			return "", true, nil
		},
		diffBetweenProfiles: func(_, _, _ string) (string, error) {
			return "", nil
		},
		loadProfile: func(_ string) (*profile.RootConfig, error) {
			return &profile.RootConfig{}, nil
		},
		saveProfile: func(_ string, _ *profile.RootConfig) error {
			return nil
		},
		backupProfile: func(_, _ string) (string, error) {
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
