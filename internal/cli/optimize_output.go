package cli

import (
	"fmt"
	"io"
)

type optimizeOutput struct {
	Path         string          `json:"path"`
	Changed      bool            `json:"changed"`
	IssueCount   int             `json:"issue_count"`
	AutoFixCount int             `json:"auto_fix_count"`
	ManualCount  int             `json:"manual_count"`
	Applied      []string        `json:"applied"`
	Manual       []string        `json:"manual"`
	Findings     []findingOutput `json:"findings"`
}

func writeOptimizeText(w io.Writer, out optimizeOutput, wrote bool) {
	if out.IssueCount == 0 {
		fmt.Fprintf(w, "%s: no issues found\n", out.Path)
		return
	}

	fmt.Fprintf(w, "%s: %d issue(s), %d auto-fix(es), %d manual action(s)\n",
		out.Path, out.IssueCount, out.AutoFixCount, out.ManualCount)
	for _, a := range out.Applied {
		fmt.Fprintln(w, "fixed:    "+a)
	}
	for _, m := range out.Manual {
		fmt.Fprintln(w, "manual:   "+m)
	}

	switch {
	case wrote && out.Changed:
		fmt.Fprintf(w, "%s: wrote optimised Dockerfile\n", out.Path)
	case out.Changed:
		fmt.Fprintf(w, "%s: optimisations available; rerun with --write to update the file\n", out.Path)
	default:
		fmt.Fprintf(w, "%s: no automatic edits available; manual actions remain\n", out.Path)
	}
}
