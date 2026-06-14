package report

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuxiangchang/docker-image-optimiser/internal/rules"
)

func TestGitHubAnnotations(t *testing.T) {
	findings := []rules.Finding{
		{Rule: "DIO004", Severity: rules.Warning, Line: 3, Message: "npm install re-downloads", Fix: "use a cache mount"},
		{Rule: "DIO006", Severity: rules.Info, Line: 1, Message: "base image is large", Fix: "use -slim"},
		{Rule: "DIO010", Severity: rules.Error, Line: 0, Message: "runs as root"},
	}

	var buf bytes.Buffer
	n := GitHub(&buf, "build/Dockerfile", findings)
	if n != 3 {
		t.Fatalf("expected 3 findings, got %d", n)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 annotation lines, got %d:\n%s", len(lines), buf.String())
	}

	// Sorted by raw line number: the file-level finding (line 0) sorts first
	// and clamps to line 1 in output, mapping Error -> error.
	if !strings.HasPrefix(lines[0], "::error file=build/Dockerfile,line=1,title=dio DIO010::") {
		t.Errorf("unexpected first line: %q", lines[0])
	}
	// No fix on this finding, so no trailing fix suffix.
	if strings.Contains(lines[0], "fix:") {
		t.Errorf("did not expect fix suffix: %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "::notice file=build/Dockerfile,line=1,title=dio DIO006::") {
		t.Errorf("unexpected second line: %q", lines[1])
	}
	if !strings.Contains(lines[1], "fix: use -slim") {
		t.Errorf("expected fix appended: %q", lines[1])
	}
	if !strings.HasPrefix(lines[2], "::warning file=build/Dockerfile,line=3,title=dio DIO004::") {
		t.Errorf("unexpected third line: %q", lines[2])
	}
}

func TestGitHubEscaping(t *testing.T) {
	var buf bytes.Buffer
	GitHub(&buf, "Dockerfile", []rules.Finding{
		{Rule: "DIO001", Severity: rules.Warning, Line: 2, Message: "line one\nline two", Fix: "do x"},
	})
	got := buf.String()
	if strings.Contains(got, "line one\nline two") {
		t.Errorf("raw newline should be escaped: %q", got)
	}
	if !strings.Contains(got, "line one%0Aline two") {
		t.Errorf("expected %%0A-escaped newline: %q", got)
	}
}
