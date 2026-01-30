package tui

import (
	"fmt"
	"strings"
)

func (m model) viewProfiles() string {
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
	b.WriteString(hintStyle.Render("enter apply · e agents · b backups · d diff · q/esc/ctrl+c quit · j/k/arrows move"))
	b.WriteString("\n")

	return b.String()
}
