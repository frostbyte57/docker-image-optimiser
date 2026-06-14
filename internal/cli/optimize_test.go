package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOptimizeCheckJSONReportsPendingFixes(t *testing.T) {
	dir := t.TempDir()
	writeDockerignore(t, dir)
	path := writeDockerfile(t, dir, `FROM node:20-slim
COPY package.json ./
RUN npm ci
USER node
`)

	var stdout, stderr bytes.Buffer
	cmd := NewRootCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"optimize", "--check", "--format", "json", "--context", dir, path})

	err := cmd.Execute()
	if !errors.Is(err, ErrFindings) {
		t.Fatalf("expected ErrFindings, got %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected silent sentinel error, got stderr %q", stderr.String())
	}

	var got optimizeOutput
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode json: %v\n%s", err, stdout.String())
	}
	if !got.Changed {
		t.Fatalf("expected changed=true, got %+v", got)
	}
	if got.IssueCount != 1 || got.AutoFixCount != 1 || got.ManualCount != 0 {
		t.Fatalf("unexpected summary: %+v", got)
	}
	if len(got.Findings) != 1 || got.Findings[0].Rule != "DIO004" {
		t.Fatalf("expected DIO004 finding, got %+v", got.Findings)
	}
}

func TestOptimizeWriteUpdatesDockerfile(t *testing.T) {
	dir := t.TempDir()
	writeDockerignore(t, dir)
	path := writeDockerfile(t, dir, `FROM python:3.12-slim
RUN pip install -r requirements.txt
USER nobody
`)

	var stdout bytes.Buffer
	cmd := NewRootCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"optimize", "--write", "--context", dir, path})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v\n%s", err, stdout.String())
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(updated)
	if !strings.Contains(content, "# syntax=docker/dockerfile:1") {
		t.Fatalf("expected syntax directive:\n%s", content)
	}
	if !strings.Contains(content, "--mount=type=cache,target=/root/.cache/pip") {
		t.Fatalf("expected pip cache mount:\n%s", content)
	}
	if !strings.Contains(stdout.String(), "wrote optimised Dockerfile") {
		t.Fatalf("expected write summary, got %q", stdout.String())
	}
}

func TestLintGitHubFormatEmitsAnnotations(t *testing.T) {
	dir := t.TempDir()
	writeDockerignore(t, dir)
	path := writeDockerfile(t, dir, `FROM node:20
COPY . .
RUN npm ci
`)

	var stdout, stderr bytes.Buffer
	cmd := NewRootCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"lint", "--format", "github", "--context", dir, path})

	if err := cmd.Execute(); !errors.Is(err, ErrFindings) {
		t.Fatalf("expected ErrFindings, got %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "::warning file="+path) {
		t.Fatalf("expected GitHub warning annotation, got:\n%s", out)
	}
	if !strings.Contains(out, "title=dio DIO001") {
		t.Fatalf("expected rule id in annotation title, got:\n%s", out)
	}
}

func TestOptimiseAlias(t *testing.T) {
	dir := t.TempDir()
	writeDockerignore(t, dir)
	path := writeDockerfile(t, dir, `# syntax=docker/dockerfile:1
FROM node:20-slim
RUN --mount=type=cache,target=/root/.npm npm ci
USER node
`)

	var stdout bytes.Buffer
	cmd := NewRootCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"optimise", "--context", dir, path})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(stdout.String(), "no issues found") {
		t.Fatalf("expected clean result, got %q", stdout.String())
	}
}

func writeDockerfile(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "Dockerfile")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeDockerignore(t *testing.T, dir string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, ".dockerignore"), []byte(".git\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}
