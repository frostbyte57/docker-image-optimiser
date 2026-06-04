package cli

import (
	"github.com/spf13/cobra"

	"github.com/yuxiangchang/docker-image-optimiser/internal/build"
	"github.com/yuxiangchang/docker-image-optimiser/internal/inspect"
)

// newInspectCmd implements `dio inspect <image>`: show where an image's bytes
// went, layer by layer.
func newInspectCmd() *cobra.Command {
	var top int

	cmd := &cobra.Command{
		Use:   "inspect <image>",
		Short: "Show per-layer sizes for an image, largest first",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := build.Available(); err != nil {
				return err
			}
			layers, err := inspect.History(args[0])
			if err != nil {
				return err
			}
			inspect.Report(cmd.OutOrStdout(), args[0], layers, top)
			return nil
		},
	}

	cmd.Flags().IntVarP(&top, "top", "n", 0, "show only the N largest layers (0 = all)")
	return cmd
}
