package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverProfiles(t *testing.T) {
	dir := t.TempDir()

	files := []string{
		"oh-my-opencode.json.zeta",
		"oh-my-opencode.json.alpha",
		"oh-my-opencode.json.beta.bak.1",
		"notes.txt",
	}
	for _, name := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("{}"), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	profiles, err := DiscoverProfiles(dir)
	if err != nil {
		t.Fatalf("DiscoverProfiles: %v", err)
	}

	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}

	if profiles[0].Name != "alpha" || profiles[1].Name != "zeta" {
		t.Fatalf("unexpected order: %#v", profiles)
	}

	if profiles[0].Path != filepath.Join(dir, "oh-my-opencode.json.alpha") {
		t.Fatalf("unexpected path: %s", profiles[0].Path)
	}
}
