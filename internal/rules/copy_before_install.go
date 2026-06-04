package rules

import (
	"strings"

	"github.com/yuxiangchang/docker-image-optimiser/internal/ecosystem"
	"github.com/yuxiangchang/docker-image-optimiser/internal/parser"
)

// copyBeforeInstall (DIO001) flags a broad `COPY . .` that appears before a
// language dependency install in the same stage. That ordering busts the build
// cache on every source change, forcing a full reinstall. The fix is to copy
// the manifest files first, install, then copy the rest — which is structural,
// so this is annotate-only.
type copyBeforeInstall struct{}

func (copyBeforeInstall) ID() string { return "DIO001" }

func (r copyBeforeInstall) Check(ins []parser.Instruction, _ Options) []Finding {
	var findings []Finding
	for i, in := range ins {
		if in.Cmd != "COPY" || !isBroadCopy(in.Args) {
			continue
		}
		// Does a language dependency install follow, before the stage ends?
		for _, later := range ins[i+1:] {
			if later.Cmd == "FROM" {
				break
			}
			if later.Cmd != "RUN" {
				continue
			}
			eco, ok := ecosystem.ForCommand(later.Args)
			if !ok || eco.Kind != ecosystem.Language {
				continue
			}
			hint := "Copy " + manifestHint(eco) + " first, run the install, then COPY the rest"
			findings = append(findings, Finding{
				Rule:     r.ID(),
				Severity: Warning,
				Line:     in.StartLine,
				Message:  "`COPY . .` before the " + eco.Name + " install busts the cache on every source change",
				Fix:      hint,
			})
			break
		}
	}
	return findings
}

// manifestHint renders an ecosystem's manifest files for a suggestion.
func manifestHint(e ecosystem.Ecosystem) string {
	if len(e.Manifests) == 0 {
		return "the dependency manifest"
	}
	return strings.Join(e.Manifests, " and ")
}

// isBroadCopy reports whether a COPY pulls in the whole build context. Only the
// source arguments are considered; the final field is the destination (which is
// commonly "./" even for a narrow copy) and flags like --from are ignored.
func isBroadCopy(args string) bool {
	fields := strings.Fields(args)
	var srcs []string
	for _, f := range fields {
		if strings.HasPrefix(f, "--") {
			continue
		}
		srcs = append(srcs, f)
	}
	if len(srcs) < 2 {
		return false // need at least one source and a destination
	}
	for _, src := range srcs[:len(srcs)-1] { // drop destination
		if src == "." || src == "./" {
			return true
		}
	}
	return false
}
