package tui

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"moirai/internal/profile"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) viewModels() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Model Picker: %s\n", m.modelTargetAgent)
	fmt.Fprintf(&b, "Search: %s\n\n", m.modelSearch)
	if len(m.modelFiltered) == 0 {
		b.WriteString("  (none)\n")
	} else {
		for i, modelName := range m.modelFiltered {
			prefix := "  "
			if i == m.modelSelected {
				prefix = "> "
			}
			line := fmt.Sprintf("%s%s", prefix, modelName)
			if i == m.modelSelected {
				line = selectedStyle.Render(line)
			}
			fmt.Fprintln(&b, line)
		}
	}
	b.WriteString("\n")
	b.WriteString(hintStyle.Render("enter select · esc cancel · j/k/arrows move · type to filter"))
	b.WriteString("\n")
	return b.String()
}

func (m model) handleModelPickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.screen = screenAgents
		return m, nil
	case tea.KeyEnter:
		return m.selectModel()
	case tea.KeyBackspace, tea.KeyDelete:
		if m.modelSearch != "" {
			runes := []rune(m.modelSearch)
			if len(runes) > 0 {
				m.modelSearch = string(runes[:len(runes)-1])
			}
			m.updateModelFilter()
		}
		return m, nil
	case tea.KeyRunes:
		m.modelSearch += string(msg.Runes)
		m.updateModelFilter()
		return m, nil
	}

	switch msg.String() {
	case "j", "down":
		m.moveModelSelection(1)
	case "k", "up":
		m.moveModelSelection(-1)
	}
	return m, nil
}

func (m model) selectModel() (tea.Model, tea.Cmd) {
	if m.modelSelected < 0 || m.modelSelected >= len(m.modelFiltered) {
		m.agentsStatus = "No model selected."
		m.screen = screenAgents
		return m, nil
	}
	modelName := m.modelFiltered[m.modelSelected]
	changed, err := profile.SetAgentModel(m.agentsConfig, m.modelTargetAgent, modelName)
	if err != nil {
		m.agentsStatus = err.Error()
		m.screen = screenAgents
		return m, nil
	}
	if changed {
		m.agentsDirty = true
	}
	m.updateAgentsEntries()
	m.screen = screenAgents
	return m, nil
}

func (m *model) updateModelFilter() {
	m.modelFiltered = filterModelList(m.modelAll, m.modelSearch)
	if len(m.modelFiltered) == 0 {
		m.modelSelected = -1
		return
	}
	if m.modelSelected < 0 || m.modelSelected >= len(m.modelFiltered) {
		m.modelSelected = 0
	}
}

func (m *model) moveModelSelection(delta int) {
	if len(m.modelFiltered) == 0 {
		return
	}
	m.modelSelected += delta
	if m.modelSelected < 0 {
		m.modelSelected = 0
	}
	if m.modelSelected >= len(m.modelFiltered) {
		m.modelSelected = len(m.modelFiltered) - 1
	}
}

func filterModelList(models []string, query string) []string {
	if query == "" {
		return append([]string(nil), models...)
	}
	filtered := make([]string, 0, len(models))
	lowered := strings.ToLower(query)
	for _, modelName := range models {
		if strings.Contains(strings.ToLower(modelName), lowered) {
			filtered = append(filtered, modelName)
		}
	}
	return filtered
}

func loadModelList() []string {
	path := filepath.Join("configs", "models.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		return defaultModelList()
	}
	models := parseModelList(data)
	if len(models) == 0 {
		return defaultModelList()
	}
	return models
}

func parseModelList(data []byte) []string {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	models := make([]string, 0)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		models = append(models, line)
	}
	return models
}

func defaultModelList() []string {
	return []string{
		"gpt-4o-mini",
		"gpt-4o",
		"gpt-4.1",
		"gpt-4.1-mini",
		"gpt-4.1-nano",
		"gpt-4.5-preview",
		"o1-mini",
		"o1-preview",
	}
}
