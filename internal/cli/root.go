// Package cli wires up the dio command tree.
package cli

import "github.com/spf13/cobra"

// NewRootCmd builds the top-level `dio` command with all subcommands attached.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "dio",
		Short: "dio optimises Dockerfiles for smaller images and faster builds",
		Long: `docker-image-optimiser (dio) is a CLI for making Dockerfiles better
for CI/CD builds. It lints size and build-speed anti-patterns, applies safe
rewrites, reports manual fixes, and can benchmark the result.

Available workflows:
  dio lint       Report Dockerfile optimisation issues
  dio fix        Print or write safe Dockerfile rewrites
  dio optimize   CI-friendly optimise/check command
  dio bench      Compare original vs optimised image builds
  dio inspect    Show image layer sizes`,
		Example: `  dio --help
  dio lint Dockerfile
  dio fix --write Dockerfile
  dio optimize --check --format json Dockerfile
  dio bench --incremental Dockerfile
  dio inspect myimage:latest --top 10`,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.AddCommand(
		newLintCmd(),
		newFixCmd(),
		newOptimizeCmd(),
		newBenchCmd(),
		newInspectCmd(),
	)
	return root
}
