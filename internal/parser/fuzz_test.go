package parser_test

import (
	"strings"
	"testing"

	"github.com/yuxiangchang/docker-image-optimiser/internal/parser"
)

func FuzzParse(f *testing.F) {
	seeds := []string{
		"FROM alpine:3.20\nRUN apk add --no-cache curl\n",
		"# syntax=docker/dockerfile:1\nFROM node:20\nRUN --mount=type=cache,target=/root/.npm npm ci\n",
		"FROM alpine\nRUN <<EOF\napk add curl\nEOF\n",
		"# escape=`\nFROM base\nRUN echo hello `\n    && echo world\n",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, src string) {
		_, _ = parser.Parse(strings.NewReader(src))
	})
}
