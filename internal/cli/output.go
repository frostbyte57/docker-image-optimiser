package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/yuxiangchang/docker-image-optimiser/internal/rules"
)

const (
	outputText = "text"
	outputJSON = "json"
)

type findingOutput struct {
	Rule     string         `json:"rule"`
	Severity rules.Severity `json:"severity"`
	Line     int            `json:"line"`
	Message  string         `json:"message"`
	Fix      string         `json:"fix"`
}

func validateOutputFormat(format string) error {
	switch format {
	case outputText, outputJSON:
		return nil
	default:
		return fmt.Errorf("unsupported output format %q (expected text or json)", format)
	}
}

func findingOutputs(findings []rules.Finding) []findingOutput {
	out := make([]findingOutput, 0, len(findings))
	for _, f := range findings {
		out = append(out, findingOutput{
			Rule:     f.Rule,
			Severity: f.Severity,
			Line:     f.Line,
			Message:  f.Message,
			Fix:      f.Fix,
		})
	}
	return out
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
