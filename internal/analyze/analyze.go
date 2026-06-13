// Package analyze runs Dockerfile source through the parser and rule engine.
package analyze

import (
	"bytes"

	"github.com/yuxiangchang/docker-image-optimiser/internal/parser"
	"github.com/yuxiangchang/docker-image-optimiser/internal/rules"
)

// Dockerfile parses src and returns every optimisation finding from the rule
// engine. If opts.Source is empty it is filled from src for source-aware rules.
func Dockerfile(src []byte, opts rules.Options) ([]rules.Finding, error) {
	if opts.Source == "" {
		opts.Source = string(src)
	}

	ins, err := parser.Parse(bytes.NewReader(src))
	if err != nil {
		return nil, err
	}
	return rules.Run(ins, opts), nil
}
