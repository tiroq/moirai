package tui

import (
	"fmt"
	"strings"

	"moirai/internal/profile"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	configDir  string
	profiles   []profile.ProfileInfo
	activeName string
	hasActive  bool
	selected   int
	status     string
}

func newModel(configDir string, profiles []profile.ProfileInfo, activeName string, hasActive bool) model {
	selected := -1
	if len(profiles) > 0 {
		selected = 0
		if hasActive {
			for i, profileInfo := range profiles {
				if profileInfo.Name == activeName {
					selected = i
					break
				}
			}
		}
	}
	return model{
		configDir:  configDir,
		profiles:   profiles,
		activeName: activeName,
		hasActive:  hasActive,
		selected:   selected,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "j", "down":
			m.moveSelection(1)
		case "k", "up":
			m.moveSelection(-1)
		case "enter":
			m.status = "Actions will come in next phase."
		}
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder
	fmt.Fprintf(&b, "ConfigDir: %s\n", m.configDir)
	if m.hasActive {
		fmt.Fprintf(&b, "Active: %s\n", m.activeName)
	} else {
		b.WriteString("Active: (none)\n")
	}
	b.WriteString("\nProfiles:\n")
	if len(m.profiles) == 0 {
		b.WriteString("  (none)\n")
	} else {
		for i, profileInfo := range m.profiles {
			prefix := "  "
			isSelected := i == m.selected
			if isSelected {
				prefix = "> "
			}
			name := profileInfo.Name
			isActive := m.hasActive && profileInfo.Name == m.activeName
			if isActive {
				name = activeStyle.Render(name)
			}
			if isSelected {
				name = selectedStyle.Render(name)
			}
			fmt.Fprintf(&b, "%s%s\n", prefix, name)
		}
	}

	b.WriteString("\n")
	if m.status != "" {
		b.WriteString(hintStyle.Render(m.status))
		b.WriteString("\n")
	}
	b.WriteString(hintStyle.Render("q/esc/ctrl+c quit · j/k/arrows move · enter actions"))
	b.WriteString("\n")

	return lipgloss.NewStyle().Render(b.String())
}

func (m *model) moveSelection(delta int) {
	if len(m.profiles) == 0 {
		return
	}
	m.selected += delta
	if m.selected < 0 {
		m.selected = 0
	}
	if m.selected >= len(m.profiles) {
		m.selected = len(m.profiles) - 1
	}
}
