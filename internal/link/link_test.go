package link

import (
	"os"
	"path/filepath"
	"testing"
)

func TestActiveProfileSymlink(t *testing.T) {
	dir := t.TempDir()
	profileName := "antigravity"
	profileFile := "oh-my-opencode.json." + profileName
	profilePath := filepath.Join(dir, profileFile)

	if err := os.WriteFile(profilePath, []byte("{}"), 0o600); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	activePath := filepath.Join(dir, "oh-my-opencode.json")
	if err := os.Symlink(profileFile, activePath); err != nil {
		t.Skipf("symlink not supported: %v", err)
	}

	name, ok, err := ActiveProfile(dir)
	if err != nil {
		t.Fatalf("ActiveProfile: %v", err)
	}
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if name != profileName {
		t.Fatalf("expected %q, got %q", profileName, name)
	}
}

func TestActiveProfileMissingOrRegularFile(t *testing.T) {
	dir := t.TempDir()

	name, ok, err := ActiveProfile(dir)
	if err != nil {
		t.Fatalf("ActiveProfile missing: %v", err)
	}
	if ok || name != "" {
		t.Fatalf("expected missing to be inactive")
	}

	activePath := filepath.Join(dir, "oh-my-opencode.json")
	if err := os.WriteFile(activePath, []byte("{}"), 0o600); err != nil {
		t.Fatalf("write active file: %v", err)
	}

	name, ok, err = ActiveProfile(dir)
	if err != nil {
		t.Fatalf("ActiveProfile regular file: %v", err)
	}
	if ok || name != "" {
		t.Fatalf("expected regular file to be inactive")
	}
}
