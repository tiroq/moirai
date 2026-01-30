package tui

import (
	"moirai/internal/app"
	"moirai/internal/link"
	"moirai/internal/profile"

	tea "github.com/charmbracelet/bubbletea"
)

var programRunner = func(m tea.Model) error {
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}

// Run starts the TUI.
func Run(config app.AppConfig) error {
	m, err := loadModel(config)
	if err != nil {
		return err
	}
	return programRunner(m)
}

func loadModel(config app.AppConfig) (model, error) {
	profiles, err := profile.DiscoverProfiles(config.ConfigDir)
	if err != nil {
		return model{}, err
	}
	activeName, ok, err := link.ActiveProfile(config.ConfigDir)
	if err != nil {
		return model{}, err
	}
	return newModel(config.ConfigDir, profiles, activeName, ok), nil
}
