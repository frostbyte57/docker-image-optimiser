package analyze_test

import (
	"testing"

	"github.com/yuxiangchang/docker-image-optimiser/internal/analyze"
	"github.com/yuxiangchang/docker-image-optimiser/internal/rules"
)

func TestDockerfileRunsRules(t *testing.T) {
	src := []byte("FROM node:latest\nRUN npm ci\n")

	findings, err := analyze.Dockerfile(src, rules.Options{})
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}

	for _, f := range findings {
		if f.Rule == "DIO004" {
			return
		}
	}
	t.Fatalf("expected DIO004 finding, got %+v", findings)
}
