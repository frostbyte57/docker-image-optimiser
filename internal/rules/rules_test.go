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
	for _, f := range rules.Run(ins, rules.Options{Source: dockerfile}) {
		got[f.Rule] = true
	}
	return got
}

// Each case lists rule ids that MUST fire for that (deliberately suboptimal)
// Dockerfile. It is a subset check, not exhaustive.
func TestRulesFireAcrossEcosystems(t *testing.T) {
	cases := []struct {
		name   string
		df     string
		expect []string
	}{
		{
			name: "node",
			df:   "FROM node:latest\nCOPY . .\nRUN npm install\n",
			// DIO001 copy-before-install, DIO004 npm no cache mount,
			// DIO005 latest, DIO006 fat base, DIO010 root.
			expect: []string{"DIO001", "DIO004", "DIO005", "DIO006", "DIO010"},
		},
		{
			name:   "python apt",
			df:     "FROM python:3.12\nRUN apt-get install -y curl\nRUN pip install flask\n",
			expect: []string{"DIO002", "DIO003", "DIO004", "DIO006", "DIO010"},
		},
		{
			name:   "go single stage",
			df:     "FROM golang:1.22\nWORKDIR /app\nCOPY . .\nRUN go build -o /app/server .\n",
			expect: []string{"DIO001", "DIO004", "DIO007", "DIO010"},
		},
		{
			name:   "rust single stage",
			df:     "FROM rust:1.77\nCOPY . .\nRUN cargo build --release\n",
			expect: []string{"DIO001", "DIO004", "DIO007", "DIO010"},
		},
		{
			name:   "java maven",
			df:     "FROM maven:3.9\nCOPY . .\nRUN mvn -B package\n",
			expect: []string{"DIO001", "DIO004", "DIO007", "DIO010"},
		},
		{
			name: "ruby",
			df:   "FROM ruby:3.3\nCOPY . .\nRUN bundle install\n",
			// No DIO004: bundler installs into BUNDLE_PATH, so no cache mount.
			expect: []string{"DIO001", "DIO006", "DIO010"},
		},
		{
			name:   "alpine apk",
			df:     "FROM python:3.12-alpine\nRUN apk add curl\nRUN pip install flask\n",
			expect: []string{"DIO003", "DIO004", "DIO010"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := lint(t, tc.df)
			for _, id := range tc.expect {
				if !got[id] {
					t.Errorf("%s: expected %s to fire, got %v", tc.name, id, keys(got))
				}
			}
		})
	}
}

func TestCleanDockerfilePasses(t *testing.T) {
	good := `# syntax=docker/dockerfile:1
FROM node:20-slim
WORKDIR /app
COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci
COPY . .
USER node
`
	if got := lint(t, good); len(got) != 0 {
		t.Errorf("expected no findings, got %v", keys(got))
	}
}

func keys(m map[string]bool) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
}
