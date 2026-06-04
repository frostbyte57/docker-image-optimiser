package rules_test

import (
	"strings"
	"testing"

	"github.com/yuxiangchang/docker-image-optimiser/internal/parser"
	"github.com/yuxiangchang/docker-image-optimiser/internal/rules"
)

// lint parses a Dockerfile snippet and returns the rule ids that fired.
func lint(t *testing.T, dockerfile string) map[string]bool {
	t.Helper()
	ins, err := parser.Parse(strings.NewReader(dockerfile))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	got := map[string]bool{}
	for _, f := range rules.Run(ins) {
		got[f.Rule] = true
	}
	return got
}

func TestRulesFire(t *testing.T) {
	bad := `
FROM node:latest
COPY . .
RUN npm install
RUN apt-get install -y curl
RUN pip install flask
`
	got := lint(t, bad)
	for _, id := range []string{"DIO001", "DIO002", "DIO003", "DIO004", "DIO005"} {
		if !got[id] {
			t.Errorf("expected rule %s to fire", id)
		}
	}
}

func TestCleanDockerfilePasses(t *testing.T) {
	good := `
FROM node:20-slim
COPY package*.json ./
RUN npm ci
RUN pip install --no-cache-dir flask
RUN apt-get install -y --no-install-recommends curl \
    && rm -rf /var/lib/apt/lists/*
COPY . .
`
	if got := lint(t, good); len(got) != 0 {
		t.Errorf("expected no findings, got %v", got)
	}
}
