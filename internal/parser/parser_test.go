package parser_test

import (
	"strings"
	"testing"

	"github.com/yuxiangchang/docker-image-optimiser/internal/parser"
)

func TestParseBuildKitHeredoc(t *testing.T) {
	src := `# syntax=docker/dockerfile:1
FROM alpine:3.20
RUN <<EOF
apk add curl
EOF
`

	ins, err := parser.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	if len(ins) != 2 {
		t.Fatalf("expected 2 instructions, got %d: %#v", len(ins), ins)
	}

	run := ins[1]
	if run.Cmd != "RUN" {
		t.Fatalf("expected RUN, got %q", run.Cmd)
	}
	if !run.HasHeredoc {
		t.Fatalf("expected RUN heredoc to be marked: %#v", run)
	}
	if run.StartLine != 3 || run.EndLine != 5 {
		t.Fatalf("unexpected heredoc range: %d-%d", run.StartLine, run.EndLine)
	}
}

func TestParseEscapeDirective(t *testing.T) {
	src := "# escape=`\nFROM mcr.microsoft.com/windows/nanoserver:ltsc2022\nRUN echo hello `\n    && echo world\n"

	ins, err := parser.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	if len(ins) != 2 {
		t.Fatalf("expected 2 instructions, got %d: %#v", len(ins), ins)
	}
	if got := ins[1].Raw; !strings.Contains(got, "echo hello") || strings.Contains(got, "`") {
		t.Fatalf("escape directive was not applied to continuation: %q", got)
	}
}

func TestParseOnlyComments(t *testing.T) {
	ins, err := parser.Parse(strings.NewReader("# syntax=docker/dockerfile:1\n\n# comment\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(ins) != 0 {
		t.Fatalf("expected no instructions, got %#v", ins)
	}
}
