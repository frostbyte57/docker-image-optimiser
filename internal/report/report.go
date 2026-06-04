// Package report renders lint findings for the terminal.
package report

import (
	"fmt"
	"io"
	"sort"

	"github.com/yuxiangchang/docker-image-optimiser/internal/rules"
)

// Text writes findings grouped by line, returning the number printed.
func Text(w io.Writer, path string, findings []rules.Finding) int {
	if len(findings) == 0 {
		fmt.Fprintf(w, "%s: no issues found ✓\n", path)
		return 0
	}

	sort.SliceStable(findings, func(i, j int) bool {
		return findings[i].Line < findings[j].Line
	})

	for _, f := range findings {
		fmt.Fprintf(w, "%s:%d  [%s] %s\n", path, f.Line, f.Severity, f.Rule)
		fmt.Fprintf(w, "    %s\n", f.Message)
		fmt.Fprintf(w, "    fix: %s\n\n", f.Fix)
	}
	fmt.Fprintf(w, "%d issue(s) found\n", len(findings))
	return len(findings)
}
