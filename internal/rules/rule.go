// Package rules holds the lint checks that detect Dockerfile anti-patterns.
//
// Each rule is small and self-contained: it implements Rule and is registered
// in registry.go. To add a check, drop a new file in this package and append
// it to All(). Rules are mostly driven by the ecosystem registry, so supporting
// a new language is usually a table entry rather than a new rule.
package rules

import "github.com/yuxiangchang/docker-image-optimiser/internal/parser"

// Severity ranks how much a finding matters.
type Severity string

const (
	Info    Severity = "info"
	Warning Severity = "warning"
	Error   Severity = "error"
)

// Options controls how rules behave.
type Options struct {
	// Conservative makes cache-related rules prefer image-size cleanup
	// (--no-cache-dir, rm caches) over BuildKit cache mounts, for environments
	// without BuildKit.
	Conservative bool
	// ContextDir is the Docker build context. When non-empty, rules that need to
	// look at the filesystem (e.g. the .dockerignore check) are enabled.
	ContextDir string
	// Source is the raw Dockerfile text, for rules that must inspect comments or
	// the first line (the parser strips comments). May be empty in unit tests.
	Source string
}

// Finding is a single problem a rule detected.
type Finding struct {
	Rule     string // stable rule id, e.g. "DIO001"
	Severity Severity
	Line     int    // 1-based line in the Dockerfile (0 = file-level)
	Message  string // what is wrong
	Fix      string // how to fix it

	// Rewrite, if non-nil, transforms the offending instruction's raw text into
	// its fixed form. When nil the finding is annotate-only (the rewriter leaves
	// a `# dio[...]` comment instead of editing the instruction).
	Rewrite func(raw string) string
}

// Rule inspects parsed instructions and reports any problems it finds.
type Rule interface {
	ID() string
	Check(ins []parser.Instruction, opts Options) []Finding
}
