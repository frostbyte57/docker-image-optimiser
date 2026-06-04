package cli

import "errors"

// ErrFindings signals that lint found issues. main exits non-zero on it but
// prints no extra error text (the findings are already on stdout).
var ErrFindings = errors.New("lint findings")

// errSilent is the internal alias returned by commands.
var errSilent = ErrFindings
