package rules

import (
	"strings"

	"github.com/yuxiangchang/docker-image-optimiser/internal/parser"
)

// syntaxDirective (DIO008) flags a Dockerfile that already uses BuildKit cache
// mounts by hand but is missing the `# syntax=docker/dockerfile:1` directive on
// the first line, which is required for the mount syntax on older Docker. (When
// the rewriter injects cache mounts itself it prepends this directive
// automatically, so this rule targets hand-written mounts.)
//
// It works on the raw source via Options.Source because the parser strips
// comments, so a directive is invisible at the instruction level.
type syntaxDirective struct{}

func (syntaxDirective) ID() string { return "DIO008" }

func (r syntaxDirective) Check(_ []parser.Instruction, opts Options) []Finding {
	if opts.Source == "" {
		return nil
	}
	if !strings.Contains(opts.Source, "--mount=type=cache") || HasSyntaxDirective(opts.Source) {
		return nil
	}
	return []Finding{{
		Rule:     r.ID(),
		Severity: Warning,
		Line:     0, // file-level: prepended, not tied to an instruction
		Message:  "cache mounts are used but the `# syntax=docker/dockerfile:1` directive is missing",
		Fix:      "Add `# syntax=docker/dockerfile:1` as the first line",
	}}
}

// HasSyntaxDirective reports whether the first non-empty line of src is a
// Dockerfile syntax directive. Exported so the rewriter can reuse it.
func HasSyntaxDirective(src string) bool {
	for _, line := range strings.Split(src, "\n") {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		return strings.HasPrefix(t, "#") && strings.Contains(strings.ToLower(t), "syntax=")
	}
	return false
}
