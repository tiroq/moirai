package tui

import (
	"fmt"
	"strings"
)

func (m model) viewBackups() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Backups: %s\n\n", m.backupsProfile)
	if m.backupsMessage != "" {
		fmt.Fprintf(&b, "  %s\n", m.backupsMessage)
	} else {
		for _, name := range m.backups {
			fmt.Fprintf(&b, "  %s\n", name)
		}
	}
	return b.String()
}
