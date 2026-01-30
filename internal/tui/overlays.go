package tui

import (
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
)

type statusKind int

const (
	statusKindNone statusKind = iota
	statusKindInfo
	statusKindSuccess
	statusKindError
)

type statusBarState struct {
	Kind           statusKind
	Message        string
	ClearOnNextKey bool
}

func (m *model) setStatus(kind statusKind, msg string) {
	m.status = statusBarState{
		Kind:           kind,
		Message:        msg,
		ClearOnNextKey: msg != "",
	}
}

func (m *model) clearStatusForKey(key string) {
	if !m.status.ClearOnNextKey || m.status.Message == "" {
		return
	}
	if key == "?" {
		return
	}
	m.status = statusBarState{}
}

func (m model) renderStatusBar() string {
	hints := m.statusHints()
	left := m.status.Message
	right := ""
	if hints != "" {
		right = hintStyle.Render(hints)
	}
	if m.width <= 0 {
		if left == "" {
			return right
		}
		if right == "" {
			return statusTextStyle(m.status.Kind).Render(left)
		}
		return statusTextStyle(m.status.Kind).Render(left) + " " + right
	}

	// Default/idle state: no status message. Show hints centered to make them
	// easier to discover.
	if left == "" {
		hintWidth := runeLen(hints)
		if hintWidth <= 0 {
			return ""
		}
		padding := (m.width - hintWidth) / 2
		if padding < 0 {
			padding = 0
		}
		return strings.Repeat(" ", padding) + right
	}

	rightWidth := runeLen(hints)
	maxLeft := m.width - rightWidth - 1
	if maxLeft < 0 {
		maxLeft = 0
	}
	if left != "" {
		left = truncateRunes(left, maxLeft)
		left = statusTextStyle(m.status.Kind).Render(left)
	}
	if hints == "" {
		return left
	}
	leftWidth := runeLen(stripANSI(left))
	padding := m.width - leftWidth - rightWidth
	if padding < 1 {
		padding = 1
	}
	return left + strings.Repeat(" ", padding) + right
}

func statusTextStyle(kind statusKind) textStyle {
	switch kind {
	case statusKindSuccess:
		return selectedStyle
	case statusKindError:
		return missingStyle
	case statusKindInfo:
		return hintStyle
	default:
		return textStyle{}
	}
}

type confirmState struct {
	Open   bool
	Prompt string
	OnYes  func(model) (tea.Model, tea.Cmd)
}

func (m *model) openConfirm(prompt string, onYes func(model) (tea.Model, tea.Cmd)) {
	m.confirm = confirmState{
		Open:   true,
		Prompt: prompt,
		OnYes:  onYes,
	}
}

func (m model) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		onYes := m.confirm.OnYes
		m.confirm = confirmState{}
		if onYes == nil {
			return m, nil
		}
		return onYes(m)
	case "n", "N", "esc":
		m.confirm = confirmState{}
		return m, nil
	}
	return m, nil
}

func (m model) renderConfirmModal() string {
	lines := []string{
		m.confirm.Prompt,
		"",
		"y yes · n no · esc cancel",
	}
	return renderBox("Confirm", lines, m.width)
}

func (m model) renderHelpModal() string {
	lines := m.helpLines()
	title := "Help"
	switch m.screen {
	case screenProfiles:
		title = "Help: Profiles"
	case screenBackups:
		title = "Help: Backups"
	case screenDiff:
		title = "Help: Diff"
	case screenAgents:
		title = "Help: Agents"
	case screenModels:
		title = "Help: Model Picker"
	}
	return renderBox(title, lines, m.width)
}

