package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootHelpListsCommandsAndExamples(t *testing.T) {
	var stdout bytes.Buffer
	cmd := NewRootCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute help: %v", err)
	}

	help := stdout.String()
	for _, want := range []string{
		"Available workflows:",
		"dio lint",
		"dio fix",
		"dio optimize",
		"dio bench",
		"dio inspect",
		"dio optimize --check --format github Dockerfile",
		"Available Commands:",
	} {
		if !strings.Contains(help, want) {
			t.Fatalf("help output missing %q:\n%s", want, help)
		}
	}
}
