package rules

import (
	"strings"

	"github.com/yuxiangchang/docker-image-optimiser/internal/parser"
)

// aptNoRecommends (DIO002) flags `apt-get install` without
// --no-install-recommends, which pulls in suggested packages and bloats the
// image. Auto-fixable.
type aptNoRecommends struct{}

func (aptNoRecommends) ID() string { return "DIO002" }

func (r aptNoRecommends) Check(ins []parser.Instruction, _ Options) []Finding {
	var findings []Finding
	for _, in := range ins {
		if in.Cmd != "RUN" || !strings.Contains(in.Args, "apt-get install") {
			continue
		}
		if strings.Contains(in.Args, "--no-install-recommends") {
			continue
		}
		findings = append(findings, Finding{
			Rule:     r.ID(),
			Severity: Warning,
			Line:     in.StartLine,
			Message:  "apt-get install without --no-install-recommends pulls in extra packages",
			Fix:      "Add --no-install-recommends to the apt-get install command",
			Rewrite: func(raw string) string {
				return strings.Replace(raw, "apt-get install", "apt-get install --no-install-recommends", 1)
			},
		})
	}
	return findings
}
