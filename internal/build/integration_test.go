//go:build integration

package build_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yuxiangchang/docker-image-optimiser/internal/build"
)

func TestDockerBuildAndInspectIntegration(t *testing.T) {
	if err := build.Available(); err != nil {
		t.Skipf("docker not available: %v", err)
	}

	dir := t.TempDir()
	dockerfile := filepath.Join(dir, "Dockerfile")
	if err := os.WriteFile(dockerfile, []byte("FROM alpine:3.20\nRUN echo dio\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	tag := "dio-integration-test:latest"
	t.Cleanup(func() { build.Remove(tag) })

	res, err := build.Build(dir, dockerfile, tag, true)
	if err != nil {
		t.Fatal(err)
	}
	if res.Size <= 0 {
		t.Fatalf("expected positive image size, got %d", res.Size)
	}
}
