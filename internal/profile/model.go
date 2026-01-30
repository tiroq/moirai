package profile

import "encoding/json"

// RootConfig is the minimal configuration structure used by doctor.
type RootConfig struct {
	Schema string                 `json:"$schema,omitempty"`
	Agents map[string]AgentConfig `json:"agents,omitempty"`
	Extra  map[string]json.RawMessage
}

// AgentConfig is the minimal agent configuration used by doctor.
type AgentConfig struct {
	Model string `json:"model,omitempty"`
	Extra map[string]json.RawMessage
}

// UnmarshalJSON preserves unknown fields alongside known fields.
func (cfg *RootConfig) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if value, ok := raw["$schema"]; ok {
		if err := json.Unmarshal(value, &cfg.Schema); err != nil {
			return err
		}
		delete(raw, "$schema")
	}
	if value, ok := raw["agents"]; ok {
		var agents map[string]AgentConfig
		if err := json.Unmarshal(value, &agents); err != nil {
			return err
		}
		cfg.Agents = agents
		delete(raw, "agents")
	}

	if len(raw) == 0 {
		cfg.Extra = nil
		return nil
	}
	cfg.Extra = raw
	return nil
}

// MarshalJSON preserves unknown fields alongside known fields.
func (cfg RootConfig) MarshalJSON() ([]byte, error) {
	merged := make(map[string]json.RawMessage, len(cfg.Extra)+2)
	for key, value := range cfg.Extra {
		merged[key] = value
	}

	if cfg.Schema != "" {
		encoded, err := json.Marshal(cfg.Schema)
		if err != nil {
			return nil, err
		}
		merged["$schema"] = encoded
	}
	if cfg.Agents != nil {
		encoded, err := json.Marshal(cfg.Agents)
		if err != nil {
			return nil, err
		}
		merged["agents"] = encoded
	}

	return json.Marshal(merged)
}

// UnmarshalJSON preserves unknown fields alongside known fields.
func (cfg *AgentConfig) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if value, ok := raw["model"]; ok {
		if err := json.Unmarshal(value, &cfg.Model); err != nil {
			return err
		}
		delete(raw, "model")
	}

	if len(raw) == 0 {
		cfg.Extra = nil
		return nil
	}
	cfg.Extra = raw
	return nil
}

// MarshalJSON preserves unknown fields alongside known fields.
func (cfg AgentConfig) MarshalJSON() ([]byte, error) {
	merged := make(map[string]json.RawMessage, len(cfg.Extra)+1)
	for key, value := range cfg.Extra {
		merged[key] = value
	}

	if cfg.Model != "" {
		encoded, err := json.Marshal(cfg.Model)
		if err != nil {
			return nil, err
		}
		merged["model"] = encoded
	}

	return json.Marshal(merged)
}
