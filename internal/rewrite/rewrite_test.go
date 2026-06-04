package rewrite_test

import (
	"strings"
	"testing"

	"github.com/yuxiangchang/docker-image-optimiser/internal/parser"
	"github.com/yuxiangchang/docker-image-optimiser/internal/rewrite"
	"github.com/yuxiangchang/docker-image-optimiser/internal/rules"
)

func TestApplyAutoFixes(t *testing.T) {
	src := []byte(`FROM node:latest
COPY . .
RUN npm install
RUN apt-get update && apt-get install -y curl
RUN pip install requests
`)

	res, err := rewrite.Apply(src)
	if err != nil {
		t.Fatal(err)
	}

	// Safe fixes are applied to the text.
	for _, want := range []string{
		"apt-get install --no-install-recommends",
		"rm -rf /var/lib/apt/lists/*",
		"pip install --no-cache-dir",
	} {
		if !strings.Contains(res.Content, want) {
			t.Errorf("expected rewritten content to contain %q\n---\n%s", want, res.Content)
		}
	}

	// Judgement calls are annotated, not silently changed.
	if !strings.Contains(res.Content, "# dio[DIO001]") || !strings.Contains(res.Content, "# dio[DIO005]") {
		t.Errorf("expected DIO001/DIO005 annotations\n---\n%s", res.Content)
	}
}

// TestRewriteIsClean verifies that linting the rewritten file no longer reports
// the auto-fixable rules.
func TestRewriteIsClean(t *testing.T) {
	src := []byte("FROM node:20-slim\nRUN apt-get install -y curl\nRUN pip install requests\n")

	res, err := rewrite.Apply(src)
	if err != nil {
		t.Fatal(err)
	}
	ins, err := parser.Parse(strings.NewReader(res.Content))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range rules.Run(ins) {
		if f.Rule == "DIO002" || f.Rule == "DIO003" || f.Rule == "DIO004" {
			t.Errorf("rule %s still fires after rewrite:\n%s", f.Rule, res.Content)
		}
	}
}
