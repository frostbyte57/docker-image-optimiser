package inspect

import (
	"strings"
	"testing"
	"unicode/utf8"
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

// truncate must cut on a rune boundary so a multibyte command never becomes
// invalid UTF-8 in the report.
func TestTruncate(t *testing.T) {
	// Short strings pass through untouched.
	if got := truncate("abc", 5); got != "abc" {
		t.Errorf("truncate(\"abc\", 5) = %q, want \"abc\"", got)
	}
	// 5 multibyte runes truncated to 3 → 2 runes kept + the ellipsis, still valid.
	long := "αβγδε"
	got := truncate(long, 3)
	if !utf8.ValidString(got) {
		t.Errorf("truncate produced invalid UTF-8: %q", got)
	}
	if n := utf8.RuneCountInString(got); n != 3 {
		t.Errorf("truncate(%q, 3) = %q (%d runes), want 3", long, got, n)
	}
	if !strings.HasSuffix(got, "…") {
		t.Errorf("truncate(%q, 3) = %q, want a trailing ellipsis", long, got)
	}
}

func TestCleanCmd(t *testing.T) {
	cases := map[string]string{
		"/bin/sh -c #(nop)  CMD [\"node\"]":       `CMD ["node"]`,
		"/bin/sh -c apt-get update":               "apt-get update",
		"RUN /bin/sh -c pip install x # buildkit": "RUN /bin/sh -c pip install x",
		"COPY dir:abc /app # buildkit":            "COPY dir:abc /app",
	}
	for in, want := range cases {
		if got := cleanCmd(in); got != want {
			t.Errorf("cleanCmd(%q) = %q, want %q", in, got, want)
		}
	}
}
