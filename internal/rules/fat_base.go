package rules

import (
	"strings"

	"github.com/yuxiangchang/docker-image-optimiser/internal/parser"
)

// fatBase (DIO006) flags base images that have a much smaller official variant.
// Swapping the base is a judgement call (slim/alpine/distroless have trade-offs),
// so this is annotate-only.
type fatBase struct{}

func (fatBase) ID() string { return "DIO006" }

// slimmer maps a fat base image to a suggested smaller variant.
var slimmer = map[string]string{
	"python":  "python:<ver>-slim",
	"node":    "node:<ver>-slim",
	"ruby":    "ruby:<ver>-slim",
	"php":     "php:<ver>-fpm-alpine",
	"openjdk": "eclipse-temurin:<ver>-jre (runtime) or a -jdk-slim",
	"debian":  "debian:<ver>-slim",
	"ubuntu":  "debian:<ver>-slim or distroless",
	"gcc":     "a slim build stage + distroless/static runtime",
}

func (r fatBase) Check(ins []parser.Instruction, _ Options) []Finding {
	var findings []Finding
	for _, st := range stages(ins) {
		ref := st.image
		if strings.HasPrefix(ref, "$") {
			continue // ARG-based base, can't judge
		}
		suggestion, fat := slimmer[baseName(ref)]
		if !fat {
			continue
		}
		if isAlreadySlim(ref) {
			continue
		}
		findings = append(findings, Finding{
			Rule:     r.ID(),
			Severity: Info,
			Line:     st.from.StartLine,
			Message:  "base image " + ref + " is large; a smaller variant usually works",
			Fix:      "Consider " + suggestion,
		})
	}
	return findings
}

// isAlreadySlim reports whether a reference already targets a small variant.
func isAlreadySlim(ref string) bool {
	for _, s := range []string{"slim", "alpine", "distroless", "-jre"} {
		if strings.Contains(ref, s) {
			return true
		}
	}
	return false
}
