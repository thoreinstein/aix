// Package cmd contains build-time variables injected via ldflags.
package cmd

// Build-time variables set via ldflags.
var (
	// Version is the semantic version of the build.
	Version = "dev"
	// Commit is the git commit SHA of the build.
	Commit = "none"
	// Date is the build date.
	Date = "unknown"
)
