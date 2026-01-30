package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	modelsCache "moirai/internal/models"
	"moirai/internal/opencode"
	"moirai/internal/profile"

	tea "github.com/charmbracelet/bubbletea"
)

const modelsCacheTTL = 24 * time.Hour
const modelsLoadBudget = 50 * time.Millisecond

func (m model) viewModels() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Model Picker: %s\n", m.modelTargetAgent)
	fmt.Fprintf(&b, "Search: %s\n\n", m.modelSearch)
	if m.modelStatus != "" {
		b.WriteString(hintStyle.Render(m.modelStatus))
		b.WriteString("\n\n")
	}
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
	b.WriteString(hintStyle.Render("enter select · R refresh · esc cancel · j/k/arrows move · type to filter"))
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
		if string(msg.Runes) == "R" {
			m.modelStatus = "Refreshing models..."
			return m, m.refreshModelsCmd(true)
		}
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
	configHome, err := resolveConfigHome()
	if err != nil {
		return defaultModelList()
	}

	type result struct {
		models []string
		ok     bool
		err    error
	}
	ch := make(chan result, 1)
	go func() {
		models, ok, err := modelsCache.LoadCachedModels(configHome)
		ch <- result{models: models, ok: ok, err: err}
	}()

	select {
	case res := <-ch:
		if res.err != nil || !res.ok || len(res.models) == 0 {
			return defaultModelList()
		}
		return res.models
	case <-time.After(modelsLoadBudget):
		return defaultModelList()
	}
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

func resolveConfigHome() (string, error) {
	if xdgHome := os.Getenv("XDG_CONFIG_HOME"); xdgHome != "" {
		return filepath.Clean(xdgHome), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config"), nil
}

func (m model) refreshModelsCmd(force bool) tea.Cmd {
	configHome, err := resolveConfigHome()
	if err != nil {
		return func() tea.Msg { return ModelsRefreshFailedMsg{Err: err} }
	}

	current := append([]string(nil), m.modelAll...)
	if !force {
		age, ok, err := modelsCache.CacheAge(configHome)
		if err != nil {
			return func() tea.Msg { return ModelsRefreshFailedMsg{Err: err} }
		}
		if ok && age < modelsCacheTTL {
			return nil
		}
	}

	return func() tea.Msg {
		ctx := context.Background()
		models, err := opencode.ListModels(ctx)
		if err != nil {
			return ModelsRefreshFailedMsg{Err: err}
		}
		if !reflect.DeepEqual(models, current) {
			if err := modelsCache.SaveCachedModelsAtomic(configHome, models); err != nil {
				return ModelsRefreshFailedMsg{Err: err}
			}
		}
		return ModelsRefreshedMsg{Models: models}
	}
}
