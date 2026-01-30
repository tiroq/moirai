package profile

import (
	"encoding/json"
	"testing"
)

func TestSetAgentModelPreservesExtra(t *testing.T) {
	cfg := &RootConfig{
		Agents: map[string]AgentConfig{
			"sisyphus": {
				Model: "old",
				Extra: map[string]json.RawMessage{
					"temperature": json.RawMessage("0.2"),
				},
			},
		},
	}
	changed, err := SetAgentModel(cfg, "sisyphus", "new")
	if err != nil {
		t.Fatalf("SetAgentModel: %v", err)
	}
	if !changed {
		t.Fatalf("expected change to be reported")
	}
	entry := cfg.Agents["sisyphus"]
	if entry.Model != "new" {
		t.Fatalf("expected model updated, got %q", entry.Model)
	}
	if _, ok := entry.Extra["temperature"]; !ok {
		t.Fatalf("expected extra fields preserved")
	}
}

func TestSetAgentModelCreatesAgent(t *testing.T) {
	cfg := &RootConfig{}
	changed, err := SetAgentModel(cfg, "oracle", "gpt-4o-mini")
	if err != nil {
		t.Fatalf("SetAgentModel: %v", err)
	}
	if !changed {
		t.Fatalf("expected change to be reported")
	}
	entry := cfg.Agents["oracle"]
	if entry.Model != "gpt-4o-mini" {
		t.Fatalf("expected model set, got %q", entry.Model)
	}
}
