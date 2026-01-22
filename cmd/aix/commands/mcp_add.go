package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/platform/claude"
	"github.com/thoreinstein/aix/internal/platform/opencode"
)

// Sentinel errors for MCP add operations.
var (
	errMCPAddMissingCommandOrURL = errors.New("either command or --url is required")
	errMCPAddBothCommandAndURL   = errors.New("cannot specify both command and --url")
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
	mcpAddCmd.Flags().StringVar(&mcpAddURL, "url", "",
		"remote server endpoint for SSE transport")
	mcpAddCmd.Flags().StringSliceVar(&mcpAddEnv, "env", nil,
		"environment variables in KEY=VALUE format (repeatable)")
	mcpAddCmd.Flags().StringVar(&mcpAddTransport, "transport", "",
		"explicit transport type: stdio, sse")
	mcpAddCmd.Flags().StringSliceVar(&mcpAddHeaders, "headers", nil,
		"HTTP headers for SSE auth in KEY=VALUE format (repeatable)")
	mcpAddCmd.Flags().StringSliceVar(&mcpAddPlatforms, "platform", nil,
		"restrict server to specific platform(s): darwin, linux, windows (repeatable)")
	mcpAddCmd.Flags().BoolVarP(&mcpAddForce, "force", "f", false,
		"overwrite if server already exists")
	mcpCmd.AddCommand(mcpAddCmd)
}

var mcpAddCmd = &cobra.Command{
	Use:   "add <name> [command] [args...]",
	Short: "Add an MCP server configuration",
	Long: `Add an MCP server configuration to the targeted platform(s).

For local stdio servers, provide a command and optional arguments:
  aix mcp add github npx -y @modelcontextprotocol/server-github

For remote SSE servers, use the --url flag:
  aix mcp add api-gateway --url=https://api.example.com/mcp

Environment variables can be set with --env (repeatable):
  aix mcp add github npx -y @modelcontextprotocol/server-github \
    --env GITHUB_TOKEN=ghp_xxx

HTTP headers for SSE authentication can be set with --headers (repeatable):
  aix mcp add api-gateway --url=https://api.example.com/mcp \
    --headers "Authorization=Bearer token123"

Platform restrictions (for Claude Code only, lossy for OpenCode):
  aix mcp add macos-tools /usr/local/bin/macos-mcp --platform darwin

Examples:
  aix mcp add github npx -y @modelcontextprotocol/server-github
  aix mcp add api --url=https://api.example.com/mcp --headers "Auth=Bearer token"
  aix mcp add db-tools ./db-mcp --env DB_HOST=localhost --env DB_PORT=5432
  aix mcp add github npx @modelcontextprotocol/server-github --force`,
	Args: cobra.MinimumNArgs(1),
	RunE: runMCPAdd,
}

// runMCPAdd implements the mcp add command logic.
func runMCPAdd(_ *cobra.Command, args []string) error {
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

	// Validate transport value
	switch transport {
	case "stdio", "sse":
		// Valid
	default:
		return fmt.Errorf("invalid --transport %q: must be 'stdio' or 'sse'", transport)
	}

	// Get target platforms
	platforms, err := cli.ResolvePlatforms(GetPlatformFlag())
	if err != nil {
		return err
	}

	// Check for existing servers (unless --force)
	if !mcpAddForce {
		for _, plat := range platforms {
			if _, err := plat.GetMCP(name); err == nil {
				return fmt.Errorf("server %q already exists on %s (use --force to overwrite)",
					name, plat.DisplayName())
			}
		}
	}

	// Add to each platform
	var addedCount int
	for _, plat := range platforms {
		fmt.Printf("Adding '%s' to %s... ", name, plat.DisplayName())

		// Create platform-specific server and add it
		if err := addMCPToPlatform(plat, name, command, cmdArgs, transport, envMap, headersMap); err != nil {
			fmt.Println("failed")
			return fmt.Errorf("failed to add to %s: %w", plat.DisplayName(), err)
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

// addMCPToPlatform adds an MCP server to the specified platform.
func addMCPToPlatform(
	plat cli.Platform,
	name, command string,
	args []string,
	transport string,
	env, headers map[string]string,
) error {
	switch plat.Name() {
	case "claude":
		server := &claude.MCPServer{
			Name:      name,
			Command:   command,
			Args:      args,
			Transport: transport,
			URL:       mcpAddURL,
			Env:       env,
			Headers:   headers,
			Platforms: mcpAddPlatforms,
		}
		return plat.AddMCP(server)

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
		return plat.AddMCP(server)

	default:
		return fmt.Errorf("unsupported platform: %s", plat.Name())
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
			return nil, fmt.Errorf("invalid %s format %q: expected KEY=VALUE", flagName, entry)
		}
		result[key] = value
	}
	return result, nil
}
