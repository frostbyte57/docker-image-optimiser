package rules

import (
	"strings"

	"github.com/yuxiangchang/docker-image-optimiser/internal/ecosystem"
	"github.com/yuxiangchang/docker-image-optimiser/internal/parser"
)

// systemCacheClean (DIO003) flags a system package install (apt/apk/dnf) that
// leaves its cache in the image. Cleaning in a later layer does not shrink the
// earlier one, so the cleanup must run in the same RUN. System managers default
// to conservative cleanup (cache mounts for apt need extra, error-prone setup).
type systemCacheClean struct{}

func (systemCacheClean) ID() string { return "DIO003" }

func (r systemCacheClean) Check(ins []parser.Instruction, _ Options) []Finding {
	var findings []Finding
	for _, in := range ins {
		if !isShellRewritableRun(in) {
			continue
		}
		eco, ok := ecosystem.ForCommand(in.Args)
		if !ok || eco.Kind != ecosystem.System {
			continue
		}

		switch {
		case eco.CacheFlag != "": // apk: idiomatic --no-cache flag
			if strings.Contains(in.Args, eco.CacheFlag) {
				continue
			}
			flag := eco.CacheFlag
			verb := eco.Matched(in.Args) // the exact verb written, e.g. "apk add"
			findings = append(findings, Finding{
				Rule:     r.ID(),
				Severity: Warning,
				Line:     in.StartLine,
				Message:  eco.Name + " install caches packages in the image",
				Fix:      "Add " + flag + " to the " + eco.Name + " command",
				Rewrite: func(raw string) string {
					return strings.Replace(raw, verb, verb+" "+flag, 1)
				},
			})
		case eco.Cleanup != "": // apt/dnf: remove the package lists in the same RUN
			if strings.Contains(in.Args, cleanupMarker(eco)) {
				continue
			}
			cleanup := eco.Cleanup
			findings = append(findings, Finding{
				Rule:     r.ID(),
				Severity: Warning,
				Line:     in.StartLine,
				Message:  eco.Name + " package cache is left in the image, adding wasted size",
				Fix:      "End the RUN with: && " + cleanup,
				Rewrite: func(raw string) string {
					return strings.TrimRight(raw, " ") + " && " + cleanup
				},
			})
		}
	}
	return findings
}

// cleanupMarker returns a stable substring that indicates the cleanup is already
// present, so the rule is idempotent.
func cleanupMarker(e ecosystem.Ecosystem) string {
	switch e.Name {
	case "apt":
		return "/var/lib/apt/lists"
	case "dnf":
		return "/var/cache/dnf"
	}
	return e.Cleanup
}
