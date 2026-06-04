package build

import (
	"strings"
	"testing"
	"time"
)

func TestHumanBytes(t *testing.T) {
	cases := map[int64]string{
		512:        "512 B",
		1500:       "1.5 kB",
		142_600_000: "142.6 MB",
	}
	for in, want := range cases {
		if got := HumanBytes(in); got != want {
			t.Errorf("HumanBytes(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestCompareShowsShrinkAndSpeedup(t *testing.T) {
	before := Result{Size: 900_000_000, Duration: 60 * time.Second}
	after := Result{Size: 120_000_000, Duration: 42 * time.Second}

	out := Compare(before, after)
	if !strings.Contains(out, "-780.0 MB") {
		t.Errorf("expected size reduction in output:\n%s", out)
	}
	if !strings.Contains(out, "-18s") {
		t.Errorf("expected time reduction in output:\n%s", out)
	}
}
