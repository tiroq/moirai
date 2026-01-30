package tui

import (
	"fmt"
	"sort"
	"strings"

	"moirai/internal/profile"

	tea "github.com/charmbracelet/bubbletea"
)

type agentEntry struct {
	Name    string
	Model   string
	Missing bool
}

func (m model) viewAgents() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Profile: %s\n", m.agentsProfile.Name)
	if m.agentsDirty {
		b.WriteString(dirtyStyle.Render("Unsaved changes"))
		b.WriteString("\n")
	}
	b.WriteString("\nAgents:\n")
	if len(m.agentsEntries) == 0 {
		b.WriteString("  (none)\n")
	} else {
		for i, entry := range m.agentsEntries {
			prefix := "  "
			if i == m.agentsSelected {
				prefix = "> "
			}
			name := entry.Name
			modelLabel := entry.Model
			if strings.TrimSpace(modelLabel) == "" {
				modelLabel = missingStyle.Render("(missing)")
			}
			line := fmt.Sprintf("%s%s: %s", prefix, name, modelLabel)
			if i == m.agentsSelected {
				line = selectedStyle.Render(line)
			}
			fmt.Fprintln(&b, line)
		}
	}
	b.WriteString("\n")
	if m.agentsStatus != "" {
		b.WriteString(hintStyle.Render(m.agentsStatus))
		b.WriteString("\n")
	}
	b.WriteString(hintStyle.Render("enter model · s save · r reload · a autofill · esc back · q quit · j/k/arrows move"))
	b.WriteString("\n")
	return b.String()
}

func (m model) openAgents() (tea.Model, tea.Cmd) {
	info, ok := m.selectedProfileInfo()
	if !ok {
		m.status = "No profiles available."
		return m, nil
	}
	return m, func() tea.Msg {
		cfg, err := m.actions.loadProfile(info.Path)
		return agentsLoadMsg{profile: info, cfg: cfg, err: err}
	}
}

func (m model) reloadAgents() (tea.Model, tea.Cmd) {
	if m.agentsProfile.Path == "" {
		m.agentsStatus = "No profile loaded."
		return m, nil
	}
	return m, func() tea.Msg {
		cfg, err := m.actions.loadProfile(m.agentsProfile.Path)
		return agentsLoadMsg{profile: m.agentsProfile, cfg: cfg, err: err}
	}
}

func (m model) saveAgents() (tea.Model, tea.Cmd) {
	if !m.agentsDirty {
		m.agentsStatus = "No changes to save."
		return m, nil
	}
	if m.agentsProfile.Name == "" || m.agentsProfile.Path == "" {
		m.agentsStatus = "No profile loaded."
		return m, nil
	}
	cfg := m.agentsConfig
	return m, func() tea.Msg {
		if _, err := m.actions.backupProfile(m.configDir, m.agentsProfile.Name); err != nil {
			return agentsSaveMsg{err: err}
		}
		if err := m.actions.saveProfile(m.agentsProfile.Path, cfg); err != nil {
			return agentsSaveMsg{err: err}
		}
		return agentsSaveMsg{}
	}
}

func (m model) autofillAgents() (tea.Model, tea.Cmd) {
	if !m.enableAutofill {
		m.agentsStatus = "Autofill disabled. Run with --enable-autofill."
		return m, nil
	}
	if m.agentsProfile.Name == "" || m.agentsProfile.Path == "" {
		m.agentsStatus = "No profile loaded."
		return m, nil
	}
	preset, ok := profile.PresetByName("openai")
	if !ok {
		m.agentsStatus = "Autofill preset unavailable."
		return m, nil
	}
	known := profile.KnownAgents()
	return m, func() tea.Msg {
		before := len(profile.MissingAgents(m.agentsConfig, known))
		changed := m.actions.applyAutofill(m.agentsConfig, known, preset)
		after := len(profile.MissingAgents(m.agentsConfig, known))
		filled := before - after
		if filled < 0 {
			filled = 0
		}
		if !changed {
			return agentsAutofillMsg{filled: filled, changed: false, saved: false}
		}
		if _, err := m.actions.backupProfile(m.configDir, m.agentsProfile.Name); err != nil {
			return agentsAutofillMsg{filled: filled, changed: true, saved: false, err: err}
		}
		if err := m.actions.saveProfile(m.agentsProfile.Path, m.agentsConfig); err != nil {
			return agentsAutofillMsg{filled: filled, changed: true, saved: false, err: err}
		}
		return agentsAutofillMsg{filled: filled, changed: true, saved: true}
	}
}

