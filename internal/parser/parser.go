// Package parser turns a Dockerfile into a flat list of instructions.
//
// It delegates Dockerfile grammar handling to BuildKit's parser, then adapts
// the resulting AST to the smaller shape that dio's rules need.
package parser

import (
	"bytes"
	"io"
	"strings"

	buildkitparser "github.com/moby/buildkit/frontend/dockerfile/parser"
)

// Instruction is a single Dockerfile directive, e.g. `RUN apt-get update`.
type Instruction struct {
	Cmd        string // upper-cased command, e.g. "RUN", "COPY", "FROM"
	Args       string // everything after the command, continuations joined
	StartLine  int    // 1-based line where the instruction begins
	EndLine    int    // 1-based line where the instruction ends
	Raw        string // the parsed instruction as a single line
	HasHeredoc bool   // whether the instruction owns heredoc body content
}

// Parse reads a Dockerfile and returns its instructions in order.
func Parse(r io.Reader) ([]Instruction, error) {
	src, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if hasNoInstructions(src) {
		return nil, nil
	}

	parsed, err := buildkitparser.Parse(bytes.NewReader(src))
	if err != nil {
		return nil, err
	}

	var out []Instruction
	for _, node := range parsed.AST.Children {
		raw := node.Original
		cmd, args, _ := strings.Cut(strings.TrimSpace(raw), " ")
		out = append(out, Instruction{
			Cmd:        strings.ToUpper(cmd),
			Args:       strings.TrimSpace(args),
			StartLine:  node.StartLine,
			EndLine:    node.EndLine,
			Raw:        raw,
			HasHeredoc: len(node.Heredocs) > 0,
		})
	}
	return out, nil
}

func hasNoInstructions(src []byte) bool {
	for _, line := range strings.Split(string(src), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		return false
	}
	return true
}