func (m model) helpLines() []string {
	switch m.screen {
	case screenBackups:
		return []string{
			"esc back",
			"q quit",
			"? help",
		}
	case screenDiff:
		return []string{
			"j/k, arrows scroll",
			"pgup/pgdown page scroll",
			"a diff vs active profile",
			"d diff vs last-backup",
			"esc back",
			"q quit",
			"? help",
		}
	case screenAgents:
		return []string{
			"j/k, arrows move selection",
			"enter pick model",
			"s save (confirm when dirty)",
			"r reload",
			"a autofill (confirm)",
			"esc back",
			"q quit",
			"? help",
		}
	case screenModels:
		return []string{
			"type to search",
			"ctrl+u clear search",
			"j/k, arrows move selection",
			"enter select model",
			"R refresh models",
			"esc cancel",
			"q quit",
			"? help",
		}
	default:
		lines := []string{
			"j/k, arrows move selection",
			"/ filter profiles",
			"ctrl+u clear filter",
			"enter apply profile (confirm)",
			"e edit agents",
			"b view backups",
			"d view diff",
			"q quit",
			"? help",
		}
		if m.profileFilterMode {
			lines = append([]string{"enter exit filter mode"}, lines...)
		}
		return lines
	}
}

func (m model) statusHints() string {
	if m.confirm.Open {
		return "y yes · n no · esc cancel · q quit"
	}
	if m.helpOpen {
		return "esc close · q quit"
	}
	switch m.screen {
	case screenBackups:
		return "esc back · ? help · q quit"
	case screenDiff:
		return "j/k scroll · pgup/pgdown page · a/d diff mode · esc back · ? help · q quit"
	case screenAgents:
		return "j/k move · enter models · s save · r reload · a autofill · esc back · ? help · q quit"
	case screenModels:
		return "type search · ctrl+u clear · j/k move · pgup/pgdown page · enter select · R refresh · esc cancel · ? help · q quit"
	default:
		if m.profileFilterMode {
			return "type filter · ctrl+u clear · enter done · esc cancel · j/k move · ? help · q quit"
		}
		return "j/k move · / filter · enter apply · e agents · b backups · d diff · ? help · q quit"
	}
}

func renderBox(title string, lines []string, width int) string {
	if width <= 0 {
		width = 80
	}
	maxContent := runeLen(title)
	for _, line := range lines {
		if w := runeLen(line); w > maxContent {
			maxContent = w
		}
	}
	maxBox := width - 4
	if maxBox < 10 {
		maxBox = 10
	}
	if maxContent > maxBox {
		maxContent = maxBox
	}

	title = truncateRunes(title, maxContent)
	truncated := make([]string, 0, len(lines))
	for _, line := range lines {
		truncated = append(truncated, truncateRunes(line, maxContent))
	}

	var b strings.Builder
	borderTop := "┌" + strings.Repeat("─", maxContent+2) + "┐"
	borderBot := "└" + strings.Repeat("─", maxContent+2) + "┘"
	b.WriteString(borderTop)
	b.WriteString("\n")
	header := "│ " + padRight(title, maxContent) + " │"
	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString("├" + strings.Repeat("─", maxContent+2) + "┤")
	b.WriteString("\n")
	for _, line := range truncated {
		b.WriteString("│ " + padRight(line, maxContent) + " │")
		b.WriteString("\n")
	}
	b.WriteString(borderBot)
	return b.String()
}

func padRight(s string, width int) string {
	pad := width - runeLen(s)
	if pad <= 0 {
		return s
	}
	return s + strings.Repeat(" ", pad)
}

func truncateRunes(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	if runeLen(s) <= maxRunes {
		return s
	}
	if maxRunes == 1 {
		return "…"
	}
	runes := []rune(s)
	return string(runes[:maxRunes-1]) + "…"
}

func runeLen(s string) int {
	return utf8.RuneCountInString(s)
}

func stripANSI(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	inEsc := false
	for _, r := range s {
		switch {
		case !inEsc && r == 0x1b:
			inEsc = true
		case inEsc && r == 'm':
			inEsc = false
		case !inEsc:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func isCtrlU(msg tea.KeyMsg) bool {
	if msg.String() == "ctrl+u" {
		return true
	}
	return msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 0x15
}
