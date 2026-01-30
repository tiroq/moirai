package profile

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
