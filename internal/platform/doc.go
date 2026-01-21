// Package platform provides a platform adapter framework for AI coding assistant
// configuration management.
//
// This package detects and manages configurations for supported AI coding
// assistants: Claude Code, OpenCode, Codex, and Gemini CLI. It provides
// detection capabilities to determine which platforms are installed on
// the current system.
//
// # Platform Detection
//
// Use [DetectPlatform] to check if a specific platform is installed:
//
//	result := platform.DetectPlatform(paths.PlatformClaude)
//	if result != nil && result.Status == platform.StatusInstalled {
//	    fmt.Printf("Claude is installed at %s\n", result.GlobalConfig)
//	}
//
// Use [DetectAll] to discover all platforms regardless of installation status:
//
//	for _, result := range platform.DetectAll() {
//	    fmt.Printf("%s: %s\n", result.Name, result.Status)
//	}
//
// Use [DetectInstalled] to get only platforms that are currently installed:
//
//	installed := platform.DetectInstalled()
//	if len(installed) == 0 {
//	    fmt.Println("No AI coding assistants found")
//	}
//
// # Installation Status
//
// The [InstallStatus] type indicates the installation state of a platform:
//
//   - [StatusInstalled]: Platform config directory exists
//   - [StatusNotInstalled]: Platform config directory does not exist
//   - [StatusPartial]: Reserved for future use (e.g., config exists but binary missing)
//
// # Detection Results
//
// [DetectionResult] contains information about a detected platform:
//
//   - Name: Platform identifier (claude, opencode, codex, gemini)
//   - GlobalConfig: Path to global configuration directory
//   - MCPConfig: Path to MCP configuration file
//   - Status: Current installation status
//
// # Thread Safety
//
// All functions in this package are safe for concurrent use.
package platform
