package rules

import (
	"github.com/yuxiangchang/docker-image-optimiser/internal/parser"
)

// rootUser (DIO010) flags an image whose final stage never drops privileges with
// a USER instruction, so the container runs as root. Choosing/creating the user
// is app-specific, so this is annotate-only.
type rootUser struct{}

func (rootUser) ID() string { return "DIO010" }

func (r rootUser) Check(ins []parser.Instruction, _ Options) []Finding {
	sts := stages(ins)
	if len(sts) == 0 {
		return nil
	}
	final := sts[len(sts)-1]

	for _, in := range final.body {
		if in.Cmd == "USER" {
			return nil // drops privileges somewhere in the final stage
		}
	}
	return []Finding{{
		Rule:     r.ID(),
		Severity: Info,
		Line:     final.from.StartLine,
		Message:  "the final stage runs as root; no USER instruction drops privileges",
		Fix:      "Create and switch to a non-root user, e.g. USER nonroot (or a distroless :nonroot base)",
	}}
}
