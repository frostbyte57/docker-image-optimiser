package inspect

import (
	"fmt"
	"io"
	"sort"

	"github.com/yuxiangchang/docker-image-optimiser/internal/build"
)

// Report writes a per-layer breakdown sorted largest-first. If top > 0 only the
// top N layers are listed; the rest are summarised on a final line.
func Report(w io.Writer, image string, layers []Layer, top int) {
	total := Total(layers)

	sorted := make([]Layer, len(layers))
	copy(sorted, layers)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Size > sorted[j].Size })

	fmt.Fprintf(w, "%s — %s across %d layers\n\n", image, build.HumanBytes(total), len(layers))

	shown := sorted
	if top > 0 && top < len(sorted) {
		shown = sorted[:top]
	}

	for _, l := range shown {
		fmt.Fprintf(w, "%10s  %5s  %s\n", build.HumanBytes(l.Size), pct(l.Size, total), truncate(l.CreatedBy, 80))
	}

	if hidden := len(sorted) - len(shown); hidden > 0 {
		var rest int64
		for _, l := range sorted[len(shown):] {
			rest += l.Size
		}
		fmt.Fprintf(w, "%10s  %5s  (%d smaller layers)\n", build.HumanBytes(rest), pct(rest, total), hidden)
	}
}

func pct(part, total int64) string {
	if total == 0 {
		return "0%"
	}
	return fmt.Sprintf("%.0f%%", float64(part)/float64(total)*100)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
