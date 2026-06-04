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

// The command below is a placeholder for the remaining milestone. It keeps the
// command tree visible while the heavier machinery is built out.

func newInspectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect [image]",
		Short: "Show per-layer sizes and wasted space for an image (not yet implemented)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("inspect: not yet implemented — see internal/inspect")
		},
	}
}
