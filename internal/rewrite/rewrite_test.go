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

// Conservative mode must add the flag even when the user wrote a verb alias
// (pip3) that differs from the registry's canonical "pip install". The rewrite
// previously searched for "pip install", found nothing, and silently no-oped
// while still reporting the line as fixed.
func TestConservativeHandlesVerbAlias(t *testing.T) {
	src := []byte("FROM python:3.12-slim\nRUN pip3 install flask\n")

	res, err := rewrite.Apply(src, rules.Options{Conservative: true})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Content, "pip3 install --no-cache-dir") {
		t.Errorf("expected --no-cache-dir added to pip3 install:\n%s", res.Content)
	}
	if !res.Changed {
		t.Error("expected Changed=true, but the rewrite reported no change")
	}
}

// `apt install` (not `apt-get install`) must also receive --no-install-recommends,
// so DIO002 stays in lockstep with the ecosystem and DIO003.
func TestAptInstallGetsNoRecommends(t *testing.T) {
	src := []byte("FROM debian:12-slim\nRUN apt update && apt install -y curl\n")

	res, err := rewrite.Apply(src, rules.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Content, "apt install --no-install-recommends") {
		t.Errorf("expected --no-install-recommends on apt install:\n%s", res.Content)
	}
}

// Exec-form RUN must be left untouched: shell-form flag/mount injection would
// corrupt the JSON array.
func TestExecFormRunUntouched(t *testing.T) {
	src := []byte("FROM python:3.12-slim\n" + `RUN ["sh", "-c", "pip install flask"]` + "\n")

	res, err := rewrite.Apply(src, rules.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(res.Content, "--mount=type=cache") {
		t.Errorf("exec-form RUN should not get a cache mount:\n%s", res.Content)
	}
	if !strings.Contains(res.Content, `RUN ["sh", "-c", "pip install flask"]`) {
		t.Errorf("exec-form RUN was altered:\n%s", res.Content)
	}
}

// DIO008 (missing syntax directive for a hand-written cache mount) is auto-applied
// by the rewriter, so it belongs in Applied, not Manual.
func TestSyntaxDirectiveReportedAsApplied(t *testing.T) {
	src := []byte("FROM node:20-slim\nRUN --mount=type=cache,target=/root/.npm npm ci\n")

	res, err := rewrite.Apply(src, rules.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(res.Content, "# syntax=docker/dockerfile:1") {
		t.Fatalf("expected syntax directive prepended:\n%s", res.Content)
	}
	if !anyContains(res.Applied, "DIO008") {
		t.Errorf("DIO008 should be reported as applied, got Applied=%v", res.Applied)
	}
	if anyContains(res.Manual, "DIO008") {
		t.Errorf("DIO008 should not be annotated as manual, got Manual=%v", res.Manual)
	}
}

func anyContains(ss []string, sub string) bool {
	for _, s := range ss {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
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
