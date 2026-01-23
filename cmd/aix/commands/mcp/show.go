package mcp

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/cmd/aix/commands/flags"
	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/doctor"
	"github.com/thoreinstein/aix/internal/platform/claude"
	"github.com/thoreinstein/aix/internal/platform/opencode"
)

var (
	showJSON        bool
	showShowSecrets bool
)

func init() {
	showCmd.Flags().BoolVar(&showJSON, "json", false, "Output as JSON")
	showCmd.Flags().BoolVar(&showShowSecrets, "show-secrets", false, "Reveal masked secrets in environment variables and headers")
	Cmd.AddCommand(showCmd)
}

var showCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Display MCP server details",
	Long: `Display detailed information about an MCP server configuration.

Searches for the server across all detected platforms (or only the specified
--platform). Shows transport, command, args, environment variables, headers,
and status. Highlights any configuration differences between platforms.

Environment variables and headers are masked by default to protect secrets.
Use --show-secrets to reveal the full values.`,
	Example: `  # Show details of a server
  aix mcp show github

  # Show details including secrets
  aix mcp show github --show-secrets

  # Output as JSON
  aix mcp show github --json

  # Show details for a specific platform
  aix mcp show github --platform claude

  See Also:
    aix mcp list     - List all servers
    aix mcp remove   - Remove this server`,
	Args: cobra.ExactArgs(1),
	RunE: runShow,
}

