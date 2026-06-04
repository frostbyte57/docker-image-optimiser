package build

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyDir(t *testing.T) {
	src := t.TempDir()
	// A nested file and a .git dir that must be skipped.
	mustWrite(t, filepath.Join(src, "app", "main.go"), "package main")
	mustWrite(t, filepath.Join(src, ".git", "config"), "[core]")

	dst, err := copyDir(src)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dst)

	if got, err := os.ReadFile(filepath.Join(dst, "app", "main.go")); err != nil || string(got) != "package main" {
		t.Errorf("nested file not copied: %v %q", err, got)
	}
	if _, err := os.Stat(filepath.Join(dst, ".git")); !os.IsNotExist(err) {
		t.Errorf(".git should be skipped, got err=%v", err)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
