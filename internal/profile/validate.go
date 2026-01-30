package profile

import (
	"sort"
	"strings"
)

// MissingAgents returns known agents that are absent or missing model values.
func MissingAgents(cfg *RootConfig, knownAgents []string) []string {
	if cfg == nil {
		missing := append([]string(nil), knownAgents...)
		sort.Strings(missing)
		return missing
	}

	missing := make([]string, 0)
	for _, agent := range knownAgents {
		entry, ok := cfg.Agents[agent]
		if !ok || strings.TrimSpace(entry.Model) == "" {
			missing = append(missing, agent)
		}
	}

	sort.Strings(missing)
	return missing
}
