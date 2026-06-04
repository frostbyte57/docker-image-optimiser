// Package rewrite applies lint fixes to a Dockerfile.
//
// It reuses the rules engine to find issues, then either rewrites the offending
// instruction in place (for deterministic, safe fixes) or leaves an annotated
// `# dio[...]` comment above it (for fixes that need human judgement, such as
// reordering layers or choosing a version tag).
package rewrite

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/yuxiangchang/docker-image-optimiser/internal/parser"
	"github.com/yuxiangchang/docker-image-optimiser/internal/rules"
)

// autoFixable maps a rule id to the transform that rewrites an instruction's
// raw text. Rules absent here are handled with an annotation instead.
var autoFixable = map[string]func(raw string) string{
	"DIO002": func(raw string) string {
		return strings.Replace(raw, "apt-get install", "apt-get install --no-install-recommends", 1)
	},
	"DIO003": func(raw string) string {
		return strings.TrimRight(raw, " ") + " && rm -rf /var/lib/apt/lists/*"
	},
	"DIO004": func(raw string) string {
		return strings.Replace(raw, "pip install", "pip install --no-cache-dir", 1)
	},
}

// Result holds the rewritten file and a human-readable change log.
type Result struct {
	Content string   // the rewritten Dockerfile
	Applied []string // fixes written automatically
	Manual  []string // issues annotated for manual attention
	Changed bool
}

// Apply rewrites src and returns the result.
func Apply(src []byte) (Result, error) {
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
		fixedRaw = map[int]string{}      // startLine -> rewritten instruction
		comments = map[int][]string{}    // startLine -> annotations to prepend
	)

	lines := strings.Split(string(src), "\n")

	for _, f := range rules.Run(ins) {
		if fix, ok := autoFixable[f.Rule]; ok {
			raw := fixedRaw[f.Line]
			if raw == "" {
				raw = startIndex[f.Line].Raw
			}
			fixedRaw[f.Line] = fix(raw)
			res.Applied = append(res.Applied, summary(f))
		} else {
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

	res.Content = strings.Join(out, "\n")
	res.Changed = res.Content != string(src)
	return res, nil
}

// alreadyAnnotated reports whether the line just above startLine is the given
// annotation, so repeated `dio fix` runs don't stack duplicate comments.
func alreadyAnnotated(lines []string, startLine int, note string) bool {
	above := startLine - 1 // 1-based line above the instruction
	return above >= 1 && above <= len(lines) && strings.TrimSpace(lines[above-1]) == note
}

func summary(f rules.Finding) string {
	return "line " + strconv.Itoa(f.Line) + " [" + f.Rule + "] " + f.Message
}
