package tui

import (
	"fmt"
	"strings"
)

func (m model) viewDiff() string {
	var b strings.Builder
	var title string
	switch m.diffMode {
	case diffModeActiveProfile:
		if m.diffAgainst != "" {
			title = fmt.Sprintf("Diff: %s vs active (%s)", m.diffProfile, m.diffAgainst)
		} else {
			title = fmt.Sprintf("Diff: %s vs active", m.diffProfile)
		}
	default:
		title = fmt.Sprintf("Diff: %s vs last-backup", m.diffProfile)
	}
	b.WriteString(title)
	b.WriteString("\n\n")

	if m.diffMessage != "" {
		b.WriteString(m.diffMessage)
		b.WriteString("\n")
	} else {
		view := m.viewport.View()
		b.WriteString(view)
		if !strings.HasSuffix(view, "\n") {
			b.WriteString("\n")
		}
	}

	return b.String()
}