func (m model) openModelPicker() (tea.Model, tea.Cmd) {
	agent, ok := m.selectedAgent()
	if !ok {
		m.agentsStatus = "No agents available."
		return m, nil
	}
	m.modelTargetAgent = agent.Name
	m.modelAll = m.actions.loadModels()
	m.modelSearch = ""
	m.modelStatus = ""
	m.modelFiltered = filterModelList(m.modelAll, m.modelSearch)
	if len(m.modelFiltered) == 0 {
		m.modelSelected = -1
	} else {
		m.modelSelected = 0
	}
	m.screen = screenModels
	cmd := m.refreshModelsCmd(false)
	if cmd != nil {
		m.modelStatus = "Refreshing models..."
	}
	return m, cmd
}

func (m model) handleAgentsLoad(msg agentsLoadMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.status = msg.err.Error()
		return m, nil
	}
	m.screen = screenAgents
	m.agentsProfile = msg.profile
	m.agentsConfig = msg.cfg
	m.agentsEntries = collectAgentEntries(msg.cfg, profile.KnownAgents())
	if len(m.agentsEntries) == 0 {
		m.agentsSelected = -1
	} else {
		m.agentsSelected = 0
	}
	m.agentsDirty = false
	m.agentsStatus = ""
	return m, nil
}

func (m model) handleAgentsSave(msg agentsSaveMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.agentsStatus = msg.err.Error()
		return m, nil
	}
	m.agentsDirty = false
	m.agentsStatus = "Saved"
	return m, nil
}

func (m model) handleAgentsAutofill(msg agentsAutofillMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.agentsStatus = msg.err.Error()
		if msg.changed {
			m.agentsDirty = true
			m.agentsEntries = collectAgentEntries(m.agentsConfig, profile.KnownAgents())
		}
		return m, nil
	}
	if !msg.changed {
		m.agentsStatus = "No missing models to autofill."
		return m, nil
	}
	m.agentsEntries = collectAgentEntries(m.agentsConfig, profile.KnownAgents())
	m.agentsDirty = !msg.saved
	m.agentsStatus = fmt.Sprintf("Autofilled %d agents", msg.filled)
	return m, nil
}

func (m *model) moveAgentsSelection(delta int) {
	if len(m.agentsEntries) == 0 {
		return
	}
	m.agentsSelected += delta
	if m.agentsSelected < 0 {
		m.agentsSelected = 0
	}
	if m.agentsSelected >= len(m.agentsEntries) {
		m.agentsSelected = len(m.agentsEntries) - 1
	}
}

func (m *model) updateAgentsEntries() {
	selectedName := ""
	if m.agentsSelected >= 0 && m.agentsSelected < len(m.agentsEntries) {
		selectedName = m.agentsEntries[m.agentsSelected].Name
	}
	m.agentsEntries = collectAgentEntries(m.agentsConfig, profile.KnownAgents())
	if len(m.agentsEntries) == 0 {
		m.agentsSelected = -1
		return
	}
	if selectedName == "" {
		m.agentsSelected = 0
		return
	}
	for i, entry := range m.agentsEntries {
		if entry.Name == selectedName {
			m.agentsSelected = i
			return
		}
	}
	m.agentsSelected = 0
}

func (m model) selectedAgent() (agentEntry, bool) {
	if m.agentsSelected < 0 || m.agentsSelected >= len(m.agentsEntries) {
		return agentEntry{}, false
	}
	return m.agentsEntries[m.agentsSelected], true
}

func collectAgentEntries(cfg *profile.RootConfig, knownAgents []string) []agentEntry {
	entries := make([]agentEntry, 0, len(knownAgents))
	seen := make(map[string]struct{}, len(knownAgents))
	for _, name := range knownAgents {
		seen[name] = struct{}{}
		model := ""
		if cfg != nil && cfg.Agents != nil {
			if entry, ok := cfg.Agents[name]; ok {
				model = entry.Model
			}
		}
		entries = append(entries, agentEntry{
			Name:    name,
			Model:   model,
			Missing: strings.TrimSpace(model) == "",
		})
	}
	custom := make([]string, 0)
	if cfg != nil && cfg.Agents != nil {
		for name := range cfg.Agents {
			if _, ok := seen[name]; ok {
				continue
			}
			custom = append(custom, name)
		}
	}
	sort.Strings(custom)
	for _, name := range custom {
		model := ""
		if cfg != nil && cfg.Agents != nil {
			model = cfg.Agents[name].Model
		}
		entries = append(entries, agentEntry{
			Name:    name,
			Model:   model,
			Missing: strings.TrimSpace(model) == "",
		})
	}
	return entries
}
