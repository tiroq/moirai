package opencode

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"
)

const listModelsTimeout = 4 * time.Second

// Runner executes a command and returns its stdout and stderr.
type Runner func(ctx context.Context, name string, args ...string) (stdout []byte, stderr []byte, err error)

var runCommand Runner = execRunner

func execRunner(ctx context.Context, name string, args ...string) ([]byte, []byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	return out, stderr.Bytes(), err
}

// SetRunnerForTest replaces the command runner for tests.
// It returns a restore func that resets the runner to its previous value.
func SetRunnerForTest(r Runner) (restore func()) {
	prev := runCommand
	runCommand = r
	return func() { runCommand = prev }
}

// ListModels lists available models via `opencode models`.
func ListModels(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, listModelsTimeout)
	defer cancel()

	stdout, stderr, err := runCommand(ctx, "opencode", "models")
	if err != nil {
		msg := strings.TrimSpace(string(stderr))
		if msg == "" {
			return nil, err
		}
		return nil, fmt.Errorf("%w: %s", err, msg)
	}

	models := parseModels(stdout)
	if len(models) == 0 {
		return nil, fmt.Errorf("opencode models returned no models")
	}
	return models, nil
}

func parseModels(data []byte) []string {
	lines := strings.Split(string(data), "\n")
	seen := make(map[string]struct{}, len(lines))
	models := make([]string, 0, len(lines))
	for _, line := range lines {
		model := strings.TrimSpace(line)
		if model == "" {
			continue
		}
		if _, ok := seen[model]; ok {
			continue
		}
		seen[model] = struct{}{}
		models = append(models, model)
	}
	sort.Strings(models)
	return models
}

