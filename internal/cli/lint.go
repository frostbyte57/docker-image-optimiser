package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yuxiangchang/docker-image-optimiser/internal/parser"
	"github.com/yuxiangchang/docker-image-optimiser/internal/report"
	"github.com/yuxiangchang/docker-image-optimiser/internal/rules"
)

// newLintCmd implements `dio lint <Dockerfile>`: parse, run rules, report.
func newLintCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lint [Dockerfile]",
		Short: "Report size and build-speed anti-patterns in a Dockerfile",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "Dockerfile"
			if len(args) == 1 {
				path = args[0]
			}

			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			ins, err := parser.Parse(f)
			if err != nil {
				return fmt.Errorf("parsing %s: %w", path, err)
			}

			findings := rules.Run(ins)
			n := report.Text(cmd.OutOrStdout(), path, findings)

			// Non-zero exit on findings, so CI can gate on it.
			if n > 0 {
				return errSilent
			}
			return nil
		},
	}
}
