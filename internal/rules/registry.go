package rules

import "github.com/yuxiangchang/docker-image-optimiser/internal/parser"

// All returns every registered rule, in reporting order.
func All() []Rule {
	return []Rule{
		copyBeforeInstall{}, // DIO001
		aptNoRecommends{},   // DIO002
		systemCacheClean{},  // DIO003
		packageCache{},      // DIO004
		latestTag{},         // DIO005
		fatBase{},           // DIO006
		multiStage{},        // DIO007
		syntaxDirective{},   // DIO008
		dockerignore{},      // DIO009
		rootUser{},          // DIO010
	}
}

// Run executes every rule against the instructions and collects all findings.
func Run(ins []parser.Instruction, opts Options) []Finding {
	var findings []Finding
	for _, r := range All() {
		findings = append(findings, r.Check(ins, opts)...)
	}
	return findings
}
