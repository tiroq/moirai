package tui

import (
	"fmt"

	"moirai/internal/profile"

	tea "github.com/charmbracelet/bubbletea"
)

type screenID int

const (
	screenProfiles screenID = iota
	screenBackups
	screenDiff
)

type diffMode int

const (
	diffModeLastBackup diffMode = iota
	diffModeActiveProfile
)

type model struct {
	configDir  string
	profiles   []profile.ProfileInfo
	activeName string
	hasActive  bool
	selected   int
	status     string

	screen screenID

	backupsProfile string
	backups        []string
	backupsMessage string

	diffMode    diffMode
	diffProfile string
	diffAgainst string
	diffContent string
	diffMessage string
	viewport    diffViewport

	width  int
	height int

	actions modelActions
}

type applyResultMsg struct {
	profile string
	err     error
}

type backupsResultMsg struct {
	profile string
	backups []string
	err     error
}

type diffResultMsg struct {
	mode      diffMode
	profile   string
	against   string
	diff      string
	hasBackup bool
	err       error
}

func newModel(configDir string, profiles []profile.ProfileInfo, activeName string, hasActive bool) model {
	return newModelWithActions(configDir, profiles, activeName, hasActive, defaultActions())
}

func newModelWithActions(configDir string, profiles []profile.ProfileInfo, activeName string, hasActive bool, actions modelActions) model {
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

	actions = normalizeActions(actions)
	m := model{
		configDir:  configDir,
		profiles:   profiles,
		activeName: activeName,
		hasActive:  hasActive,
		selected:   selected,
		screen:     screenProfiles,
		width:      80,
		height:     24,
		actions:    actions,
	}
	m.viewport = newDiffViewport()
	m.resizeViewport()
	return m
}

func normalizeActions(actions modelActions) modelActions {
	defaults := defaultActions()
	if actions.applyProfile == nil {
		actions.applyProfile = defaults.applyProfile
	}
	if actions.listProfileBackups == nil {
		actions.listProfileBackups = defaults.listProfileBackups
	}
	if actions.activeProfile == nil {
		actions.activeProfile = defaults.activeProfile
	}
	if actions.diffAgainstLastBackup == nil {
		actions.diffAgainstLastBackup = defaults.diffAgainstLastBackup
	}
	if actions.diffBetweenProfiles == nil {
		actions.diffBetweenProfiles = defaults.diffBetweenProfiles
	}
	return actions
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeViewport()
	case applyResultMsg:
		return m.handleApplyResult(msg)
	case backupsResultMsg:
		return m.handleBackupsResult(msg)
	case diffResultMsg:
		return m.handleDiffResult(msg)
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m model) View() string {
	switch m.screen {
	case screenBackups:
		return m.viewBackups()
	case screenDiff:
		return m.viewDiff()
	default:
		return m.viewProfiles()
	}
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if key == "q" || key == "ctrl+c" {
		return m, tea.Quit
	}
	if key == "esc" && m.screen == screenProfiles {
		return m, tea.Quit
	}

	switch m.screen {
	case screenProfiles:
		switch key {
		case "j", "down":
			m.status = ""
			m.moveSelection(1)
		case "k", "up":
			m.status = ""
			m.moveSelection(-1)
		case "enter":
			return m.applySelected()
		case "b":
			return m.openBackups()
		case "d":
			return m.openDiff(diffModeLastBackup)
		}
	case screenBackups:
		if key == "esc" {
			m.screen = screenProfiles
		}
	case screenDiff:
		switch key {
		case "esc":
			m.screen = screenProfiles
			return m, nil
		case "j", "down":
			m.viewport.LineDown(1)
			return m, nil
		case "k", "up":
			m.viewport.LineUp(1)
			return m, nil
		case "pgdown":
			m.viewport.PageDown()
			return m, nil
		case "pgup":
			m.viewport.PageUp()
			return m, nil
		case "a":
			return m.openDiff(diffModeActiveProfile)
		case "d":
			return m.openDiff(diffModeLastBackup)
		}
		return m, nil
	}

	return m, nil
}

