package profile

import "strings"

// Preset describes autofill defaults for agents.
type Preset struct {
	Name  string
	Model string
}

// PresetByName resolves a preset by name.
func PresetByName(name string) (Preset, bool) {
	switch name {
	case "openai":
		return Preset{Name: "openai", Model: "gpt-4o-mini"}, true
	default:
		return Preset{}, false
	}
}

// ApplyAutofill updates missing agents or empty models and reports changes.
func ApplyAutofill(cfg *RootConfig, knownAgents []string, preset Preset) bool {
	if cfg == nil {
		return false
	}
	if cfg.Agents == nil {
		cfg.Agents = make(map[string]AgentConfig)
	}

	changed := false
	for _, agent := range knownAgents {
		entry, ok := cfg.Agents[agent]
		if !ok {
			cfg.Agents[agent] = AgentConfig{Model: preset.Model}
			changed = true
			continue
		}
		if strings.TrimSpace(entry.Model) == "" {
			entry.Model = preset.Model
			cfg.Agents[agent] = entry
			changed = true
		}
	}
	return changed
}
