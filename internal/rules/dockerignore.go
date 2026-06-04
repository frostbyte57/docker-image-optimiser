package rules

import (
	"os"
	"path/filepath"

	"github.com/yuxiangchang/docker-image-optimiser/internal/parser"
)

// dockerignore (DIO009) flags a build context with no .dockerignore file. Its
// absence sends junk (.git, node_modules, build output) into the build context,
// slowing builds and bloating COPY layers. Only runs when a context dir is
// provided (Options.ContextDir); needs the filesystem.
type dockerignore struct{}

func (dockerignore) ID() string { return "DIO009" }

func (r dockerignore) Check(_ []parser.Instruction, opts Options) []Finding {
	if opts.ContextDir == "" {
		return nil
	}
	if _, err := os.Stat(filepath.Join(opts.ContextDir, ".dockerignore")); err == nil {
		return nil
	}
	return []Finding{{
		Rule:     r.ID(),
		Severity: Info,
		Line:     0,
		Message:  "no .dockerignore in the build context — junk files inflate the context and COPY layers",
		Fix:      "Add a .dockerignore excluding .git, node_modules, build output, etc.",
	}}
}
