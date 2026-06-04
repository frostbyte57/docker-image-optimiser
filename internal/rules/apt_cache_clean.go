package rules

import (
	"strings"

	"github.com/yuxiangchang/docker-image-optimiser/internal/parser"
)

// aptCacheClean flags `apt-get install` that does not remove the apt lists in
// the same RUN. Cleaning in a later layer does not shrink the earlier one.
type aptCacheClean struct{}

func (aptCacheClean) ID() string { return "DIO003" }

func (r aptCacheClean) Check(ins []parser.Instruction) []Finding {
	var findings []Finding
	for _, in := range ins {
		if in.Cmd != "RUN" || !strings.Contains(in.Args, "apt-get install") {
			continue
		}
		if strings.Contains(in.Args, "/var/lib/apt/lists") {
			continue
		}
		findings = append(findings, Finding{
			Rule:     r.ID(),
			Severity: Warning,
			Line:     in.StartLine,
			Message:  "apt package lists are left in the image, adding wasted size",
			Fix:      "End the RUN with: && rm -rf /var/lib/apt/lists/*",
		})
	}
	return findings
}
