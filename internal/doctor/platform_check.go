package doctor

import (
	"fmt"

	"github.com/thoreinstein/aix/internal/platform"
)

// PlatformCheck verifies installed AI coding platforms.
type PlatformCheck struct{}

// Ensure PlatformCheck implements Check interface.
var _ Check = (*PlatformCheck)(nil)

// NewPlatformCheck creates a new platform detection check.
func NewPlatformCheck() *PlatformCheck {
	return &PlatformCheck{}
}

// Name returns the unique identifier for this check.
func (c *PlatformCheck) Name() string {
	return "platform-detection"
}

// Category returns the grouping for this check.
func (c *PlatformCheck) Category() string {
	return "platform"
}

// Run executes the platform detection check and returns its result.
func (c *PlatformCheck) Run() *CheckResult {
	results := platform.DetectAll()

	// Build details map with platform status information
	platforms := make(map[string]any)
	var installed, notInstalled, partial int

	for _, r := range results {
		info := map[string]any{
			"status":        string(r.Status),
			"global_config": r.GlobalConfig,
			"mcp_config":    r.MCPConfig,
		}
		platforms[r.Name] = info

		switch r.Status {
		case platform.StatusInstalled:
			installed++
		case platform.StatusNotInstalled:
			notInstalled++
		case platform.StatusPartial:
			partial++
		}
	}

	details := map[string]any{
		"platforms":     platforms,
		"installed":     installed,
		"not_installed": notInstalled,
		"partial":       partial,
		"total":         len(results),
	}

	// Determine severity and message based on results
	switch {
	case installed == 0 && partial == 0:
		// No platforms detected at all - warning (aix has nothing to manage)
		return &CheckResult{
			Name:     c.Name(),
			Category: c.Category(),
			Status:   SeverityWarning,
			Message:  "no AI coding platforms detected; aix has nothing to manage",
			Details:  details,
			FixHint:  "install Claude Code, OpenCode, Codex, or Gemini CLI to use aix",
		}
	case partial > 0:
		// Partial installations need attention
		return &CheckResult{
			Name:     c.Name(),
			Category: c.Category(),
			Status:   SeverityWarning,
			Message:  fmt.Sprintf("%d platform(s) have incomplete setup", partial),
			Details:  details,
			FixHint:  "check platform documentation to complete installation",
		}
	default:
		// At least one platform fully installed
		msg := fmt.Sprintf("%d platform(s) installed", installed)
		if notInstalled > 0 {
			msg += fmt.Sprintf(", %d not configured", notInstalled)
		}
		return &CheckResult{
			Name:     c.Name(),
			Category: c.Category(),
			Status:   SeverityPass,
			Message:  msg,
			Details:  details,
		}
	}
}
