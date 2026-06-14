package report

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/yuxiangchang/docker-image-optimiser/internal/rules"
)

// GitHub writes findings as GitHub Actions workflow commands so they surface as
// inline annotations on pull requests. It returns the number of findings.
//
// See: https://docs.github.com/actions/using-workflows/workflow-commands-for-github-actions
func GitHub(w io.Writer, path string, findings []rules.Finding) int {
	sort.SliceStable(findings, func(i, j int) bool {
		return findings[i].Line < findings[j].Line
	})

	for _, f := range findings {
		msg := f.Message
		if f.Fix != "" {
			msg += " — fix: " + f.Fix
		}
		// file/line are properties; the title carries the rule id.
		fmt.Fprintf(w, "::%s file=%s,line=%d,title=%s::%s\n",
			ghLevel(f.Severity), escapeProperty(path), max(f.Line, 1),
			escapeProperty("dio "+f.Rule), escapeData(msg))
	}
	return len(findings)
}

// ghLevel maps a dio severity to a GitHub annotation level.
func ghLevel(s rules.Severity) string {
	switch s {
	case rules.Error:
		return "error"
	case rules.Warning:
		return "warning"
	default:
		return "notice"
	}
}

// escapeData escapes characters that are special inside workflow-command data.
func escapeData(s string) string {
	r := strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A")
	return r.Replace(s)
}

// escapeProperty escapes characters that are special inside command properties.
func escapeProperty(s string) string {
	r := strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A", ":", "%3A", ",", "%2C")
	return r.Replace(s)
}
