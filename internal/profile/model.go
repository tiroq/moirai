package profile

// RootConfig is the minimal configuration structure used by doctor.
type RootConfig struct {
	Schema string                 `json:"$schema,omitempty"`
	Agents map[string]AgentConfig `json:"agents,omitempty"`
}

// AgentConfig is the minimal agent configuration used by doctor.
type AgentConfig struct {
	Model string `json:"model,omitempty"`
}
