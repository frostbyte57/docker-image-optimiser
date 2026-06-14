package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yuxiangchang/docker-image-optimiser/internal/analyze"
	"github.com/yuxiangchang/docker-image-optimiser/internal/report"
	"github.com/yuxiangchang/docker-image-optimiser/internal/rules"
)

// newLintCmd implements `dio lint <Dockerfile>`: parse, run rules, report.
func newLintCmd() *cobra.Command {
	var (
		contextDir string
		format     string
	)

	cmd := &cobra.Command{
		Use:   "lint [Dockerfile]",
		Short: "Report size and build-speed anti-patterns in a Dockerfile",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFormat(format); err != nil {
				return err
			}

			path := "Dockerfile"
			if len(args) == 1 {
				path = args[0]
			}

			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			findings, err := analyze.Dockerfile(src, rules.Options{
				ContextDir: contextDir,
				Source:     string(src),
			})
			if err != nil {
				return fmt.Errorf("parsing %s: %w", path, err)
			}

			var n int
			switch format {
			case outputJSON:
				if err := writeJSON(cmd.OutOrStdout(), findingOutputs(findings)); err != nil {
					return err
				}
				n = len(findings)
			case outputGitHub:
				n = report.GitHub(cmd.OutOrStdout(), path, findings)
			default:
				n = report.Text(cmd.OutOrStdout(), path, findings)
			}

			// Non-zero exit on findings, so CI can gate on it.
			if n > 0 {
				return errSilent
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&contextDir, "context", "c", ".", "build context dir (enables the .dockerignore check)")
	cmd.Flags().StringVar(&format, "format", outputText, "output format: text, json, or github (CI annotations)")
	return cmd
}
