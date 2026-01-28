package mcp

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/cmd/aix/commands/flags"
	"github.com/thoreinstein/aix/internal/backup"
	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/internal/platform/claude"
	"github.com/thoreinstein/aix/internal/platform/opencode"
)

// Sentinel errors for MCP add operations.
var (
	errMCPAddMissingCommandOrURL = errors.New("either command or --url is required")
	errMCPAddBothCommandAndURL   = errors.New("cannot specify both command and --url")
	errMCPAddMissingName         = errors.New("server name is required")
	errMCPAddMissingCommand      = errors.New("command is required for stdio transport")
	errMCPAddMissingURL          = errors.New("URL is required for sse transport")
)

// Package-level flag variables for mcp add command.
var (
	mcpAddURL       string
	mcpAddEnv       []string
	mcpAddTransport string
	mcpAddHeaders   []string
	mcpAddPlatforms []string
	mcpAddForce     bool
)

func init() {
	addCmd.Flags().StringVar(&mcpAddURL, "url", "",
		"remote server endpoint for SSE transport")
	addCmd.Flags().StringSliceVar(&mcpAddEnv, "env", nil,
		"environment variables in KEY=VALUE format (repeatable)")
	addCmd.Flags().StringVar(&mcpAddTransport, "transport", "",
		"explicit transport type: stdio, sse")
	addCmd.Flags().StringSliceVar(&mcpAddHeaders, "headers", nil,
		"HTTP headers for SSE auth in KEY=VALUE format (repeatable)")
	addCmd.Flags().StringSliceVar(&mcpAddPlatforms, "platform", nil,
		"restrict server to specific platform(s): darwin, linux, windows (repeatable)")
	addCmd.Flags().BoolVarP(&mcpAddForce, "force", "f", false,
		"overwrite if server already exists")
	flags.AddScopeFlag(addCmd)
	Cmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:   "add [name] [command] [args...]",
	Short: "Add an MCP server configuration",
	Long: `Add an MCP server configuration to the targeted platform(s).

When called without arguments, runs in interactive mode.

For local stdio servers, provide a command and optional arguments.
For remote SSE servers, use the --url flag.
Environment variables can be set with --env (repeatable).
HTTP headers for SSE authentication can be set with --headers (repeatable).
Platform restrictions (for Claude Code only) can be set with --platform.`,
	Example: `  # Interactive mode
  aix mcp add

  # Add a local stdio server
  aix mcp add github npx -y @modelcontextprotocol/server-github

  # Add a remote SSE server with headers
  aix mcp add api --url=https://api.example.com/mcp --headers "Auth=Bearer token"

  # Add a local server with environment variables
  aix mcp add db-tools ./db-mcp --env DB_HOST=localhost --env DB_PORT=5432

  # Overwrite existing server
  aix mcp add github npx @modelcontextprotocol/server-github --force

  See Also:
    aix mcp list     - List configured servers
    aix mcp remove   - Remove a server`,
	Args: cobra.ArbitraryArgs,
	RunE: runMCPAdd,
}

// runMCPAdd implements the mcp add command logic.
func runMCPAdd(cmd *cobra.Command, args []string) error {
	// Check if interactive mode (no args and no --url flag)
	if len(args) == 0 && mcpAddURL == "" {
		return runMCPAddInteractive(cmd)
	}

	// Non-interactive mode requires at least a name
	if len(args) < 1 {
		return errMCPAddMissingName
	}

	return runMCPAddCore(args)
}

