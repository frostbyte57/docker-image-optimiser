// Command dio is the entrypoint for the docker-image-optimiser CLI.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/yuxiangchang/docker-image-optimiser/internal/cli"
)

func main() {
	err := cli.NewRootCmd().Execute()
	switch {
	case err == nil:
		return
	case errors.Is(err, cli.ErrFindings):
		os.Exit(1) // findings already printed to stdout
	default:
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
