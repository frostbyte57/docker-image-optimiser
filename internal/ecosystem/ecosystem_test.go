package ecosystem

import "testing"

func TestForCommand(t *testing.T) {
	cases := map[string]string{
		"npm ci":                             "npm",
		"pnpm install --frozen-lockfile":     "pnpm",
		"yarn install":                       "yarn",
		"pip install -r requirements.txt":    "pip",
		"pip3 install flask":                 "pip",
		"poetry install --no-root":           "poetry",
		"uv pip install -r requirements.txt": "uv",
		"go build -o /app ./...":             "go",
		"cargo build --release":              "cargo",
		"mvn -B package":                     "maven",
		"./gradlew build":                    "gradle",
		"composer install --no-dev":          "composer",
		"bundle install":                     "bundler",
		"dotnet restore":                     "dotnet",
		"apt-get install -y curl":            "apt",
		"apk add --no-cache curl":            "apk",
		"dnf install -y gcc":                 "dnf",
		"yum install -y gcc":                 "dnf",
	}
	for cmd, want := range cases {
		got, ok := ForCommand(cmd)
		if !ok {
			t.Errorf("ForCommand(%q): no match, want %s", cmd, want)
			continue
		}
		if got.Name != want {
			t.Errorf("ForCommand(%q) = %s, want %s", cmd, got.Name, want)
		}
	}
}

// Matched must return the exact verb the user wrote, not the registry's first
// alias — otherwise verb-targeted rewrites (e.g. adding --no-cache-dir) silently
// no-op on aliases like "pip3 install".
func TestMatched(t *testing.T) {
	pip, ok := ForCommand("pip3 install flask")
	if !ok {
		t.Fatal("expected pip3 to match the pip ecosystem")
	}
	cases := map[string]string{
		"pip3 install flask":              "pip3 install",
		"pip install -r requirements.txt": "pip install",
		"npm ci":                          "", // different ecosystem, no match
	}
	for args, want := range cases {
		if got := pip.Matched(args); got != want {
			t.Errorf("pip.Matched(%q) = %q, want %q", args, got, want)
		}
	}
}

func TestForCommandNoMatch(t *testing.T) {
	if _, ok := ForCommand("echo hello"); ok {
		t.Error("expected no ecosystem match for a plain echo")
	}
}

func TestKindSplit(t *testing.T) {
	for _, e := range All() {
		switch e.Name {
		case "apt", "apk", "dnf":
			if e.Kind != System {
				t.Errorf("%s should be System", e.Name)
			}
		default:
			if e.Kind != Language {
				t.Errorf("%s should be Language", e.Name)
			}
		}
	}
}
