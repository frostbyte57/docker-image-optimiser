package rules

import (
	"github.com/yuxiangchang/docker-image-optimiser/internal/parser"
)

// multiStage (DIO007) flags a single-stage build whose base ships a full build
// toolchain (compiler/SDK), meaning that toolchain ends up in the final image.
// The fix — split into a build stage and a slim runtime stage — needs the user
// to wire up the artifact path, so this is annotate-only with a template.
type multiStage struct{}

func (multiStage) ID() string { return "DIO007" }

// buildToolchains maps a toolchain base image to a suggested multi-stage runtime.
var buildToolchains = map[string]string{
	"golang": "gcr.io/distroless/static (static binary)",
	"rust":   "gcr.io/distroless/cc or debian:slim",
	"maven":  "eclipse-temurin:<ver>-jre",
	"gradle": "eclipse-temurin:<ver>-jre",
}

func (r multiStage) Check(ins []parser.Instruction, _ Options) []Finding {
	if hasCopyFromStage(ins) {
		return nil // already multi-stage
	}
	var findings []Finding
	for _, st := range stages(ins) {
		runtime, isToolchain := buildToolchains[baseName(st.image)]
		if !isToolchain {
			continue
		}
		findings = append(findings, Finding{
			Rule:     r.ID(),
			Severity: Warning,
			Line:     st.from.StartLine,
			Message:  "single-stage build on " + st.image + " ships the build toolchain in the final image",
			Fix:      "Convert to multi-stage: build here, then COPY --from=build into " + runtime,
		})
	}
	return findings
}
