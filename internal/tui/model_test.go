package tui

import (
	"testing"

	"moirai/internal/profile"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModelSelectsActive(t *testing.T) {
	profiles := []profile.ProfileInfo{
		{Name: "alpha"},
		{Name: "beta"},
	}
	m := newModel("/config", profiles, "beta", true)
	if m.selected != 1 {
		t.Fatalf("expected selected 1, got %d", m.selected)
	}
}

func TestModelUpdateMovementAndHint(t *testing.T) {
	profiles := []profile.ProfileInfo{
		{Name: "alpha"},
		{Name: "beta"},
	}
	m := newModel("/config", profiles, "", false)

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

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if m.status == "" {
		t.Fatalf("expected status hint after enter")
	}
}

func TestModelUpdateQuit(t *testing.T) {
	m := newModel("/config", nil, "", false)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatalf("expected quit command")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("expected quit message, got %T", msg)
	}
}
