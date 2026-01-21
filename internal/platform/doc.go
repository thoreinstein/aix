// Package platform provides a platform adapter framework for AI coding assistant
// configuration management.
//
// This package detects and manages configurations for supported AI coding
// assistants: Claude Code, OpenCode, Codex, and Gemini CLI. It provides
// detection capabilities to determine which platforms are installed on
// the current system, and a registry for managing platform adapters.
//
// # Platform Interface
//
// The [Platform] interface defines the contract that all platform adapters must
// implement. Each supported AI coding assistant (Claude, OpenCode, Codex, Gemini)
// provides an adapter that implements this interface:
//
//	type Platform interface {
//	    Name() string              // Platform identifier
//	    GlobalConfigDir() string   // Path to global config directory
//	    MCPConfigPath() string     // Path to MCP config file
//	    InstructionFilename() string // Instruction file name (e.g., CLAUDE.md)
//	}
//
// # Platform Registry
//
// The [Registry] manages platform adapter registration and lookup. Create a new
// registry and register adapters:
//
//	registry := platform.NewRegistry()
//
//	// Register platform adapters
//	if err := registry.Register(claude.NewPlatform()); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get a specific platform
//	if p := registry.Get("claude"); p != nil {
//	    fmt.Printf("Claude config: %s\n", p.GlobalConfigDir())
//	}
//
//	// List all registered platforms
//	for _, p := range registry.All() {
//	    fmt.Printf("%s: %s\n", p.Name(), p.GlobalConfigDir())
//	}
//
//	// List only installed platforms
//	for _, p := range registry.Available() {
//	    fmt.Printf("Installed: %s\n", p.Name())
//	}
//
// The registry is safe for concurrent use and returns platforms in
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
//   - [ErrPlatformNotRegistered]: Platform not found in registry
//   - [ErrPlatformAlreadyRegistered]: Platform name already in use
//   - [ErrInvalidPlatformName]: Platform name not recognized
//   - [ErrNilPlatform]: Attempted to register nil platform
//
// # Thread Safety
//
// All functions and types in this package are safe for concurrent use.
// The [Registry] uses sync.RWMutex for thread-safe operations.
package platform
