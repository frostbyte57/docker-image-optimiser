package rewrite

import "strings"

// alreadyAnnotated reports whether note already appears in the contiguous
// block of dio annotations directly above startLine.
func alreadyAnnotated(lines []string, startLine int, note string) bool {
	for i := startLine - 2; i >= 0 && i < len(lines); i-- {
		t := strings.TrimSpace(lines[i])
		if t == note {
			return true
		}
		if !strings.HasPrefix(t, "# dio[") {
			break
		}
	}
	return false
}