func (m model) applySelected() (tea.Model, tea.Cmd) {
	name, ok := m.selectedProfile()
	if !ok {
		m.status = "No profiles available."
		return m, nil
	}
	return m, func() tea.Msg {
		err := m.actions.applyProfile(m.configDir, name)
		return applyResultMsg{profile: name, err: err}
	}
}

func (m model) openBackups() (tea.Model, tea.Cmd) {
	name, ok := m.selectedProfile()
	if !ok {
		m.status = "No profiles available."
		return m, nil
	}
	return m, func() tea.Msg {
		backups, err := m.actions.listProfileBackups(m.configDir, name)
		return backupsResultMsg{profile: name, backups: backups, err: err}
	}
}

func (m model) openDiff(mode diffMode) (tea.Model, tea.Cmd) {
	name, ok := m.selectedProfile()
	if !ok {
		m.status = "No profiles available."
		return m, nil
	}

	switch mode {
	case diffModeActiveProfile:
		if !m.hasActive {
			return m, func() tea.Msg {
				return diffResultMsg{
					mode:    mode,
					profile: name,
					err:     fmt.Errorf("no active profile to diff against"),
				}
			}
		}
		activeName := m.activeName
		return m, func() tea.Msg {
			diff, err := m.actions.diffBetweenProfiles(m.configDir, activeName, name)
			return diffResultMsg{
				mode:    mode,
				profile: name,
				against: activeName,
				diff:    diff,
				err:     err,
			}
		}
	default:
		return m, func() tea.Msg {
			diff, hasBackup, err := m.actions.diffAgainstLastBackup(m.configDir, name)
			return diffResultMsg{
				mode:      mode,
				profile:   name,
				against:   "last-backup",
				diff:      diff,
				hasBackup: hasBackup,
				err:       err,
			}
		}
	}
}

func (m model) handleApplyResult(msg applyResultMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.status = msg.err.Error()
		return m, nil
	}

	m.status = fmt.Sprintf("Applied: %s", msg.profile)
	activeName, ok, err := m.actions.activeProfile(m.configDir)
	if err != nil {
		m.status = err.Error()
		return m, nil
	}
	m.activeName = activeName
	m.hasActive = ok
	return m, nil
}

func (m model) handleBackupsResult(msg backupsResultMsg) (tea.Model, tea.Cmd) {
	m.screen = screenBackups
	m.backupsProfile = msg.profile
	m.backups = msg.backups
	if msg.err != nil {
		m.backupsMessage = msg.err.Error()
		return m, nil
	}
	if len(msg.backups) == 0 {
		m.backupsMessage = "(none)"
		return m, nil
	}
	m.backupsMessage = ""
	return m, nil
}

func (m model) handleDiffResult(msg diffResultMsg) (tea.Model, tea.Cmd) {
	m.screen = screenDiff
	m.diffMode = msg.mode
	m.diffProfile = msg.profile
	m.diffAgainst = msg.against
	m.diffContent = msg.diff
	m.diffMessage = ""

	if msg.err != nil {
		m.diffMessage = msg.err.Error()
		m.diffContent = ""
	}
	if msg.mode == diffModeLastBackup && !msg.hasBackup && msg.err == nil {
		m.diffMessage = fmt.Sprintf("No backups found for profile: %s", msg.profile)
		m.diffContent = ""
	}

	m.viewport.SetContent(m.diffContent)
	m.viewport.GotoTop()
	m.resizeViewport()
	return m, nil
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

func (m model) selectedProfile() (string, bool) {
	if m.selected < 0 || m.selected >= len(m.profiles) {
		return "", false
	}
	return m.profiles[m.selected].Name, true
}

func (m *model) resizeViewport() {
	reserved := 4
	if m.height-reserved < 1 {
		m.viewport.height = 1
	} else {
		m.viewport.height = m.height - reserved
	}
	if m.width < 1 {
		m.viewport.width = 1
	} else {
		m.viewport.width = m.width
	}
	m.viewport.clamp()
}
