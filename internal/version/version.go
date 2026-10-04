// Package version provides build-time version information for procscope.
//
// Values are injected via -ldflags at build time. See the Makefile for details.
package version

import "fmt"

// These variables are set at build time via -ldflags.
// Do NOT use timestamps — use deterministic values from git tags only.
var (
	// Version is the semantic version tag (e.g., "0.1.0", "dev").
	Version = "dev"
	// Commit is the short git commit hash.
	Commit = "unknown"
)

// Full returns a human-readable version string.
func Full() string {
	return fmt.Sprintf("procscope %s (commit %s)", Version, Commit)
}

// Short returns just the version tag.
func Short() string {
	return Version
}
