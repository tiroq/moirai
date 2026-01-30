package models

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadCachedModelsMissing(t *testing.T) {
	configHome := t.TempDir()
	models, ok, err := LoadCachedModels(configHome)
	if err != nil {
		t.Fatalf("LoadCachedModels: %v", err)
	}
	if ok {
		t.Fatalf("expected ok=false")
	}
	if models != nil {
		t.Fatalf("expected nil models, got %#v", models)
	}
}

func TestLoadCachedModelsReturnsModels(t *testing.T) {
	configHome := t.TempDir()
	path := filepath.Join(configHome, "opencode", "moirai", "models.txt")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte("gpt-4o\n\no1-mini\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	models, ok, err := LoadCachedModels(configHome)
	if err != nil {
		t.Fatalf("LoadCachedModels: %v", err)
	}
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if got, want := strings.Join(models, ","), "gpt-4o,o1-mini"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestSaveCachedModelsAtomicWritesAndCleansTemp(t *testing.T) {
	configHome := t.TempDir()
	if err := SaveCachedModelsAtomic(configHome, []string{"gpt-4o", "o1-mini"}); err != nil {
		t.Fatalf("SaveCachedModelsAtomic: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(configHome, "opencode", "moirai", "models.txt"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if got, want := string(data), "gpt-4o\no1-mini\n"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
	if _, err := os.Stat(filepath.Join(configHome, "opencode", "moirai", "models.meta.json")); err != nil {
		t.Fatalf("expected meta file: %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(configHome, "opencode", "moirai"))
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".tmp-") {
			t.Fatalf("unexpected temp file remaining: %s", entry.Name())
		}
	}
}

func TestCacheAgeReportsMissing(t *testing.T) {
	configHome := t.TempDir()
	_, ok, err := CacheAge(configHome)
	if err != nil {
		t.Fatalf("CacheAge: %v", err)
	}
	if ok {
		t.Fatalf("expected ok=false")
	}
}

func TestCacheAgeUsesModTime(t *testing.T) {
	configHome := t.TempDir()
	path := filepath.Join(configHome, "opencode", "moirai", "models.txt")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte("gpt-4o\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	then := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(path, then, then); err != nil {
		t.Fatalf("Chtimes: %v", err)
	}

	age, ok, err := CacheAge(configHome)
	if err != nil {
		t.Fatalf("CacheAge: %v", err)
	}
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if age < 47*time.Hour {
		t.Fatalf("expected age ~48h, got %v", age)
	}
}
