package rules

import (
	"strings"

	"github.com/yuxiangchang/docker-image-optimiser/internal/parser"
)

// stage is one FROM section of a Dockerfile.
type stage struct {
	from  parser.Instruction
	image string // image reference, e.g. "golang:1.22"
	alias string // stage name from `AS <alias>`, if any
	body  []parser.Instruction
}

// stages splits instructions into their FROM sections.
func stages(ins []parser.Instruction) []stage {
	var out []stage
	for _, in := range ins {
		if in.Cmd == "FROM" {
			img, alias := imageRef(in.Args)
			out = append(out, stage{from: in, image: img, alias: alias})
			continue
		}
		if len(out) > 0 {
			last := &out[len(out)-1]
			last.body = append(last.body, in)
		}
	}
	return out
}

// imageRef parses a FROM argument into its image reference and stage alias.
func imageRef(args string) (image, alias string) {
	f := strings.Fields(args)
	if len(f) == 0 {
		return "", ""
	}
	image = f[0]
	if len(f) >= 3 && strings.EqualFold(f[1], "AS") {
		alias = f[2]
	}
	return image, alias
}

// baseName returns the image name without registry path, tag, or digest:
// "docker.io/library/python:3.12-slim" -> "python".
func baseName(ref string) string {
	name := ref
	if i := strings.IndexAny(name, "@:"); i >= 0 {
		name = name[:i]
	}
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	return name
}

// hasCopyFromStage reports whether any instruction copies artifacts from another
// build stage, i.e. the Dockerfile already uses a multi-stage pattern.
func hasCopyFromStage(ins []parser.Instruction) bool {
	for _, in := range ins {
		if in.Cmd == "COPY" && strings.Contains(in.Args, "--from=") {
			return true
		}
	}
	return false
}
