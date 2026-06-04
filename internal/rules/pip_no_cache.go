package rules

import (
	"strings"

	"github.com/yuxiangchang/docker-image-optimiser/internal/parser"
)

// pipNoCache flags `pip install` without --no-cache-dir, which leaves the pip
// download cache baked into the layer.
type pipNoCache struct{}

func (pipNoCache) ID() string { return "DIO004" }

func (r pipNoCache) Check(ins []parser.Instruction) []Finding {
	var findings []Finding
	for _, in := range ins {
		if in.Cmd != "RUN" || !strings.Contains(in.Args, "pip install") {
			continue
		}
		if strings.Contains(in.Args, "--no-cache-dir") {
			continue
		}
		findings = append(findings, Finding{
			Rule:     r.ID(),
			Severity: Warning,
			Line:     in.StartLine,
			Message:  "pip install without --no-cache-dir bakes the download cache into the layer",
			Fix:      "Add --no-cache-dir, or use a BuildKit cache mount for ~/.cache/pip",
		})
	}
	return findings
}
