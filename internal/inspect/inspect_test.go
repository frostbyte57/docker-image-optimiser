package inspect

import (
	"strings"
	"testing"
)

func TestReport(t *testing.T) {
	layers := []Layer{
		{Size: 80_000_000, CreatedBy: "RUN pip install x"},
		{Size: 120_000_000, CreatedBy: "FROM python:3.12-slim"},
		{Size: 0, CreatedBy: `CMD ["python"]`},
	}
	var b strings.Builder
	Report(&b, "demo:latest", layers, 0)
	out := b.String()

	// Total and largest-first ordering.
	if !strings.Contains(out, "200.0 MB across 3 layers") {
		t.Errorf("missing/incorrect total:\n%s", out)
	}
	if i, j := strings.Index(out, "120.0 MB"), strings.Index(out, "80.0 MB"); i < 0 || j < 0 || i > j {
		t.Errorf("layers not sorted largest-first:\n%s", out)
	}
}

func TestParseSize(t *testing.T) {
	cases := map[string]int64{
		"0B":     0,
		"142MB":  142_000_000,
		"1.45GB": 1_450_000_000,
		"512kB":  512_000,
		"7B":     7,
	}
	for in, want := range cases {
		if got := parseSize(in); got != want {
			t.Errorf("parseSize(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestCleanCmd(t *testing.T) {
	cases := map[string]string{
		"/bin/sh -c #(nop)  CMD [\"node\"]":             `CMD ["node"]`,
		"/bin/sh -c apt-get update":                     "apt-get update",
		"RUN /bin/sh -c pip install x # buildkit":       "RUN /bin/sh -c pip install x",
		"COPY dir:abc /app # buildkit":                  "COPY dir:abc /app",
	}
	for in, want := range cases {
		if got := cleanCmd(in); got != want {
			t.Errorf("cleanCmd(%q) = %q, want %q", in, got, want)
		}
	}
}
