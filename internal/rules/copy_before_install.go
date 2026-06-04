package rules

import (
	"strings"

	"github.com/yuxiangchang/docker-image-optimiser/internal/parser"
)

// copyBeforeInstall flags a broad `COPY . .` that appears before a dependency
// install in the same stage. That ordering busts the build cache on every
// source change, forcing a full reinstall. Copy manifests first instead.
type copyBeforeInstall struct{}

func (copyBeforeInstall) ID() string { return "DIO001" }

// installMarkers indicate a dependency install that reads from copied manifest
// files, so it benefits from copying those manifests before the rest of the
// source. System package managers (apt/apk) install from remote repos and are
// deliberately excluded.
var installMarkers = []string{
	"npm install", "npm ci", "yarn install", "pnpm install",
	"pip install", "poetry install",
	"go mod download", "bundle install", "composer install",
}

func (r copyBeforeInstall) Check(ins []parser.Instruction) []Finding {
	var findings []Finding
	for i, in := range ins {
		if in.Cmd != "COPY" || !isBroadCopy(in.Args) {
			continue
		}
		// Does a dependency install follow, before the stage ends?
		for _, later := range ins[i+1:] {
			if later.Cmd == "FROM" {
				break
			}
			if later.Cmd == "RUN" && containsAny(later.Args, installMarkers) {
				findings = append(findings, Finding{
					Rule:     r.ID(),
					Severity: Warning,
					Line:     in.StartLine,
					Message:  "`COPY . .` before installing dependencies busts the cache on every source change",
					Fix:      "Copy dependency manifests first (e.g. COPY package*.json ./), run the install, then COPY the rest",
				})
				break
			}
		}
	}
	return findings
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

func containsAny(s string, subs []string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
