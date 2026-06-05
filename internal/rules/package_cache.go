package rules

import (
	"strings"

	"github.com/yuxiangchang/docker-image-optimiser/internal/ecosystem"
	"github.com/yuxiangchang/docker-image-optimiser/internal/parser"
)

// packageCache (DIO004) flags a language dependency install that neither uses a
// BuildKit cache mount nor (in conservative mode) strips its cache. By default
// it injects a cache mount, which keeps the download cache out of the image AND
// reuses it across builds — strictly better than --no-cache-dir, which only
// shrinks the image while forcing a re-download every build.
type packageCache struct{}

func (packageCache) ID() string { return "DIO004" }

func (r packageCache) Check(ins []parser.Instruction, opts Options) []Finding {
	var findings []Finding
	for _, in := range ins {
		if in.Cmd != "RUN" || isExecForm(in.Args) {
			continue
		}
		eco, ok := ecosystem.ForCommand(in.Args)
		if !ok || eco.Kind != ecosystem.Language || len(eco.CacheMounts) == 0 {
			continue
		}
		if strings.Contains(in.Raw, "--mount=type=cache") {
			continue // already using a cache mount
		}

		if opts.Conservative {
			f, ok := conservativeFix(r.ID(), in, eco)
			if ok {
				findings = append(findings, f)
			}
			continue
		}

		mounts := mountFlags(eco)
		findings = append(findings, Finding{
			Rule:     r.ID(),
			Severity: Warning,
			Line:     in.StartLine,
			Message:  eco.Name + " install re-downloads every build and bakes its cache into the image",
			Fix:      "Use a BuildKit cache mount: --mount=type=cache,target=" + eco.CacheMounts[0],
			Rewrite: func(raw string) string {
				return injectMounts(raw, mounts)
			},
		})
	}
	return findings
}

// conservativeFix builds a --no-cache-dir style finding when the ecosystem has a
// cache flag. Ecosystems without one are skipped in conservative mode (there is
// no safe, generic in-layer cleanup for them without BuildKit).
func conservativeFix(id string, in parser.Instruction, eco ecosystem.Ecosystem) (Finding, bool) {
	if eco.CacheFlag == "" {
		return Finding{}, false
	}
	if strings.Contains(in.Args, eco.CacheFlag) {
		return Finding{}, false
	}
	verb := eco.Matched(in.Args) // the exact verb written, e.g. "pip3 install"
	flag := eco.CacheFlag
	return Finding{
		Rule:     id,
		Severity: Warning,
		Line:     in.StartLine,
		Message:  eco.Name + " install bakes its download cache into the layer",
		Fix:      "Add " + flag + " to the " + eco.Name + " command",
		Rewrite: func(raw string) string {
			return strings.Replace(raw, verb, verb+" "+flag, 1)
		},
	}, true
}

// mountFlags renders the BuildKit cache mount flags for an ecosystem.
func mountFlags(e ecosystem.Ecosystem) string {
	var b strings.Builder
	for _, dir := range e.CacheMounts {
		b.WriteString("--mount=type=cache,target=")
		b.WriteString(dir)
		if e.Sharing != "" {
			b.WriteString(",sharing=")
			b.WriteString(e.Sharing)
		}
		b.WriteString(" ")
	}
	return b.String()
}

// injectMounts inserts cache-mount flags right after the RUN keyword.
func injectMounts(raw, mounts string) string {
	cmd, rest, found := strings.Cut(raw, " ")
	if !found {
		return raw
	}
	return cmd + " " + mounts + rest
}
