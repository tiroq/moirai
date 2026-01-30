package tui

import (
	"fmt"
	"strings"

	"moirai/internal/profile"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) viewProfiles() string {
	var b strings.Builder
	fmt.Fprintf(&b, "ConfigDir: %s\n", m.configDir)
	if m.hasActive {
		fmt.Fprintf(&b, "Active: %s\n", m.activeName)
	} else {
		b.WriteString("Active: (none)\n")
	}
	if m.profileFilterMode || m.profileFilter != "" {
		fmt.Fprintf(&b, "Filter: %s\n", m.profileFilter)
	}
	b.WriteString("\nProfiles:\n")
	if len(m.profilesVisible) == 0 {
		b.WriteString("  (none)\n")
	} else {
		pageSize := len(m.profilesVisible)
		if m.height > 0 {
			// Keep the overall render within the terminal height so we don't push
			// the title line into scrollback.
			//
			// Header lines here:
			//   ConfigDir, Active, optional Filter, blank, "Profiles:"
			headerLines := 4
			if m.profileFilterMode || m.profileFilter != "" {
				headerLines = 5
			}
			// Reserve title art + blank separator + status bar.
			pageSize = m.height - (titleArtHeight()+2) - headerLines
			if pageSize < 1 {
				pageSize = 1
			}
		}
		start, end := modelWindow(len(m.profilesVisible), m.selected, pageSize)
		for i := start; i < end; i++ {
			profileInfo := m.profilesVisible[i]
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

	return b.String()
}

func (m model) handleProfilesKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if m.profileFilterMode {
		if isCtrlU(msg) {
			m.profileFilter = ""
			m.updateProfilesFilter()
			return m, nil
		}
		switch msg.Type {
		case tea.KeyEsc:
			m.profileFilterMode = false
			return m, nil
		case tea.KeyEnter:
			m.profileFilterMode = false
			return m, nil
		case tea.KeyBackspace, tea.KeyDelete:
			if m.profileFilter != "" {
				runes := []rune(m.profileFilter)
				if len(runes) > 0 {
					m.profileFilter = string(runes[:len(runes)-1])
					m.updateProfilesFilter()
				}
			}
			return m, nil
		case tea.KeyRunes:
			m.profileFilter += string(msg.Runes)
			m.updateProfilesFilter()
			return m, nil
		}

		switch key {
		case "j", "down":
			m.moveSelection(1)
		case "k", "up":
			m.moveSelection(-1)
		}
		return m, nil
	}

	switch key {
	case "esc":
		return m, tea.Quit
	case "/":
		m.profileFilterMode = true
		return m, nil
	case "j", "down":
		m.moveSelection(1)
	case "k", "up":
		m.moveSelection(-1)
	case "enter":
		return m.confirmApplySelected()
	case "e":
		return m.openAgents()
	case "b":
		return m.openBackups()
	case "d":
		return m.openDiff(diffModeLastBackup)
	}
	return m, nil
}

func (m *model) updateProfilesFilter() {
	selectedName := ""
	if m.selected >= 0 && m.selected < len(m.profilesVisible) {
		selectedName = m.profilesVisible[m.selected].Name
	}
	if m.profileFilter == "" {
		m.profilesVisible = append([]profile.ProfileInfo(nil), m.profiles...)
	} else {
		lowered := strings.ToLower(m.profileFilter)
		visible := make([]profile.ProfileInfo, 0, len(m.profiles))
		for _, info := range m.profiles {
			if strings.Contains(strings.ToLower(info.Name), lowered) {
				visible = append(visible, info)
			}
		}
		m.profilesVisible = visible
	}

	if len(m.profilesVisible) == 0 {
		m.selected = -1
		return
	}
	if selectedName == "" {
		if m.selected < 0 || m.selected >= len(m.profilesVisible) {
			m.selected = 0
		}
		return
	}
	for i, info := range m.profilesVisible {
		if info.Name == selectedName {
			m.selected = i
			return
		}
	}
	m.selected = 0
}
