package app

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type AppConfig struct {
	ConfigDir      string
	EnableAutofill bool
}

type fileConfig struct {
	EnableAutofill *bool `json:"enableAutofill"`
}

func LoadConfig(configDir string, enableAutofillOverride *bool) (AppConfig, error) {
	config := AppConfig{
		ConfigDir: filepath.Clean(configDir),
	}
	configPath := filepath.Join(config.ConfigDir, "moirai.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return AppConfig{}, err
		}
	} else {
		var fileCfg fileConfig
		if err := json.Unmarshal(data, &fileCfg); err != nil {
			return AppConfig{}, err
		}
		if fileCfg.EnableAutofill != nil {
			config.EnableAutofill = *fileCfg.EnableAutofill
		}
	}

	if enableAutofillOverride != nil {
		config.EnableAutofill = *enableAutofillOverride
	}

	return config, nil
}
