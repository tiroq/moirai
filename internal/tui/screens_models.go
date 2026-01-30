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
const modelPageSize = 10

func (m model) viewModels() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Model Picker: %s\n", m.modelTargetAgent)
	fmt.Fprintf(&b, "Search: %s\n\n", m.modelSearch)
	if len(m.modelFiltered) == 0 {
		b.WriteString("  (none)\n")
	} else {
		start, end := modelWindow(len(m.modelFiltered), m.modelSelected, modelPageSize)
		fmt.Fprintf(&b, "Showing %d-%d of %d\n\n", start+1, end, len(m.modelFiltered))
		for i := start; i < end; i++ {
			modelName := m.modelFiltered[i]
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
	return b.String()
}

func (m model) handleModelPickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isCtrlU(msg) {
		if m.modelSearch != "" {
			m.modelSearch = ""
			m.updateModelFilter()
		}
		return m, nil
	}
	switch msg.Type {
	case tea.KeyEsc:
		m.screen = screenAgents
		return m, nil
	case tea.KeyEnter:
		return m.selectModel()
	case tea.KeyPgUp:
		m.moveModelSelection(-modelPageSize)
		return m, nil
	case tea.KeyPgDown:
		m.moveModelSelection(modelPageSize)
		return m, nil
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
			m.setStatus(statusKindInfo, "Refreshing models...")
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
		m.setStatus(statusKindError, "No model selected.")
		m.screen = screenAgents
		return m, nil
	}
	modelName := m.modelFiltered[m.modelSelected]
	changed, err := profile.SetAgentModel(m.agentsConfig, m.modelTargetAgent, modelName)
	if err != nil {
		m.setStatus(statusKindError, err.Error())
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

func modelWindow(total, selected, pageSize int) (start, end int) {
	if total <= 0 || pageSize <= 0 {
		return 0, 0
	}
	if selected < 0 {
		selected = 0
	}
	if selected >= total {
		selected = total - 1
	}
	start = (selected / pageSize) * pageSize
	end = start + pageSize
	if end > total {
		end = total
	}
	return start, end
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
