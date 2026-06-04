package cli

import (
	"bytes"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yuxiangchang/docker-image-optimiser/internal/parser"
	"github.com/yuxiangchang/docker-image-optimiser/internal/report"
	"github.com/yuxiangchang/docker-image-optimiser/internal/rules"
)

// newLintCmd implements `dio lint <Dockerfile>`: parse, run rules, report.
func newLintCmd() *cobra.Command {
	var contextDir string

	cmd := &cobra.Command{
		Use:   "lint [Dockerfile]",
		Short: "Report size and build-speed anti-patterns in a Dockerfile",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "Dockerfile"
			if len(args) == 1 {
				path = args[0]
			}

			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			ins, err := parser.Parse(bytes.NewReader(src))
			if err != nil {
				return fmt.Errorf("parsing %s: %w", path, err)
			}

			findings := rules.Run(ins, rules.Options{
				ContextDir: contextDir,
				Source:     string(src),
			})
			n := report.Text(cmd.OutOrStdout(), path, findings)

			// Non-zero exit on findings, so CI can gate on it.
			if n > 0 {
				return errSilent
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&contextDir, "context", "c", ".", "build context dir (enables the .dockerignore check)")
	return cmd
}
