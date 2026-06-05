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
		if in.Cmd != "RUN" || isExecForm(in.Args) {
			continue
		}
		verb := aptInstallVerb(in.Args) // "apt-get install" or "apt install"
		if verb == "" {
			continue
		}
		if strings.Contains(in.Args, "--no-install-recommends") {
			continue
		}
		findings = append(findings, Finding{
			Rule:     r.ID(),
			Severity: Warning,
			Line:     in.StartLine,
			Message:  verb + " without --no-install-recommends pulls in extra packages",
			Fix:      "Add --no-install-recommends to the " + verb + " command",
			Rewrite: func(raw string) string {
				return strings.Replace(raw, verb, verb+" --no-install-recommends", 1)
			},
		})
	}
	return findings
}

// aptInstallVerb returns the apt install verb used in args, preferring the more
// specific "apt-get install" (so a line using both doesn't double-match). It
// returns "" when neither form is present. Both verbs are flagged because the
// ecosystem detects both, so DIO002 and DIO003 stay in lockstep.
func aptInstallVerb(args string) string {
	switch {
	case strings.Contains(args, "apt-get install"):
		return "apt-get install"
	case strings.Contains(args, "apt install"):
		return "apt install"
	default:
		return ""
	}
}
