package cli

import "github.com/yuxiangchang/docker-image-optimiser/internal/rules"

type findingOutput struct {
	Rule     string         `json:"rule"`
	Severity rules.Severity `json:"severity"`
	Line     int            `json:"line"`
	Message  string         `json:"message"`
	Fix      string         `json:"fix"`
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