// runMCPAddCore contains the core add logic used by both interactive and non-interactive modes.
func runMCPAddCore(args []string) error {
	name := args[0]
	var command string
	var cmdArgs []string

	if len(args) > 1 {
		command = args[1]
		if len(args) > 2 {
			cmdArgs = args[2:]
		}
	}

	// Validate: either command or --url is required, but not both
	if command == "" && mcpAddURL == "" {
		return errMCPAddMissingCommandOrURL
	}
	if command != "" && mcpAddURL != "" {
		return errMCPAddBothCommandAndURL
	}

	// Parse environment variables
	envMap, err := parseKeyValueSlice(mcpAddEnv, "--env")
	if err != nil {
		return err
	}

	// Parse headers
	headersMap, err := parseKeyValueSlice(mcpAddHeaders, "--headers")
	if err != nil {
		return err
	}

	// Determine transport type
	transport := mcpAddTransport
	if transport == "" {
		if mcpAddURL != "" {
			transport = "sse"
		} else {
			transport = "stdio"
		}
	}

	// Validate transport value and required fields
	switch transport {
	case "stdio":
		if command == "" {
			return errMCPAddMissingCommand
		}
	case "sse":
		if mcpAddURL == "" {
			return errMCPAddMissingURL
		}
	default:
		return errors.Newf("invalid --transport %q: must be 'stdio' or 'sse'", transport)
	}

	// Get target platforms

	platforms, err := cli.ResolvePlatforms(flags.GetPlatformFlag())
	if err != nil {
		return errors.Wrap(err, "resolving platforms")
	}

	// Determine configuration scope
	scope, err := cli.DetermineScope(flags.GetScopeFlag())
	if err != nil {
		return fmt.Errorf("determining configuration scope: %w", err)
	}

	// Check for existing servers (unless --force)
	if !mcpAddForce {
		for _, plat := range platforms {
			if _, err := plat.GetMCP(name, cli.ScopeDefault); err == nil {
				return errors.Newf("server %q already exists on %s (use --force to overwrite)",
					name, plat.DisplayName())
			}
		}
	}

	// Add to each platform
	var addedCount int
	for _, plat := range platforms {
		// Ensure backup exists before modifying
		if err := backup.EnsureBackedUp(plat.Name(), plat.BackupPaths()); err != nil {
			return errors.Wrapf(err, "backing up %s before add", plat.DisplayName())
		}

		fmt.Printf("Adding '%s' to %s... ", name, plat.DisplayName())

		// Create platform-specific server and add it
		if err := addMCPToPlatform(plat, name, command, cmdArgs, transport, envMap, headersMap, scope); err != nil {
			fmt.Println("failed")
			return errors.Wrapf(err, "failed to add to %s", plat.DisplayName())
		}

		fmt.Println("done")
		addedCount++
	}

	// Print summary
	platformWord := "platform"
	if addedCount != 1 {
		platformWord = "platforms"
	}
	fmt.Printf("MCP server '%s' added to %d %s\n", name, addedCount, platformWord)

	return nil
}

// runMCPAddInteractive runs the interactive wizard for adding an MCP server.
func runMCPAddInteractive(_ *cobra.Command) error {
	reader := bufio.NewReader(os.Stdin)

	// 1. Server name
	fmt.Print("Enter server name: ")
	name, err := reader.ReadString('\n')
	if err != nil {
		return errors.Wrap(err, "reading server name")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return errMCPAddMissingName
	}

	// 2. Transport type
	fmt.Println("Select transport type:")
	fmt.Println("  [1] stdio (local command)")
	fmt.Println("  [2] sse (remote URL)")
	fmt.Print("Choice [1]: ")
	choice, err := reader.ReadString('\n')
	if err != nil {
		return errors.Wrap(err, "reading transport choice")
	}
	choice = strings.TrimSpace(choice)
	if choice == "" {
		choice = "1"
	}

	var command string
	var commandArgs []string
	var url string
	var headers map[string]string

	if choice == "1" || choice == "stdio" {
		// 3a. Stdio: command and args
		fmt.Print("Enter command: ")
		command, err = reader.ReadString('\n')
		if err != nil {
			return errors.Wrap(err, "reading command")
		}
		command = strings.TrimSpace(command)
		if command == "" {
			return errMCPAddMissingCommand
		}

		fmt.Print("Enter arguments (space-separated, or empty): ")
		argsLine, err := reader.ReadString('\n')
		if err != nil {
			return errors.Wrap(err, "reading arguments")
		}
		argsLine = strings.TrimSpace(argsLine)
		if argsLine != "" {
			commandArgs = strings.Fields(argsLine)
		}
	} else {
		// 3b. SSE: URL and headers
		fmt.Print("Enter URL: ")
		url, err = reader.ReadString('\n')
		if err != nil {
			return errors.Wrap(err, "reading URL")
		}
		url = strings.TrimSpace(url)
		if url == "" {
			return errMCPAddMissingURL
		}

		fmt.Print("Enter headers (KEY=VALUE, comma-separated, or empty): ")
		headersLine, err := reader.ReadString('\n')
		if err != nil {
			return errors.Wrap(err, "reading headers")
		}
		headersLine = strings.TrimSpace(headersLine)
		if headersLine != "" {
			headers = parseCommaKeyValueList(headersLine)
		}
	}

	// 4. Environment variables
	fmt.Print("Enter environment variables (KEY=VALUE, comma-separated, or empty): ")
	envLine, err := reader.ReadString('\n')
	if err != nil {
		return errors.Wrap(err, "reading environment variables")
	}
	envLine = strings.TrimSpace(envLine)
	var env map[string]string
	if envLine != "" {
		env = parseCommaKeyValueList(envLine)
	}

	// 5. Target platforms
	fmt.Println("Select target platforms:")
	fmt.Println("  [1] All detected platforms")
	fmt.Println("  [2] Claude Code only")
	fmt.Println("  [3] OpenCode only")
	fmt.Print("Choice [1]: ")
	platChoice, err := reader.ReadString('\n')
	if err != nil {
		return errors.Wrap(err, "reading platform choice")
	}
	platChoice = strings.TrimSpace(platChoice)

	// Store original platform flag and restore after
	origPlatform := flags.GetPlatformFlag()
	defer flags.SetPlatformFlag(origPlatform)

	switch platChoice {
	case "2":
		flags.SetPlatformFlag([]string{"claude"})
	case "3":
		flags.SetPlatformFlag([]string{"opencode"})
	default:
		// Keep original (all platforms)
	}

	// Set the package-level flag variables with collected values
	mcpAddURL = url
	mcpAddEnv = formatKeyValueSlice(env)
	mcpAddHeaders = formatKeyValueSlice(headers)

	// Build args slice for the core add function
	finalArgs := []string{name}
	if command != "" {
		finalArgs = append(finalArgs, command)
		finalArgs = append(finalArgs, commandArgs...)
	}

	return runMCPAddCore(finalArgs)
}

