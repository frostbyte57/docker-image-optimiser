// Package rules holds the lint checks that detect Dockerfile anti-patterns.
//
// Each rule is small and self-contained: it implements Rule and is registered
// in registry.go. To add a check, drop a new file in this package and append
// it to All().
package rules

import "github.com/yuxiangchang/docker-image-optimiser/internal/parser"

// Severity ranks how much a finding matters.
type Severity string

const (
	Info    Severity = "info"
	Warning Severity = "warning"
	Error   Severity = "error"
)

// Finding is a single problem a rule detected.
type Finding struct {
	Rule     string   // stable rule id, e.g. "DIO001"
	Severity Severity
	Line     int      // 1-based line in the Dockerfile
	Message  string   // what is wrong
	Fix      string   // how to fix it
}

// Rule inspects parsed instructions and reports any problems it finds.
type Rule interface {
	ID() string
	Check(ins []parser.Instruction) []Finding
}
