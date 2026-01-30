package profile

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadProfile reads and parses a profile from path.
func LoadProfile(path string) (*RootConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg RootConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse profile: %w", err)
	}

	return &cfg, nil
}
