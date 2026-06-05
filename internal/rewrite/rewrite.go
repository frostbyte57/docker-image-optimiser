// Package rewrite applies lint fixes to a Dockerfile.
//
// It reuses the rules engine to find issues, then either rewrites the offending
// instruction in place (when the finding carries a Rewrite func) or leaves an
// annotated `# dio[...]` comment above it (for fixes that need human judgement,
// such as reordering layers, choosing a version tag, or going multi-stage).
package rewrite

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/yuxiangchang/docker-image-optimiser/internal/parser"
	"github.com/yuxiangchang/docker-image-optimiser/internal/rules"
)

// Result holds the rewritten file and a human-readable change log.
type Result struct {
	Content string   // the rewritten Dockerfile
	Applied []string // fixes written automatically
	Manual  []string // issues annotated for manual attention
	Changed bool
}

const syntaxDirective = "# syntax=docker/dockerfile:1"

// Apply rewrites src and returns the result. opts is threaded to the rules
// (e.g. opts.Conservative selects --no-cache-dir-style fixes over cache mounts).
func Apply(src []byte, opts rules.Options) (Result, error) {
	if opts.Source == "" {
		opts.Source = string(src)
	}

	ins, err := parser.Parse(bytes.NewReader(src))
	if err != nil {
		return Result{}, err
	}

	startIndex := make(map[int]parser.Instruction, len(ins))
	for _, in := range ins {
		startIndex[in.StartLine] = in
	}

	var (
		res      Result
		fixedRaw = map[int]string{}   // startLine -> rewritten instruction
		comments = map[int][]string{} // startLine -> annotations to prepend
	)

	lines := strings.Split(string(src), "\n")

	for _, f := range rules.Run(ins, opts) {
		switch {
		case f.Rewrite != nil && f.Line > 0:
			raw := fixedRaw[f.Line]
			if raw == "" {
				raw = startIndex[f.Line].Raw
			}
			fixedRaw[f.Line] = f.Rewrite(raw)
			res.Applied = append(res.Applied, summary(f))
		case f.Line == 0:
			// File-level finding. DIO008 (missing syntax directive) is auto-applied
			// by ensureSyntaxDirective below, so report it as a fix; others (e.g. a
			// missing .dockerignore) genuinely need a human.
			if f.Rule == "DIO008" {
				res.Applied = append(res.Applied, summary(f))
			} else {
				res.Manual = append(res.Manual, summary(f))
			}
		default:
			note := "# dio[" + f.Rule + "]: " + f.Fix
			if alreadyAnnotated(lines, f.Line, note) {
				continue // keep fix idempotent across repeated runs
			}
			comments[f.Line] = append(comments[f.Line], note)
			res.Manual = append(res.Manual, summary(f))
		}
	}

	var out []string
	for n := 1; n <= len(lines); {
		in, isStart := startIndex[n]
		if !isStart {
			out = append(out, lines[n-1])
			n++
			continue
		}
		out = append(out, comments[n]...) // annotations first (may be empty)
		if raw, ok := fixedRaw[n]; ok {
			out = append(out, raw)
		} else {
			out = append(out, lines[in.StartLine-1:in.EndLine]...)
		}
		n = in.EndLine + 1
	}

	content := strings.Join(out, "\n")
	content = ensureSyntaxDirective(content)

	res.Content = content
	res.Changed = res.Content != string(src)
	return res, nil
}

// ensureSyntaxDirective prepends the BuildKit syntax directive when the file
// uses cache mounts but lacks it, so injected mounts work on older Docker.
func ensureSyntaxDirective(content string) string {
	if !strings.Contains(content, "--mount=type=cache") || rules.HasSyntaxDirective(content) {
		return content
	}
	return syntaxDirective + "\n" + content
}

// alreadyAnnotated reports whether the given annotation already appears in the
// contiguous block of `# dio[...]` comments directly above startLine, so repeated
// `dio fix` runs don't stack duplicates (one instruction may carry several).
func alreadyAnnotated(lines []string, startLine int, note string) bool {
	for i := startLine - 2; i >= 0 && i < len(lines); i-- { // 0-based, walking up
		t := strings.TrimSpace(lines[i])
		if t == note {
			return true
		}
		if !strings.HasPrefix(t, "# dio[") {
			break // stop at the first non-annotation line
		}
	}
	return false
}

func summary(f rules.Finding) string {
	return "line " + strconv.Itoa(f.Line) + " [" + f.Rule + "] " + f.Message
}
