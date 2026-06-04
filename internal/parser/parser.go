// Package parser turns a Dockerfile into a flat list of instructions.
//
// It is intentionally small: it handles comments and line continuations,
// which is everything the lint rules need. It is not a full Dockerfile
// grammar (no heredocs, no parser directives).
package parser

import (
	"bufio"
	"io"
	"strings"
)

// Instruction is a single Dockerfile directive, e.g. `RUN apt-get update`.
type Instruction struct {
	Cmd       string // upper-cased command, e.g. "RUN", "COPY", "FROM"
	Args      string // everything after the command, continuations joined
	StartLine int    // 1-based line where the instruction begins
	Raw       string // the original source, including continuations
}

// Parse reads a Dockerfile and returns its instructions in order.
func Parse(r io.Reader) ([]Instruction, error) {
	var (
		out     []Instruction
		sc      = bufio.NewScanner(r)
		lineNo  = 0
		pending strings.Builder // accumulates a continued line
		startAt = 0
	)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	flush := func() {
		raw := pending.String()
		pending.Reset()
		if strings.TrimSpace(raw) == "" {
			return
		}
		cmd, args, _ := strings.Cut(strings.TrimSpace(raw), " ")
		out = append(out, Instruction{
			Cmd:       strings.ToUpper(cmd),
			Args:      strings.TrimSpace(args),
			StartLine: startAt,
			Raw:       raw,
		})
	}

	for sc.Scan() {
		lineNo++
		line := sc.Text()
		trimmed := strings.TrimSpace(line)

		// Skip blank lines and comments only when not mid-continuation.
		if pending.Len() == 0 {
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			startAt = lineNo
		}

		continues := strings.HasSuffix(strings.TrimRight(line, " \t"), "\\")
		if continues {
			// Drop the trailing backslash, keep a space as a joiner.
			line = strings.TrimRight(line, " \t")
			line = strings.TrimSuffix(line, "\\")
			pending.WriteString(line)
			pending.WriteString(" ")
			continue
		}

		pending.WriteString(line)
		flush()
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	flush() // file ending mid-continuation
	return out, nil
}
