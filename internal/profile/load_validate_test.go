package profile

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadProfileParsesMinimalJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "oh-my-opencode.json.alpha")

	data := []byte(`{"agents":{"sisyphus":{"model":"gpt-4"}}}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	cfg, err := LoadProfile(path)
	if err != nil {
		t.Fatalf("LoadProfile: %v", err)
	}

	if cfg.Agents["sisyphus"].Model != "gpt-4" {
		t.Fatalf("unexpected model: %q", cfg.Agents["sisyphus"].Model)
	}
}

func TestLoadProfileInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "oh-my-opencode.json.alpha")

	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	if _, err := LoadProfile(path); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestMissingAgentsMissingKey(t *testing.T) {
	cfg := &RootConfig{
		Agents: map[string]AgentConfig{
			"alpha": {Model: "model-a"},
		},
	}
	missing := MissingAgents(cfg, []string{"alpha", "beta"})

	expect := []string{"beta"}
	if !reflect.DeepEqual(missing, expect) {
		t.Fatalf("expected %v, got %v", expect, missing)
	}
}

func TestMissingAgentsEmptyModel(t *testing.T) {
	cfg := &RootConfig{
		Agents: map[string]AgentConfig{
			"alpha": {Model: ""},
		},
	}
	missing := MissingAgents(cfg, []string{"alpha"})

	expect := []string{"alpha"}
	if !reflect.DeepEqual(missing, expect) {
		t.Fatalf("expected %v, got %v", expect, missing)
	}
}

func TestMissingAgentsAllPresent(t *testing.T) {
	cfg := &RootConfig{
		Agents: map[string]AgentConfig{
			"alpha": {Model: "model-a"},
			"beta":  {Model: "model-b"},
		},
	}
	missing := MissingAgents(cfg, []string{"alpha", "beta"})

	if len(missing) != 0 {
		t.Fatalf("expected no missing agents, got %v", missing)
	}
}

func TestMissingAgentsDeterministicOrdering(t *testing.T) {
	cfg := &RootConfig{
		Agents: map[string]AgentConfig{},
	}
	missing := MissingAgents(cfg, []string{"gamma", "alpha", "beta"})

	expect := []string{"alpha", "beta", "gamma"}
	if !reflect.DeepEqual(missing, expect) {
		t.Fatalf("expected %v, got %v", expect, missing)
	}
}

func TestMissingAgentsNilConfig(t *testing.T) {
	missing := MissingAgents(nil, []string{"beta", "alpha"})

	expect := []string{"alpha", "beta"}
	if !reflect.DeepEqual(missing, expect) {
		t.Fatalf("expected %v, got %v", expect, missing)
	}
}

func TestKnownAgents(t *testing.T) {
	agents := KnownAgents()

	if len(agents) != 9 {
		t.Fatalf("expected 9 agents, got %d", len(agents))
	}
	if agents[0] != "sisyphus" || agents[len(agents)-1] != "atlas" {
		t.Fatalf("unexpected agents list: %v", agents)
	}
}
