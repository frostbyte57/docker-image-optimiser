package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yuxiangchang/docker-image-optimiser/internal/build"
	"github.com/yuxiangchang/docker-image-optimiser/internal/rewrite"
	"github.com/yuxiangchang/docker-image-optimiser/internal/rules"
)

// newBenchCmd implements `dio bench`: build the original Dockerfile and its
// auto-fixed rewrite, then report the size and build-time difference.
func newBenchCmd() *cobra.Command {
	var (
		contextDir  string
		keep        bool
		cache       bool
		incremental bool
	)

	cmd := &cobra.Command{
		Use:   "bench [Dockerfile]",
		Short: "Build the Dockerfile and its optimised rewrite, then compare them",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "Dockerfile"
			if len(args) == 1 {
				path = args[0]
			}

			if err := build.Available(); err != nil {
				return err
			}

			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			res, err := rewrite.Apply(src, rules.Options{})
			if err != nil {
				return err
			}
			if !res.Changed {
				fmt.Fprintln(cmd.ErrOrStderr(), path+": already optimal — nothing to compare")
				return nil
			}

			// The rewrite lives in a temp file; the build context stays the same.
			fixed, err := os.CreateTemp("", "dio-bench-*.Dockerfile")
			if err != nil {
				return err
			}
			defer os.Remove(fixed.Name())
			if _, err := fixed.WriteString(res.Content); err != nil {
				return err
			}
			fixed.Close()

			const beforeTag, afterTag = "dio-bench-before:latest", "dio-bench-after:latest"
			if !keep {
				defer build.Remove(beforeTag)
				defer build.Remove(afterTag)
			}

			out := cmd.ErrOrStderr()
			fmt.Fprintln(out, "building original...")
			before, err := build.Build(contextDir, path, beforeTag, !cache)
			if err != nil {
				return err
			}
			fmt.Fprintln(out, "building optimised...")
			after, err := build.Build(contextDir, fixed.Name(), afterTag, !cache)
			if err != nil {
				return err
			}

			if incremental {
				fmt.Fprintln(out, "measuring warm rebuilds (this builds each twice)...")
				if before.WarmRebuild, err = build.WarmRebuild(contextDir, path, beforeTag); err != nil {
					return err
				}
				if after.WarmRebuild, err = build.WarmRebuild(contextDir, fixed.Name(), afterTag); err != nil {
					return err
				}
			}

			fmt.Fprint(cmd.OutOrStdout(), build.Compare(before, after))
			return nil
		},
	}

	cmd.Flags().StringVarP(&contextDir, "context", "c", ".", "build context directory")
	cmd.Flags().BoolVar(&keep, "keep", false, "keep the built images instead of removing them")
	cmd.Flags().BoolVar(&cache, "cache", false, "allow the build cache (default: --no-cache for a fair comparison)")
	cmd.Flags().BoolVar(&incremental, "incremental", false, "also measure warm rebuild time after a source change")
	return cmd
}
