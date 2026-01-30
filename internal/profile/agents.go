package profile

import "fmt"

// KnownAgents returns the static list of supported agents.
func KnownAgents() []string {
	return []string{
		"sisyphus",
		"prometheus",
		"oracle",
		"librarian",
		"explore",
		"multimodal-looker",
		"metis",
		"momus",
		"atlas",
	}
}

// SetAgentModel updates the model value for an agent while preserving extra fields.
func SetAgentModel(cfg *RootConfig, agent, model string) (bool, error) {
	if cfg == nil {
		return false, fmt.Errorf("config is required")
	}
	if agent == "" {
		return false, fmt.Errorf("agent name is required")
	}
	if cfg.Agents == nil {
		cfg.Agents = make(map[string]AgentConfig)
	}
	entry := cfg.Agents[agent]
	if entry.Model == model {
		return false, nil
	}
	entry.Model = model
	cfg.Agents[agent] = entry
	return true, nil
}