// addMCPToPlatform adds an MCP server to the specified platform.
func addMCPToPlatform(
	plat cli.Platform,
	name, command string,
	args []string,
	transport string,
	env, headers map[string]string,
	scope cli.Scope,
) error {
	switch plat.Name() {
	case "claude":
		// Map canonical transport to Claude type: stdio -> stdio, sse -> http
		claudeType := transport
		if transport == "sse" {
			claudeType = "http"
		}

		server := &claude.MCPServer{
			Name:      name,
			Command:   command,
			Args:      args,
			Type:      claudeType,
			URL:       mcpAddURL,
			Env:       env,
			Headers:   headers,
			Platforms: mcpAddPlatforms,
		}
		return errors.Wrap(plat.AddMCP(server, scope), "adding MCP server to Claude")

	case "opencode":
		// Show warning if platforms restriction is specified
		if len(mcpAddPlatforms) > 0 {
			fmt.Printf("\n  Warning: OpenCode does not support platform restrictions; "+
				"--platform %s will be ignored\n", strings.Join(mcpAddPlatforms, ", "))
		}

		// OpenCode combines command and args into a single slice
		var cmdSlice []string
		if command != "" {
			cmdSlice = append([]string{command}, args...)
		}

		// Map transport types: stdio -> local, sse -> remote
		typ := "local"
		if transport == "sse" {
			typ = "remote"
		}

		server := &opencode.MCPServer{
			Name:        name,
			Command:     cmdSlice,
			Type:        typ,
			URL:         mcpAddURL,
			Environment: env,
			Headers:     headers,
		}
		return errors.Wrap(plat.AddMCP(server, scope), "adding MCP server to OpenCode")

	default:
		return errors.Newf("unsupported platform: %s", plat.Name())
	}
}

// parseKeyValueSlice parses a slice of KEY=VALUE strings into a map.
// Returns an error if any entry is malformed.
func parseKeyValueSlice(entries []string, flagName string) (map[string]string, error) {
	if len(entries) == 0 {
		return nil, nil
	}

	result := make(map[string]string, len(entries))
	for _, entry := range entries {
		key, value, found := strings.Cut(entry, "=")
		if !found || key == "" {
			return nil, errors.Newf("invalid %s format %q: expected KEY=VALUE", flagName, entry)
		}
		result[key] = value
	}
	return result, nil
}

// parseCommaKeyValueList parses a comma-separated string of KEY=VALUE pairs into a map.
// Invalid entries are silently skipped.
func parseCommaKeyValueList(s string) map[string]string {
	result := make(map[string]string)
	pairs := strings.Split(s, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		key, value, found := strings.Cut(pair, "=")
		if found && key != "" {
			result[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
	}
	return result
}

// formatKeyValueSlice converts a map to a slice of KEY=VALUE strings.
func formatKeyValueSlice(m map[string]string) []string {
	if len(m) == 0 {
		return nil
	}
	result := make([]string, 0, len(m))
	for k, v := range m {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}
