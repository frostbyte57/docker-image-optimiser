package rewrite_test

import (
	"strings"
	"testing"

	"github.com/yuxiangchang/docker-image-optimiser/internal/parser"
	"github.com/yuxiangchang/docker-image-optimiser/internal/rewrite"
	"github.com/yuxiangchang/docker-image-optimiser/internal/rules"
)

func TestCacheMountIsDefault(t *testing.T) {
	src := []byte("FROM python:3.12-slim\nRUN pip install -r requirements.txt\n")

	res, err := rewrite.Apply(src, rules.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Content, "--mount=type=cache,target=/root/.cache/pip") {
		t.Errorf("expected a pip cache mount:\n%s", res.Content)
	}
	// Cache mounts require the syntax directive, prepended automatically.
	if !strings.HasPrefix(res.Content, "# syntax=docker/dockerfile:1") {
		t.Errorf("expected syntax directive prepended:\n%s", res.Content)
	}
}

func TestConservativeUsesFlag(t *testing.T) {
	src := []byte("FROM python:3.12-slim\nRUN pip install -r requirements.txt\n")

	res, err := rewrite.Apply(src, rules.Options{Conservative: true})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Content, "pip install --no-cache-dir") {
		t.Errorf("expected --no-cache-dir in conservative mode:\n%s", res.Content)
	}
	if strings.Contains(res.Content, "--mount=type=cache") {
		t.Errorf("conservative mode should not inject cache mounts:\n%s", res.Content)
	}
}

func TestSystemManagersStayConservative(t *testing.T) {
	// apt should get --no-install-recommends + list cleanup, never a cache mount.
	src := []byte("FROM debian:12-slim\nRUN apt-get install -y curl\n")

	res, err := rewrite.Apply(src, rules.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Content, "--no-install-recommends") ||
		!strings.Contains(res.Content, "rm -rf /var/lib/apt/lists/*") {
		t.Errorf("expected conservative apt fixes:\n%s", res.Content)
	}
	if strings.Contains(res.Content, "--mount=type=cache") {
		t.Errorf("apt should not get a cache mount by default:\n%s", res.Content)
	}
}

func TestMultiStageIsAnnotated(t *testing.T) {
	src := []byte("FROM golang:1.22\nCOPY . .\nRUN go build -o /app .\n")

	res, err := rewrite.Apply(src, rules.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Content, "# dio[DIO007]") {
		t.Errorf("expected multi-stage annotation:\n%s", res.Content)
	}
}

// Rewriting twice must be a no-op the second time.
func TestIdempotent(t *testing.T) {
	src := []byte("FROM python:3.12\nCOPY . .\nRUN pip install -r requirements.txt\n")

	once, err := rewrite.Apply(src, rules.Options{})
	if err != nil {
		t.Fatal(err)
	}
	twice, err := rewrite.Apply([]byte(once.Content), rules.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if twice.Content != once.Content {
		t.Errorf("rewrite not idempotent:\n--- once ---\n%s\n--- twice ---\n%s", once.Content, twice.Content)
	}
}

// The auto-fixable cache rule should not fire again after a default rewrite.
func TestRewriteSatisfiesCacheRule(t *testing.T) {
	src := []byte("FROM node:20-slim\nCOPY package.json ./\nRUN npm ci\n")

	res, err := rewrite.Apply(src, rules.Options{})
	if err != nil {
		t.Fatal(err)
	}
	ins, err := parser.Parse(strings.NewReader(res.Content))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range rules.Run(ins, rules.Options{Source: res.Content}) {
		if f.Rule == "DIO004" {
			t.Errorf("DIO004 still fires after rewrite:\n%s", res.Content)
		}
	}
}
