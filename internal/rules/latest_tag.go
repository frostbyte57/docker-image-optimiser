package rules

import (
	"strings"

	"github.com/yuxiangchang/docker-image-optimiser/internal/parser"
)

// latestTag flags base images pinned to :latest or with no tag at all, which
// makes builds non-reproducible.
type latestTag struct{}

func (latestTag) ID() string { return "DIO005" }

func (r latestTag) Check(ins []parser.Instruction, _ Options) []Finding {
	var findings []Finding
	for _, in := range ins {
		if in.Cmd != "FROM" {
			continue
		}
		image := strings.Fields(in.Args) // "image:tag [AS name]"
		if len(image) == 0 {
			continue
		}
		ref := image[0]
		if strings.HasPrefix(ref, "$") { // ARG-based, skip
			continue
		}
		tag := tagOf(ref)
		if tag == "" || tag == "latest" {
			findings = append(findings, Finding{
				Rule:     r.ID(),
				Severity: Info,
				Line:     in.StartLine,
				Message:  "base image uses :latest or no tag, so builds are not reproducible",
				Fix:      "Pin an explicit version tag, e.g. node:20-slim",
			})
		}
	}
	return findings
}

// tagOf returns the tag of an image reference, ignoring any digest or registry
// port. Returns "" when no tag is present.
func tagOf(ref string) string {
	// A digest reference (image@sha256:...) is considered pinned.
	if strings.Contains(ref, "@") {
		return "pinned"
	}
	slash := strings.LastIndex(ref, "/")
	colon := strings.LastIndex(ref, ":")
	if colon > slash { // colon after the last slash is a tag, not a registry port
		return ref[colon+1:]
	}
	return ""
}
