package cli

import "fmt"

const (
	outputText = "text"
	outputJSON = "json"
)

func validateOutputFormat(format string) error {
	switch format {
	case outputText, outputJSON:
		return nil
	default:
		return fmt.Errorf("unsupported output format %q (expected text or json)", format)
	}
}
