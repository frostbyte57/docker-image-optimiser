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

// Compare renders a before/after summary of two builds. The warm-rebuild row is
// only shown when both builds measured it (bench --incremental).
func Compare(before, after Result) string {
	out := fmt.Sprintf(`
                 before        after         change
  size      %12s  %12s  %s
  cold      %12s  %12s  %s
`,
		HumanBytes(before.Size), HumanBytes(after.Size), signedSize(after.Size-before.Size, before.Size),
		round(before.Duration), round(after.Duration), signedDur(after.Duration-before.Duration),
	)
	if before.WarmRebuild > 0 && after.WarmRebuild > 0 {
		out += fmt.Sprintf("  warm      %12s  %12s  %s\n",
			round(before.WarmRebuild), round(after.WarmRebuild),
			signedDur(after.WarmRebuild-before.WarmRebuild))
	}
	return out
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
