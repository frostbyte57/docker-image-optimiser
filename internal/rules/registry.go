package rules

import "github.com/yuxiangchang/docker-image-optimiser/internal/parser"

// All returns every registered rule, in reporting order.
func All() []Rule {
	return []Rule{
		copyBeforeInstall{},
		aptNoRecommends{},
		aptCacheClean{},
		pipNoCache{},
		latestTag{},
	}
}

// Run executes every rule against the instructions and collects all findings.
func Run(ins []parser.Instruction) []Finding {
	var findings []Finding
	for _, r := range All() {
		findings = append(findings, r.Check(ins)...)
	}
	return findings
}
