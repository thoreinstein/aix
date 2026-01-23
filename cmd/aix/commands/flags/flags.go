// Package flags provides shared flag accessors for CLI commands.
// This package exists to avoid import cycles between the root command
// and noun subpackages (skill, mcp, agent, etc.).
package flags

// platformFlag holds the value of the --platform flag.
var platformFlag []string

// GetPlatformFlag returns the current value of the --platform flag.
// This is used by subcommands to access the flag value.
func GetPlatformFlag() []string {
	return platformFlag
}

// SetPlatformFlag sets the platform flag value.
// This is used by the root command to set the flag value after parsing,
// and for programmatic override (e.g., interactive mode).
func SetPlatformFlag(platforms []string) {
	platformFlag = platforms
}
