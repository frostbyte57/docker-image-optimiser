package rewrite

import (
	"strings"

	"github.com/yuxiangchang/docker-image-optimiser/internal/rules"
)

const syntaxDirective = "# syntax=docker/dockerfile:1"

// ensureSyntaxDirective prepends the BuildKit syntax directive when the file
// uses cache mounts but lacks it, so injected mounts work on older Docker.
func ensureSyntaxDirective(content string) string {
	if !strings.Contains(content, "--mount=type=cache") || rules.HasSyntaxDirective(content) {
		return content
	}
	return syntaxDirective + "\n" + content
}
