package opencode

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestListModelsParsesDedupesAndSorts(t *testing.T) {
	restore := SetRunnerForTest(func(_ context.Context, _ string, _ ...string) ([]byte, []byte, error) {
		return []byte("\n gpt-4o \n\no1-mini\ngpt-4o\n"), nil, nil
	})
	defer restore()

	models, err := ListModels(context.Background())
	if err != nil {
		t.Fatalf("ListModels: %v", err)
	}
	if got, want := strings.Join(models, ","), "gpt-4o,o1-mini"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestListModelsIncludesStderrOnError(t *testing.T) {
	restore := SetRunnerForTest(func(_ context.Context, _ string, _ ...string) ([]byte, []byte, error) {
		return nil, []byte("no such command"), errors.New("exit status 1")
	})
	defer restore()

	_, err := ListModels(context.Background())
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "no such command") {
		t.Fatalf("expected stderr in error, got %q", err.Error())
	}
}

func TestListModelsErrorsOnEmptyOutput(t *testing.T) {
	restore := SetRunnerForTest(func(_ context.Context, _ string, _ ...string) ([]byte, []byte, error) {
		return []byte("\n\n"), nil, nil
	})
	defer restore()

	_, err := ListModels(context.Background())
	if err == nil {
		t.Fatalf("expected error")
	}
}