// serverDetail holds unified MCP server information for display.
type serverDetail struct {
	Platform  string            `json:"platform"`
	Transport string            `json:"transport"`
	Command   string            `json:"command"`
	Args      []string          `json:"args"`
	URL       string            `json:"url"`
	Disabled  bool              `json:"disabled"`
	Env       map[string]string `json:"env,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	Platforms []string          `json:"platforms,omitempty"` // OS platform restrictions (Claude only)
}

// showOutput is the JSON output structure.
type showOutput struct {
	Name        string                   `json:"name"`
	Platforms   map[string]*serverDetail `json:"platforms"`
	Differences []string                 `json:"differences"`
}

func runShow(_ *cobra.Command, args []string) error {
	name := args[0]

	platforms, err := cli.ResolvePlatforms(flags.GetPlatformFlag())
	if err != nil {
		return err
	}

	// Collect server info from all platforms where it exists
	details := make(map[string]*serverDetail)

	for _, p := range platforms {
		serverAny, err := p.GetMCP(name)
		if err != nil {
			// Server not found on this platform, continue to next
			continue
		}

		detail := extractServerDetail(serverAny, p.DisplayName())
		if detail != nil {
			// Apply masking if not showing secrets
			if !showShowSecrets {
				detail.Env = doctor.MaskSecrets(detail.Env)
				detail.Headers = doctor.MaskSecrets(detail.Headers)
			}
			details[p.Name()] = detail
		}
	}

	if len(details) == 0 {
		return errors.Newf("MCP server %q not found on any platform", name)
	}

	// Find differences between platform configurations
	differences := findDifferences(details)

	if showJSON {
		return outputShowJSON(name, details, differences)
	}

	return outputShowText(name, details, differences)
}

// extractServerDetail converts a platform-specific MCP server to the unified detail struct.
func extractServerDetail(server any, platformName string) *serverDetail {
	switch s := server.(type) {
	case *claude.MCPServer:
		return extractClaudeMCPServer(s, platformName)
	case *opencode.MCPServer:
		return extractOpenCodeMCPServer(s, platformName)
	default:
		return nil
	}
}

// extractClaudeMCPServer extracts details from a Claude MCP server.
func extractClaudeMCPServer(s *claude.MCPServer, platformName string) *serverDetail {
	// Map Claude type to canonical transport: stdio -> stdio, http -> sse
	transport := s.Type
	switch transport {
	case "http":
		transport = "sse"
	case "":
		// Infer transport from URL presence
		if s.URL != "" {
			transport = "sse"
		} else {
			transport = "stdio"
		}
	}

	return &serverDetail{
		Platform:  platformName,
		Transport: transport,
		Command:   s.Command,
		Args:      s.Args,
		URL:       s.URL,
		Disabled:  s.Disabled,
		Env:       s.Env,
		Headers:   s.Headers,
		Platforms: s.Platforms,
	}
}

// extractOpenCodeMCPServer extracts details from an OpenCode MCP server.
func extractOpenCodeMCPServer(s *opencode.MCPServer, platformName string) *serverDetail {
	transport := "stdio"
	if s.Type == "remote" || s.URL != "" {
		transport = "sse"
	}

	var command string
	var args []string
	if len(s.Command) > 0 {
		command = s.Command[0]
		if len(s.Command) > 1 {
			args = s.Command[1:]
		}
	}

	// Convert OpenCode's Enabled (positive) to Disabled (negative)
	disabled := s.Enabled != nil && !*s.Enabled

	return &serverDetail{
		Platform:  platformName,
		Transport: transport,
		Command:   command,
		Args:      args,
		URL:       s.URL,
		Disabled:  disabled,
		Env:       s.Environment,
		Headers:   s.Headers,
	}
}

// findDifferences compares server configurations across platforms and returns differences.
func findDifferences(details map[string]*serverDetail) []string {
	if len(details) < 2 {
		return nil
	}

	var differences []string

	// Get sorted platform names for deterministic comparison
	platformNames := make([]string, 0, len(details))
	for name := range details {
		platformNames = append(platformNames, name)
	}
	sort.Strings(platformNames)

	// Use first platform as reference
	ref := details[platformNames[0]]
	refName := platformNames[0]

	for _, pName := range platformNames[1:] {
		other := details[pName]

		if ref.Transport != other.Transport {
			differences = append(differences, fmt.Sprintf("Transport differs: %s=%s, %s=%s", refName, ref.Transport, pName, other.Transport))
		}
		if ref.Command != other.Command {
			differences = append(differences, fmt.Sprintf("Command differs: %s=%q, %s=%q", refName, ref.Command, pName, other.Command))
		}
		if !reflect.DeepEqual(ref.Args, other.Args) {
			differences = append(differences, fmt.Sprintf("Args differ: %s=%v, %s=%v", refName, ref.Args, pName, other.Args))
		}
		if ref.URL != other.URL {
			differences = append(differences, fmt.Sprintf("URL differs: %s=%q, %s=%q", refName, ref.URL, pName, other.URL))
		}
		if ref.Disabled != other.Disabled {
			differences = append(differences, fmt.Sprintf("Status differs: %s=%s, %s=%s", refName, statusString(ref.Disabled), pName, statusString(other.Disabled)))
		}

		// Compare environment variables
		envDiffs := compareMapKeys(ref.Env, other.Env, refName, pName, "Env")
		differences = append(differences, envDiffs...)

		// Compare headers
		headerDiffs := compareMapKeys(ref.Headers, other.Headers, refName, pName, "Headers")
		differences = append(differences, headerDiffs...)

		// Compare platforms (Claude-specific)
		if !reflect.DeepEqual(ref.Platforms, other.Platforms) {
			differences = append(differences, fmt.Sprintf("Platforms (OS) differs: %s=%v, %s=%v", refName, ref.Platforms, pName, other.Platforms))
		}
	}

	return differences
}

// compareMapKeys compares two maps and returns differences in their keys/values.
func compareMapKeys(m1, m2 map[string]string, name1, name2, fieldName string) []string {
	var diffs []string

	// Find keys in m1 not in m2
	for k := range m1 {
		if _, ok := m2[k]; !ok {
			diffs = append(diffs, fmt.Sprintf("%s: %s has %s, %s does not", fieldName, name1, k, name2))
		}
	}

	// Find keys in m2 not in m1
	for k := range m2 {
		if _, ok := m1[k]; !ok {
			diffs = append(diffs, fmt.Sprintf("%s: %s has %s, %s does not", fieldName, name2, k, name1))
		}
	}

	// Compare values for shared keys (only if unmasked)
	for k, v1 := range m1 {
		if v2, ok := m2[k]; ok && v1 != v2 {
			diffs = append(diffs, fmt.Sprintf("%s[%s] value differs between platforms", fieldName, k))
		}
	}

	return diffs
}

func statusString(disabled bool) string {
	if disabled {
		return "disabled"
	}
	return "enabled"
}

func outputShowJSON(name string, details map[string]*serverDetail, differences []string) error {
	output := showOutput{
		Name:        name,
		Platforms:   details,
		Differences: differences,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshaling JSON")
	}
	fmt.Println(string(data))
	return nil
}

func outputShowText(name string, details map[string]*serverDetail, differences []string) error {
	fmt.Printf("MCP Server: %s\n", name)

	// Get sorted platform names for deterministic output
	platformNames := make([]string, 0, len(details))
	for pName := range details {
		platformNames = append(platformNames, pName)
	}
	sort.Strings(platformNames)

	for _, pName := range platformNames {
		detail := details[pName]
		fmt.Printf("\nPlatform: %s\n", detail.Platform)
		fmt.Printf("  Transport:  %s\n", detail.Transport)

		if detail.Command != "" {
			fmt.Printf("  Command:    %s\n", detail.Command)
		}
		if len(detail.Args) > 0 {
			fmt.Printf("  Args:       %s\n", strings.Join(detail.Args, " "))
		}
		if detail.URL != "" {
			fmt.Printf("  URL:        %s\n", detail.URL)
		}

		fmt.Printf("  Status:     %s\n", statusString(detail.Disabled))

		if len(detail.Platforms) > 0 {
			fmt.Printf("  OS:         %s\n", strings.Join(detail.Platforms, ", "))
		}

		if len(detail.Env) > 0 {
			fmt.Println("  Environment:")
			printSortedMap(detail.Env, "    ")
		}

		if len(detail.Headers) > 0 {
			fmt.Println("  Headers:")
			printSortedMap(detail.Headers, "    ")
		}
	}

	// Print differences summary
	fmt.Println()
	if len(differences) == 0 && len(details) > 1 {
		fmt.Println("\u26a0\ufe0f  Configuration is identical across platforms")
	} else if len(differences) > 0 {
		fmt.Println("\u26a0\ufe0f  Differences detected:")
		for _, diff := range differences {
			fmt.Printf("  - %s\n", diff)
		}
	}

	return nil
}

// printSortedMap prints a map with sorted keys for deterministic output.
func printSortedMap(m map[string]string, indent string) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Printf("%s%s: %s\n", indent, k, m[k])
	}
}
