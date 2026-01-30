package tui

import "github.com/charmbracelet/lipgloss"

var (
	selectedStyle = lipgloss.NewStyle().Bold(true)
	activeStyle   = lipgloss.NewStyle().Underline(true)
	hintStyle     = lipgloss.NewStyle().Faint(true)
)
