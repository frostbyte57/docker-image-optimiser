package cli

import "fmt"

const (
	outputText   = "text"
	outputJSON   = "json"
	outputGitHub = "github"
)

func validateOutputFormat(format string) error {
	switch format {
	case outputText, outputJSON, outputGitHub:
		return nil
	default:
		return fmt.Errorf("unsupported output format %q (expected text, json, or github)", format)
	}
}
