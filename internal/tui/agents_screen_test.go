package tui

import (
	"strings"
	"testing"

	"moirai/internal/profile"

	tea "github.com/charmbracelet/bubbletea"
)

func TestAgentsScreenLoadsKnownAndCustomAgents(t *testing.T) {
	profiles := []profile.ProfileInfo{
		{Name: "alpha", Path: "/config/oh-my-opencode.json.alpha"},
	}

	cfg := &profile.RootConfig{
		Agents: map[string]profile.AgentConfig{
			"oracle":  {Model: "gpt-4o-mini"},
			"custom":  {Model: "local-model"},
			"metis":   {Model: ""},
			"another": {Model: "x"},
		},
	}

	actions := stubActions()
	actions.loadProfile = func(path string) (*profile.RootConfig, error) {
		if path != profiles[0].Path {
			t.Fatalf("expected load path %q, got %q", profiles[0].Path, path)
		}
		return cfg, nil
	}

	m := newModelWithActions("/config", false, profiles, "", false, actions)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	if cmd == nil {
		t.Fatalf("expected load command")
	}
	msg := cmd()
	updated, _ = updated.(model).Update(msg)
	m = updated.(model)

	if m.screen != screenAgents {
		t.Fatalf("expected agents screen, got %v", m.screen)
	}

	known := profile.KnownAgents()
	if len(m.agentsEntries) != len(known)+2 {
		t.Fatalf("expected %d entries, got %d", len(known)+2, len(m.agentsEntries))
	}

	// Known agents come first, in KnownAgents order.
	if m.agentsEntries[0].Name != known[0] {
		t.Fatalf("expected first entry %q, got %q", known[0], m.agentsEntries[0].Name)
	}
	// Custom agents are appended and sorted.
	if m.agentsEntries[len(m.agentsEntries)-2].Name != "another" || m.agentsEntries[len(m.agentsEntries)-1].Name != "custom" {
		t.Fatalf("expected custom agents sorted at end, got %q and %q", m.agentsEntries[len(m.agentsEntries)-2].Name, m.agentsEntries[len(m.agentsEntries)-1].Name)
	}
}

func TestAgentsScreenFlagsMissingModels(t *testing.T) {
	cfg := &profile.RootConfig{
		Agents: map[string]profile.AgentConfig{
			"oracle": {Model: "gpt-4o-mini"},
		},
	}
	entries := collectAgentEntries(cfg, profile.KnownAgents())

	var prometheus agentEntry
	foundPrometheus := false
	var oracle agentEntry
	foundOracle := false
	for _, entry := range entries {
		switch entry.Name {
		case "prometheus":
			prometheus = entry
			foundPrometheus = true
		case "oracle":
			oracle = entry
			foundOracle = true
		}
	}
	if !foundPrometheus || !foundOracle {
		t.Fatalf("expected to find prometheus and oracle entries")
	}
	if !prometheus.Missing {
		t.Fatalf("expected prometheus to be missing")
	}
	if oracle.Missing {
		t.Fatalf("expected oracle to not be missing")
	}
}

func TestModelPickerSelectionUpdatesConfigAndMarksDirty(t *testing.T) {
	profiles := []profile.ProfileInfo{
		{Name: "alpha", Path: "/config/oh-my-opencode.json.alpha"},
	}

	cfg := &profile.RootConfig{
		Agents: map[string]profile.AgentConfig{
			"sisyphus": {Model: ""},
		},
	}

	var backupCalls, saveCalls int
	actions := stubActions()
	actions.loadProfile = func(_ string) (*profile.RootConfig, error) {
		return cfg, nil
	}
	actions.backupProfile = func(_, _ string) (string, error) { backupCalls++; return "/config/backup", nil }
	actions.saveProfile = func(_ string, _ *profile.RootConfig) error { saveCalls++; return nil }
	actions.loadModels = func() []string {
		return []string{"gpt-4o-mini", "gpt-4o"}
	}

	m := newModelWithActions("/config", false, profiles, "", false, actions)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	msg := cmd()
	updated, _ = updated.(model).Update(msg)
	m = updated.(model)

	if m.screen != screenAgents {
		t.Fatalf("expected agents screen, got %v", m.screen)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if m.screen != screenModels {
		t.Fatalf("expected models screen, got %v", m.screen)
	}
	if m.modelTargetAgent != "sisyphus" {
		t.Fatalf("expected target agent sisyphus, got %q", m.modelTargetAgent)
	}

	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("expected save cmd after model select")
	}
	msg = cmd()
	updated, _ = updated.(model).Update(msg)
	m = updated.(model)
	if m.screen != screenAgents {
		t.Fatalf("expected return to agents screen, got %v", m.screen)
	}
	if m.agentsDirty {
		t.Fatalf("expected dirty cleared after auto-save")
	}
	if got := cfg.Agents["sisyphus"].Model; got != "gpt-4o-mini" {
		t.Fatalf("expected model updated, got %q", got)
	}
	if backupCalls != 1 || saveCalls != 1 {
		t.Fatalf("expected backup+save once, got backup=%d save=%d", backupCalls, saveCalls)
	}
}

