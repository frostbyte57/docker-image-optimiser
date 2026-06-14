package rewrite_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuxiangchang/docker-image-optimiser/internal/rewrite"
	"github.com/yuxiangchang/docker-image-optimiser/internal/rules"
)

func TestGoldenRewrites(t *testing.T) {
	cases := []struct {
		name string
		opts rules.Options
	}{
		{name: "python-default"},
		{name: "apt-conservative", opts: rules.Options{Conservative: true}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := readFixture(t, tc.name+".in.Dockerfile")
			want := readFixture(t, tc.name+".golden.Dockerfile")

			got, err := rewrite.Apply(input, tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			if trimFinalNewlines(got.Content) != trimFinalNewlines(string(want)) {
				t.Fatalf("rewrite mismatch\n--- want ---\n%s\n--- got ---\n%s", want, got.Content)
			}
		})
	}
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return []byte(strings.ReplaceAll(string(b), "\r\n", "\n"))
}

func trimFinalNewlines(s string) string {
	return strings.TrimRight(s, "\n")
}
