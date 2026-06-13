package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/yuxiangchang/docker-image-optimiser/internal/parser"
	"github.com/yuxiangchang/docker-image-optimiser/internal/rewrite"
	"github.com/yuxiangchang/docker-image-optimiser/internal/rules"
)

type optimizeOutput struct {
	Path         string          `json:"path"`
	Changed      bool            `json:"changed"`
	IssueCount   int             `json:"issue_count"`
	AutoFixCount int             `json:"auto_fix_count"`
	ManualCount  int             `json:"manual_count"`
	Applied      []string        `json:"applied"`
	Manual       []string        `json:"manual"`
	Findings     []findingOutput `json:"findings"`
}

// newOptimizeCmd implements the CI-friendly optimizer workflow: evaluate a
// Dockerfile, apply safe rewrites, and optionally enforce that no changes are
// pending.
func newOptimizeCmd() *cobra.Command {
	var (
		write        bool
		check        bool
		conservative bool
		contextDir   string
		format       string
	)

	cmd := &cobra.Command{
		Use:     "optimize [Dockerfile]",
		Aliases: []string{"optimise"},
		Short:   "Optimise a Dockerfile for CI builds and report pending changes",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFormat(format); err != nil {
				return err
			}
			if check && write {
				return fmt.Errorf("--check and --write cannot be used together")
			}

			path := "Dockerfile"
			if len(args) == 1 {
				path = args[0]
			}

			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			opts := rules.Options{
				Conservative: conservative,
				ContextDir:   contextDir,
				Source:       string(src),
			}
			findings, err := lintSource(src, opts)
			if err != nil {
				return fmt.Errorf("parsing %s: %w", path, err)
			}
			res, err := rewrite.Apply(src, opts)
			if err != nil {
				return fmt.Errorf("rewriting %s: %w", path, err)
			}

			summary := optimizeOutput{
				Path:         path,
				Changed:      res.Changed,
				IssueCount:   len(findings),
				AutoFixCount: len(res.Applied),
				ManualCount:  len(res.Manual),
				Applied:      res.Applied,
				Manual:       res.Manual,
				Findings:     findingOutputs(findings),
			}

			if write && res.Changed {
				if err := os.WriteFile(path, []byte(res.Content), 0o644); err != nil {
					return err
				}
			}

			switch format {
			case outputJSON:
				if err := writeJSON(cmd.OutOrStdout(), summary); err != nil {
					return err
				}
			default:
				writeOptimizeText(cmd.OutOrStdout(), summary, write)
			}

			if check && (res.Changed || len(res.Manual) > 0) {
				return errSilent
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&write, "write", "w", false, "write optimised Dockerfile back in place")
	cmd.Flags().BoolVar(&check, "check", false, "exit non-zero when optimisations or manual fixes are pending")
	cmd.Flags().BoolVar(&conservative, "conservative", false, "use --no-cache-dir style cleanup instead of BuildKit cache mounts")
	cmd.Flags().StringVarP(&contextDir, "context", "c", ".", "build context dir (enables the .dockerignore check)")
	cmd.Flags().StringVar(&format, "format", outputText, "output format: text or json")
	return cmd
}

func lintSource(src []byte, opts rules.Options) ([]rules.Finding, error) {
	ins, err := parser.Parse(bytes.NewReader(src))
	if err != nil {
		return nil, err
	}
	return rules.Run(ins, opts), nil
}

func writeOptimizeText(w io.Writer, out optimizeOutput, wrote bool) {
	if out.IssueCount == 0 {
		fmt.Fprintf(w, "%s: no issues found\n", out.Path)
		return
	}

	fmt.Fprintf(w, "%s: %d issue(s), %d auto-fix(es), %d manual action(s)\n",
		out.Path, out.IssueCount, out.AutoFixCount, out.ManualCount)
	for _, a := range out.Applied {
		fmt.Fprintln(w, "fixed:    "+a)
	}
	for _, m := range out.Manual {
		fmt.Fprintln(w, "manual:   "+m)
	}

	switch {
	case wrote && out.Changed:
		fmt.Fprintf(w, "%s: wrote optimised Dockerfile\n", out.Path)
	case out.Changed:
		fmt.Fprintf(w, "%s: optimisations available; rerun with --write to update the file\n", out.Path)
	default:
		fmt.Fprintf(w, "%s: no automatic edits available; manual actions remain\n", out.Path)
	}
}
