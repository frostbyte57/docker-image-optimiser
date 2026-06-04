// Package inspect breaks an existing image down by layer, so users can see
// where the bytes actually went. It reads `docker history` rather than diffing
// layer tarballs, which keeps it simple while answering "why is this so big?".
package inspect

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Layer is one entry from an image's history.
type Layer struct {
	Size      int64  // bytes
	CreatedBy string // the build step that produced the layer
}

// History returns the layers of an image, ordered as Docker reports them
// (most recent first).
func History(image string) ([]Layer, error) {
	out, err := exec.Command("docker", "history", "--no-trunc", "--format", "{{json .}}", image).Output()
	if err != nil {
		return nil, fmt.Errorf("docker history %s: %w", image, err)
	}

	var layers []Layer
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		var raw struct {
			Size      string `json:"Size"`
			CreatedBy string `json:"CreatedBy"`
		}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			return nil, fmt.Errorf("parsing docker history: %w", err)
		}
		layers = append(layers, Layer{
			Size:      parseSize(raw.Size),
			CreatedBy: cleanCmd(raw.CreatedBy),
		})
	}
	return layers, nil
}

// Total sums the size of all layers.
func Total(layers []Layer) int64 {
	var t int64
	for _, l := range layers {
		t += l.Size
	}
	return t
}

// sizeUnits maps the suffixes Docker prints (decimal, via go-units) to bytes.
var sizeUnits = []struct {
	suffix string
	factor float64
}{
	{"TB", 1e12}, {"GB", 1e9}, {"MB", 1e6}, {"kB", 1e3}, {"B", 1},
}

// parseSize converts a Docker size string like "142MB" or "0B" to bytes.
func parseSize(s string) int64 {
	s = strings.TrimSpace(s)
	for _, u := range sizeUnits {
		if strings.HasSuffix(s, u.suffix) {
			num := strings.TrimSpace(strings.TrimSuffix(s, u.suffix))
			v, err := strconv.ParseFloat(num, 64)
			if err != nil {
				return 0
			}
			return int64(v*u.factor + 0.5)
		}
	}
	return 0
}

// cleanCmd strips the shell/buildkit boilerplate from a history command so the
// meaningful part is readable.
func cleanCmd(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "/bin/sh -c #(nop) ")
	s = strings.TrimPrefix(s, "/bin/sh -c ")
	if i := strings.Index(s, "# buildkit"); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	return strings.Join(strings.Fields(s), " ") // collapse whitespace/newlines
}
