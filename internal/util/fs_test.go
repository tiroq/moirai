package util

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestExpandUser(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}

	got, err := ExpandUser("~")
	if err != nil {
		t.Fatalf("ExpandUser: %v", err)
	}
	if got != home {
		t.Fatalf("expected %s, got %s", home, got)
	}

	got, err = ExpandUser("~/config")
	if err != nil {
		t.Fatalf("ExpandUser: %v", err)
	}
	if got != filepath.Join(home, "config") {
		t.Fatalf("expected %s, got %s", filepath.Join(home, "config"), got)
	}

	raw := "/tmp/example"
	got, err = ExpandUser(raw)
	if err != nil {
		t.Fatalf("ExpandUser: %v", err)
	}
	if got != raw {
		t.Fatalf("expected %s, got %s", raw, got)
	}
}

func TestFileExistsAndListDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if !FileExists(path) {
		t.Fatalf("expected file to exist")
	}

	entries, err := ListDir(dir)
	if err != nil {
		t.Fatalf("ListDir: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "file.txt" {
		t.Fatalf("unexpected entries: %v", entries)
	}
}

func TestTimestampFormat(t *testing.T) {
	value := Timestamp()
	if len(value) != len("20060102-150405") {
		t.Fatalf("unexpected timestamp length: %s", value)
	}
	if matched := regexp.MustCompile(`^\d{8}-\d{6}$`).MatchString(value); !matched {
		t.Fatalf("unexpected timestamp format: %s", value)
	}
}

func TestCopyFileAtomic(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	content := []byte("atomic copy")
	if err := os.WriteFile(src, content, 0o640); err != nil {
		t.Fatalf("write src: %v", err)
	}

	if err := CopyFileAtomic(src, dst); err != nil {
		t.Fatalf("CopyFileAtomic: %v", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("unexpected content: %s", string(got))
	}
}

func TestCopyFileAtomicMissingSource(t *testing.T) {
	dir := t.TempDir()
	if err := CopyFileAtomic(filepath.Join(dir, "missing"), filepath.Join(dir, "dst.txt")); err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestRunCommandCapturesOutput(t *testing.T) {
	stdout, stderr, err := RunCommand(context.Background(), "sh", "-c", "printf 'hello'")
	if err != nil {
		t.Fatalf("RunCommand: %v", err)
	}
	if stdout != "hello" {
		t.Fatalf("expected stdout hello, got %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}
