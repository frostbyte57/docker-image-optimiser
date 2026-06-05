// Package ecosystem is a data-driven registry of package managers.
//
// Every generic lint rule is driven by this table, so supporting a new language
// is a single entry here rather than a new rule. Each ecosystem knows how to be
// detected in a RUN command, which manifest files should be copied first for
// layer-cache ordering, where its download cache lives (for BuildKit cache
// mounts), and how to clean up conservatively when cache mounts are not used.
package ecosystem

import "strings"

// Kind distinguishes language package managers (where cache mounts are a safe,
// clear win) from system package managers (apt/apk/dnf), where cache mounts
// need extra setup and conservative cleanup is the safer default.
type Kind int

const (
	Language Kind = iota
	System
)

// Ecosystem describes one package manager.
type Ecosystem struct {
	Name        string   // "pip", "npm", ...
	Kind        Kind     // Language or System
	Detect      []string // command substrings that identify an install step
	Manifests   []string // files to COPY before install for layer-cache ordering
	CacheMounts []string // BuildKit cache target directories
	Sharing     string   // cache sharing mode, "" or "locked"
	CacheFlag   string   // conservative size flag, e.g. "--no-cache-dir" ("" if none)
	Cleanup     string   // conservative cleanup appended to the RUN ("" if none)
}

// registry is the full set of supported package managers. Order matters: the
// first Detect match wins, so more specific managers come before generic ones
// (e.g. "pip install" entries are distinct, but apt before apk etc.).
var registry = []Ecosystem{
	// --- Node ---
	{
		Name: "npm", Kind: Language,
		Detect:      []string{"npm install", "npm ci", "npm i "},
		Manifests:   []string{"package.json", "package-lock.json"},
		CacheMounts: []string{"/root/.npm"},
	},
	{
		Name: "pnpm", Kind: Language,
		Detect:      []string{"pnpm install", "pnpm i "},
		Manifests:   []string{"package.json", "pnpm-lock.yaml"},
		CacheMounts: []string{"/root/.local/share/pnpm/store"},
	},
	{
		Name: "yarn", Kind: Language,
		Detect:      []string{"yarn install", "yarn add", "yarn "},
		Manifests:   []string{"package.json", "yarn.lock"},
		CacheMounts: []string{"/usr/local/share/.cache/yarn"},
	},
	// --- Python --- (uv before pip: "uv pip install" contains "pip install")
	{
		Name: "uv", Kind: Language,
		Detect:      []string{"uv pip install", "uv sync"},
		Manifests:   []string{"pyproject.toml", "uv.lock"},
		CacheMounts: []string{"/root/.cache/uv"},
	},
	{
		Name: "poetry", Kind: Language,
		Detect:      []string{"poetry install"},
		Manifests:   []string{"pyproject.toml", "poetry.lock"},
		CacheMounts: []string{"/root/.cache/pypoetry"},
	},
	{
		Name: "pip", Kind: Language,
		Detect:      []string{"pip install", "pip3 install"},
		Manifests:   []string{"requirements.txt"},
		CacheMounts: []string{"/root/.cache/pip"},
		CacheFlag:   "--no-cache-dir",
	},
	// --- Go ---
	{
		Name: "go", Kind: Language,
		Detect:      []string{"go mod download", "go build", "go install", "go test"},
		Manifests:   []string{"go.mod", "go.sum"},
		CacheMounts: []string{"/go/pkg/mod", "/root/.cache/go-build"},
	},
	// --- Rust ---
	{
		Name: "cargo", Kind: Language,
		Detect:    []string{"cargo build", "cargo install", "cargo fetch"},
		Manifests: []string{"Cargo.toml", "Cargo.lock"},
		// Only absolute CARGO_HOME paths; the build target/ dir is workdir-relative
		// so we don't mount it blindly.
		CacheMounts: []string{"/usr/local/cargo/registry", "/usr/local/cargo/git"},
	},
	// --- Java ---
	{
		Name: "maven", Kind: Language,
		Detect:      []string{"mvn "},
		Manifests:   []string{"pom.xml"},
		CacheMounts: []string{"/root/.m2/repository"},
	},
	{
		Name: "gradle", Kind: Language,
		Detect:      []string{"gradle ", "./gradlew"},
		Manifests:   []string{"build.gradle", "settings.gradle", "build.gradle.kts"},
		CacheMounts: []string{"/root/.gradle"},
	},
	// --- PHP ---
	{
		Name: "composer", Kind: Language,
		Detect:      []string{"composer install", "composer require"},
		Manifests:   []string{"composer.json", "composer.lock"},
		CacheMounts: []string{"/root/.composer/cache"},
	},
	// --- Ruby ---
	// No cache mount: bundler installs gems directly into BUNDLE_PATH
	// (/usr/local/bundle), so mounting it as a cache would keep the gems out of
	// the image. It still benefits from manifest-first layer ordering (DIO001).
	{
		Name: "bundler", Kind: Language,
		Detect:    []string{"bundle install", "gem install"},
		Manifests: []string{"Gemfile", "Gemfile.lock"},
	},
	// --- .NET ---
	{
		Name: "dotnet", Kind: Language,
		Detect:      []string{"dotnet restore", "dotnet build", "dotnet publish"},
		Manifests:   []string{"*.csproj", "*.sln"},
		CacheMounts: []string{"/root/.nuget/packages"},
	},
	// --- System package managers (conservative cleanup, not cache mounts) ---
	{
		Name: "apt", Kind: System,
		Detect:      []string{"apt-get install", "apt install"},
		CacheMounts: []string{"/var/cache/apt"},
		Sharing:     "locked",
		Cleanup:     "rm -rf /var/lib/apt/lists/*",
	},
	{
		Name: "apk", Kind: System,
		Detect:    []string{"apk add"},
		CacheFlag: "--no-cache",
	},
	{
		Name: "dnf", Kind: System,
		Detect:  []string{"dnf install", "yum install"},
		Cleanup: "rm -rf /var/cache/dnf /var/cache/yum",
	},
}

// ForCommand returns the ecosystem whose install command appears in runArgs.
// The first registry match wins. Matching respects a left word boundary so that
// e.g. "cargo build" does not match "go build" and "pnpm install" does not match
// "npm install".
func ForCommand(runArgs string) (Ecosystem, bool) {
	for _, e := range registry {
		if e.Matched(runArgs) != "" {
			return e, true
		}
	}
	return Ecosystem{}, false
}

// Matched returns the first of e's Detect phrases that appears in runArgs (with a
// left word boundary), or "" if none do. Rewrites use this to target the exact
// verb the user wrote — e.g. "pip3 install" rather than the registry's canonical
// "pip install" — so the edit is never a silent no-op.
func (e Ecosystem) Matched(runArgs string) string {
	for _, d := range e.Detect {
		if containsWord(runArgs, d) {
			return d
		}
	}
	return ""
}

// containsWord reports whether phrase occurs in s preceded by the start of the
// string or a non-alphanumeric byte (a left word boundary).
func containsWord(s, phrase string) bool {
	for from := 0; ; {
		i := strings.Index(s[from:], phrase)
		if i < 0 {
			return false
		}
		idx := from + i
		if idx == 0 || !isAlnum(s[idx-1]) {
			return true
		}
		from = idx + 1
	}
}

func isAlnum(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9'
}

// All returns every registered ecosystem.
func All() []Ecosystem {
	out := make([]Ecosystem, len(registry))
	copy(out, registry)
	return out
}
