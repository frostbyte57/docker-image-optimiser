//go:build integration

package inspect_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yuxiangchang/docker-image-optimiser/internal/build"
	"github.com/yuxiangchang/docker-image-optimiser/internal/inspect"
)

func TestDockerHistoryIntegration(t *testing.T) {
	if err := build.Available(); err != nil {
		t.Skipf("docker not available: %v", err)
	}

	dir := t.TempDir()
	dockerfile := filepath.Join(dir, "Dockerfile")
	if err := os.WriteFile(dockerfile, []byte("FROM alpine:3.20\nRUN echo dio\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	tag := "dio-history-integration-test:latest"
	t.Cleanup(func() { build.Remove(tag) })
	if _, err := build.Build(dir, dockerfile, tag, true); err != nil {
		t.Fatal(err)
	}

	layers, err := inspect.History(tag)
	if err != nil {
		t.Fatal(err)
	}
	if len(layers) == 0 {
		t.Fatal("expected at least one image layer")
	}
}
