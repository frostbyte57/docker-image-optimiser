package build

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// WarmRebuild measures how long a *rebuild* takes after a source change — the
// scenario where layer ordering and cache mounts actually pay off. It works in a
// throwaway copy of the context so the user's files are never touched:
//  1. build once with the cache enabled (warms it),
//  2. add a new file to bust the `COPY . .` layer (simulating a code edit),
//  3. rebuild with the cache enabled and time that.
func WarmRebuild(contextDir, dockerfile, tag string) (time.Duration, error) {
	tmp, err := copyDir(contextDir)
	if err != nil {
		return 0, err
	}
	defer os.RemoveAll(tmp)

	if _, err := Build(tmp, dockerfile, tag, false); err != nil { // warm
		return 0, err
	}

	bust := filepath.Join(tmp, ".dio-cachebust")
	if err := os.WriteFile(bust, []byte(fmt.Sprintf("%d", time.Now().UnixNano())), 0o644); err != nil {
		return 0, err
	}

	start := time.Now()
	if _, err := Build(tmp, dockerfile, tag, false); err != nil { // rebuild
		return 0, err
	}
	return time.Since(start), nil
}

// copyDir recursively copies src into a fresh temp directory and returns its
// path. Symlinks are followed as regular files; the .git directory is skipped
// for speed since it never affects a build.
func copyDir(src string) (string, error) {
	dst, err := os.MkdirTemp("", "dio-ctx-*")
	if err != nil {
		return "", err
	}

	err = filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == ".git" && d.IsDir() {
			return filepath.SkipDir
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
	if err != nil {
		os.RemoveAll(dst)
		return "", err
	}
	return dst, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