func TestAgentsSaveBacksUpWritesAndClearsDirty(t *testing.T) {
	profiles := []profile.ProfileInfo{
		{Name: "alpha", Path: "/config/oh-my-opencode.json.alpha"},
	}
	cfg := &profile.RootConfig{}

	var backupCalls, saveCalls int
	actions := stubActions()
	actions.loadProfile = func(_ string) (*profile.RootConfig, error) { return cfg, nil }
	actions.backupProfile = func(dir, profileName string) (string, error) {
		backupCalls++
		if dir != "/config" || profileName != "alpha" {
			t.Fatalf("unexpected backup args: %q %q", dir, profileName)
		}
		return "/config/backup", nil
	}
	actions.saveProfile = func(path string, gotCfg *profile.RootConfig) error {
		saveCalls++
		if path != profiles[0].Path {
			t.Fatalf("unexpected save path %q", path)
		}
		if gotCfg != cfg {
			t.Fatalf("expected cfg pointer to be passed through")
		}
		return nil
	}
	actions.loadModels = func() []string { return []string{"gpt-4o-mini"} }

	m := newModelWithActions("/config", false, profiles, "", false, actions)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	msg := cmd()
	updated, _ = updated.(model).Update(msg)
	m = updated.(model)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if !m.agentsDirty {
		t.Fatalf("expected dirty before save")
	}

	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	if cmd != nil {
		t.Fatalf("expected no command before confirmation")
	}
	m = updated.(model)
	if !m.confirm.Open {
		t.Fatalf("expected confirm to open for save")
	}

	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if cmd == nil {
		t.Fatalf("expected save command after confirmation")
	}
	msg = cmd()
	updated, _ = updated.(model).Update(msg)
	m = updated.(model)

	if backupCalls != 1 {
		t.Fatalf("expected 1 backup call, got %d", backupCalls)
	}
	if saveCalls != 1 {
		t.Fatalf("expected 1 save call, got %d", saveCalls)
	}
	if m.agentsDirty {
		t.Fatalf("expected dirty cleared after save")
	}
	if m.status.Message != "Saved" || m.status.Kind != statusKindSuccess {
		t.Fatalf("expected saved status, got kind=%v msg=%q", m.status.Kind, m.status.Message)
	}
}

func TestAgentsAutofillDisabledShowsMessageAndDoesNotWrite(t *testing.T) {
	profiles := []profile.ProfileInfo{
		{Name: "alpha", Path: "/config/oh-my-opencode.json.alpha"},
	}
	cfg := &profile.RootConfig{}

	var backupCalls, saveCalls, autofillCalls int
	actions := stubActions()
	actions.loadProfile = func(_ string) (*profile.RootConfig, error) { return cfg, nil }
	actions.backupProfile = func(_, _ string) (string, error) { backupCalls++; return "", nil }
	actions.saveProfile = func(_ string, _ *profile.RootConfig) error { saveCalls++; return nil }
	actions.applyAutofill = func(_ *profile.RootConfig, _ []string, _ profile.Preset) bool {
		autofillCalls++
		return false
	}

	m := newModelWithActions("/config", false, profiles, "", false, actions)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	msg := cmd()
	updated, _ = updated.(model).Update(msg)
	m = updated.(model)

	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m = updated.(model)
	if cmd != nil {
		t.Fatalf("expected no command when autofill is disabled")
	}
	if m.confirm.Open {
		t.Fatalf("expected no confirm when autofill is disabled")
	}
	if m.status.Message != "Autofill disabled. Run with --enable-autofill." || m.status.Kind != statusKindError {
		t.Fatalf("unexpected status kind=%v msg=%q", m.status.Kind, m.status.Message)
	}
	if backupCalls != 0 || saveCalls != 0 || autofillCalls != 0 {
		t.Fatalf("expected no writes when disabled, got backup=%d save=%d autofill=%d", backupCalls, saveCalls, autofillCalls)
	}
}

func TestAgentsAutofillEnabledFillsAndSaves(t *testing.T) {
	profiles := []profile.ProfileInfo{
		{Name: "alpha", Path: "/config/oh-my-opencode.json.alpha"},
	}
	cfg := &profile.RootConfig{
		Agents: map[string]profile.AgentConfig{
			"sisyphus": {Model: ""},
		},
	}

	var backupCalls, saveCalls, autofillCalls int
	actions := stubActions()
	actions.loadProfile = func(_ string) (*profile.RootConfig, error) { return cfg, nil }
	actions.backupProfile = func(_, _ string) (string, error) { backupCalls++; return "", nil }
	actions.saveProfile = func(_ string, _ *profile.RootConfig) error { saveCalls++; return nil }
	actions.applyAutofill = func(cfg *profile.RootConfig, knownAgents []string, preset profile.Preset) bool {
		autofillCalls++
		return profile.ApplyAutofill(cfg, knownAgents, preset)
	}

	m := newModelWithActions("/config", true, profiles, "", false, actions)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	msg := cmd()
	updated, _ = updated.(model).Update(msg)
	m = updated.(model)

	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	if cmd != nil {
		t.Fatalf("expected no command before confirmation")
	}
	m = updated.(model)
	if !m.confirm.Open {
		t.Fatalf("expected confirm to open for autofill")
	}

	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if cmd == nil {
		t.Fatalf("expected autofill command after confirmation")
	}
	msg = cmd()
	updated, _ = updated.(model).Update(msg)
	m = updated.(model)

	if autofillCalls != 1 {
		t.Fatalf("expected 1 autofill call, got %d", autofillCalls)
	}
	if backupCalls != 1 {
		t.Fatalf("expected 1 backup call, got %d", backupCalls)
	}
	if saveCalls != 1 {
		t.Fatalf("expected 1 save call, got %d", saveCalls)
	}
	if m.agentsDirty {
		t.Fatalf("expected dirty cleared after autofill+save")
	}
	if !strings.HasPrefix(m.status.Message, "Autofilled ") || m.status.Kind != statusKindSuccess {
		t.Fatalf("expected autofill status, got kind=%v msg=%q", m.status.Kind, m.status.Message)
	}
	if got := cfg.Agents["sisyphus"].Model; got == "" {
		t.Fatalf("expected sisyphus to be filled")
	}
}
