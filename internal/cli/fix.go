package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yuxiangchang/docker-image-optimiser/internal/rewrite"
)

// newFixCmd implements `dio fix <Dockerfile>`: rewrite safe issues in place and
// annotate the rest. By default it prints to stdout; -w writes back to the file.
func newFixCmd() *cobra.Command {
	var write bool

	cmd := &cobra.Command{
		Use:   "fix [Dockerfile]",
		Short: "Rewrite a Dockerfile, applying safe fixes and annotating the rest",
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

			res, err := rewrite.Apply(src)
			if err != nil {
				return fmt.Errorf("rewriting %s: %w", path, err)
			}

			// The change log goes to stderr so stdout stays a clean Dockerfile.
			for _, a := range res.Applied {
				fmt.Fprintln(cmd.ErrOrStderr(), "fixed:    "+a)
			}
			for _, m := range res.Manual {
				fmt.Fprintln(cmd.ErrOrStderr(), "annotated:"+m)
			}

			if !write {
				fmt.Fprint(cmd.OutOrStdout(), res.Content)
				return nil
			}

			if !res.Changed {
				fmt.Fprintln(cmd.ErrOrStderr(), path+": already optimal ✓")
				return nil
			}
			if err := os.WriteFile(path, []byte(res.Content), 0o644); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "%s: wrote %d fix(es), %d annotation(s)\n",
				path, len(res.Applied), len(res.Manual))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&write, "write", "w", false, "write changes back to the file in place")
	return cmd
}
