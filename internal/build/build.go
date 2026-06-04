// Package build shells out to Docker to build images and measure them, so
// `dio bench` can compare a Dockerfile against its optimised rewrite.
package build

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Result captures one image build.
type Result struct {
	Tag         string
	Duration    time.Duration // cold build wall-clock
	Size        int64         // bytes
	WarmRebuild time.Duration // rebuild time after a source change (0 if not measured)
}

// Available reports whether the docker CLI is installed and its daemon is
// reachable, so bench can fail early with a clear message.
func Available() error {
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker CLI not found on PATH: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := exec.CommandContext(ctx, "docker", "info").Run(); err != nil {
		return fmt.Errorf("docker daemon not reachable — is Docker running? (%w)", err)
	}
	return nil
}

// Build builds dockerfile against contextDir, tags it, and measures wall-clock
// time and resulting image size. BuildKit is enabled for parallelism.
func Build(contextDir, dockerfile, tag string, noCache bool) (Result, error) {
	args := []string{"build", "-f", dockerfile, "-t", tag}
	if noCache {
		args = append(args, "--no-cache")
	}
	args = append(args, contextDir)

	cmd := exec.Command("docker", args...)
	cmd.Env = append(os.Environ(), "DOCKER_BUILDKIT=1")

	start := time.Now()
	out, err := cmd.CombinedOutput()
	dur := time.Since(start)
	if err != nil {
		return Result{}, fmt.Errorf("docker build failed for %s: %w\n%s", tag, err, out)
	}

	size, err := imageSize(tag)
	if err != nil {
		return Result{}, err
	}
	return Result{Tag: tag, Duration: dur, Size: size}, nil
}

// Remove deletes an image, ignoring errors (best-effort cleanup).
func Remove(tag string) {
	_ = exec.Command("docker", "rmi", "-f", tag).Run()
}

func imageSize(tag string) (int64, error) {
	out, err := exec.Command("docker", "image", "inspect", "-f", "{{.Size}}", tag).Output()
	if err != nil {
		return 0, fmt.Errorf("inspecting size of %s: %w", tag, err)
	}
	return strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
}
