package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestApplyAutofillAddsMissingAgents(t *testing.T) {
	cfg := &RootConfig{
		Agents: map[string]AgentConfig{
			"sisyphus": {Model: ""},
		},
	}
	preset, ok := PresetByName("openai")
	if !ok {
		t.Fatal("expected openai preset")
	}

	changed := ApplyAutofill(cfg, []string{"sisyphus", "oracle"}, preset)
	if !changed {
		t.Fatal("expected changes from autofill")
	}
	if cfg.Agents["sisyphus"].Model != preset.Model {
		t.Fatalf("expected sisyphus model %q, got %q", preset.Model, cfg.Agents["sisyphus"].Model)
	}
	if cfg.Agents["oracle"].Model != preset.Model {
		t.Fatalf("expected oracle model %q, got %q", preset.Model, cfg.Agents["oracle"].Model)
	}
}

func TestApplyAutofillKeepsExistingModel(t *testing.T) {
	cfg := &RootConfig{
		Agents: map[string]AgentConfig{
			"sisyphus": {Model: "custom-model"},
		},
	}
	preset, ok := PresetByName("openai")
	if !ok {
		t.Fatal("expected openai preset")
	}

	changed := ApplyAutofill(cfg, []string{"sisyphus"}, preset)
	if changed {
		t.Fatal("expected no changes from autofill")
	}
	if cfg.Agents["sisyphus"].Model != "custom-model" {
		t.Fatalf("unexpected model change: %q", cfg.Agents["sisyphus"].Model)
	}
}

func TestSaveProfileAtomicPreservesUnknownFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "oh-my-opencode.json.alpha")
	original := []byte(`{
  "$schema": "schema",
  "agents": {
    "sisyphus": {
      "model": "",
      "custom": "value"
    },
    "oracle": {
      "model": "keep-me"
    }
  },
  "extra": {
    "note": "preserve"
  }
}`)
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	cfg, err := LoadProfile(path)
	if err != nil {
		t.Fatalf("LoadProfile: %v", err)
	}

	preset, ok := PresetByName("openai")
	if !ok {
		t.Fatal("expected openai preset")
	}
	if !ApplyAutofill(cfg, []string{"sisyphus", "oracle"}, preset) {
		t.Fatal("expected changes from autofill")
	}
	if err := SaveProfileAtomic(path, cfg); err != nil {
		t.Fatalf("SaveProfileAtomic: %v", err)
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read updated profile: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(updated, &parsed); err != nil {
		t.Fatalf("parse updated JSON: %v", err)
	}

	if _, ok := parsed["extra"]; !ok {
		t.Fatal("expected extra field preserved")
	}
	agents, ok := parsed["agents"].(map[string]any)
	if !ok {
		t.Fatal("expected agents map preserved")
	}
	sisyphus, ok := agents["sisyphus"].(map[string]any)
	if !ok {
		t.Fatal("expected sisyphus agent preserved")
	}
	if sisyphus["custom"] != "value" {
		t.Fatalf("expected custom field preserved, got %v", sisyphus["custom"])
	}
	if sisyphus["model"] != preset.Model {
		t.Fatalf("expected model %q, got %v", preset.Model, sisyphus["model"])
	}
	oracle, ok := agents["oracle"].(map[string]any)
	if !ok {
		t.Fatal("expected oracle agent preserved")
	}
	if oracle["model"] != "keep-me" {
		t.Fatalf("expected oracle model unchanged, got %v", oracle["model"])
	}
}
