// Package platform provides platform detection and registration for AI coding
// assistant configuration management.
//
// This package detects and manages configurations for supported AI coding
// assistants: Claude Code, OpenCode, Codex, and Gemini CLI. It provides
// detection capabilities to determine which platforms are installed on
// the current system, and a registry for tracking registered platform names.
//
// # Platform Registry
//
// The [Registry] tracks which platform names are registered. Create a new
// registry and register platform names:
//
//	registry := platform.NewRegistry()
//
//	// Register platform names
//	if err := registry.Register("claude"); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Check if a platform is registered
//	if registry.Get("claude") {
//	    fmt.Println("Claude is registered")
//	}
//
//	// List all registered platform names
//	for _, name := range registry.All() {
//	    fmt.Printf("Registered: %s\n", name)
//	}
//
//	// List only installed platforms
//	for _, name := range registry.Available() {
//	    fmt.Printf("Installed: %s\n", name)
//	}
//
// The registry is safe for concurrent use and returns platform names in
// deterministic (alphabetical) order.
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
// # Sentinel Errors
//
// The registry operations return specific sentinel errors:
//
//   - [ErrPlatformAlreadyRegistered]: Platform name already in use
//   - [ErrInvalidPlatformName]: Platform name not recognized
//
// # Thread Safety
//
// All functions and types in this package are safe for concurrent use.
// The [Registry] uses sync.RWMutex for thread-safe operations.
package platform
