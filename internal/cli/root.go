// Package cli wires up the dio command tree.
package cli

import "github.com/spf13/cobra"

// NewRootCmd builds the top-level `dio` command with all subcommands attached.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "dio",
		Short:         "dio optimises Dockerfiles for smaller images and faster builds",
		Long:          rootLong,
		Example:       rootExamples,
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
