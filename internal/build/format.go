package build

import (
	"fmt"
	"time"
)

// HumanBytes renders a byte count like "142.6 MB".
func HumanBytes(n int64) string {
	const unit = 1000
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "kMGT"[exp])
}

// Compare renders a before/after summary of two builds.
func Compare(before, after Result) string {
	sizeDelta := after.Size - before.Size
	timeDelta := after.Duration - before.Duration

	return fmt.Sprintf(`
                 before        after         change
  size      %12s  %12s  %s
  build     %12s  %12s  %s
`,
		HumanBytes(before.Size), HumanBytes(after.Size), signedSize(sizeDelta, before.Size),
		round(before.Duration), round(after.Duration), signedDur(timeDelta),
	)
}

func round(d time.Duration) string { return d.Round(100 * time.Millisecond).String() }

func signedSize(delta, base int64) string {
	pct := ""
	if base > 0 {
		pct = fmt.Sprintf(" (%+.1f%%)", float64(delta)/float64(base)*100)
	}
	sign := "+"
	if delta < 0 {
		sign, delta = "-", -delta
	}
	return sign + HumanBytes(delta) + pct
}

func signedDur(delta time.Duration) string {
	if delta < 0 {
		return "-" + round(-delta)
	}
	return "+" + round(delta)
}
