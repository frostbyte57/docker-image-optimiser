package cli

import (
	"errors"

	"github.com/spf13/cobra"
)

// ErrFindings signals that lint found issues. main exits non-zero on it but
// prints no extra error text (the findings are already on stdout).
var ErrFindings = errors.New("lint findings")

// errSilent is the internal alias returned by commands.
var errSilent = ErrFindings

// The commands below are placeholders for the remaining milestones. They keep
// the command tree visible while the heavier machinery is built out.

func newBenchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bench [Dockerfile]",
		Short: "Build before/after and compare size and build time (not yet implemented)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("bench: not yet implemented — see internal/build")
		},
	}
}

func newInspectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect [image]",
		Short: "Show per-layer sizes and wasted space for an image (not yet implemented)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("inspect: not yet implemented — see internal/inspect")
		},
	}
}
