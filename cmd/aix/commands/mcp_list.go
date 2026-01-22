package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/cli"
)

var (
	mcpListJSON        bool
	mcpListShowSecrets bool
)

func init() {
	mcpListCmd.Flags().BoolVar(&mcpListJSON, "json", false, "Output in JSON format")
	mcpListCmd.Flags().BoolVar(&mcpListShowSecrets, "show-secrets", false, "Reveal masked secrets in env values")
	mcpCmd.AddCommand(mcpListCmd)
}

var mcpListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured MCP servers",
	Long: `List all configured MCP servers grouped by platform.

By default, lists MCP servers for all detected platforms. Use the --platform flag
to limit to specific platforms.

Environment variables containing secrets (TOKEN, KEY, SECRET, PASSWORD, AUTH,
CREDENTIAL, API_KEY) are masked by default. Use --show-secrets to reveal them.

Examples:
  # List all MCP servers
  aix mcp list

  # List MCP servers for a specific platform
  aix mcp list --platform claude

  # Output as JSON
  aix mcp list --json

  # Show secret values in environment variables
  aix mcp list --show-secrets`,
	RunE: runMCPList,
}

// mcpListPlatformOutput represents a single platform's MCP servers in JSON output.
type mcpListPlatformOutput struct {
	Platform string              `json:"platform"`
	Servers  []mcpServerInfoJSON `json:"servers"`
}

// mcpServerInfoJSON represents an MCP server in JSON output format.
type mcpServerInfoJSON struct {
	Name      string            `json:"name"`
	Transport string            `json:"transport"`
	Command   string            `json:"command,omitempty"`
	URL       string            `json:"url,omitempty"`
	Disabled  bool              `json:"disabled"`
	Env       map[string]string `json:"env,omitempty"`
}

func runMCPList(_ *cobra.Command, _ []string) error {
	return runMCPListWithWriter(os.Stdout)
}

// runMCPListWithWriter allows injecting a writer for testing.
func runMCPListWithWriter(w io.Writer) error {
	platforms, err := cli.ResolvePlatforms(GetPlatformFlag())
	if err != nil {
		return err
	}

	if mcpListJSON {
		return outputMCPJSON(w, platforms)
	}
	return outputMCPTabular(w, platforms)
}

// outputMCPJSON outputs MCP servers in JSON format.
func outputMCPJSON(w io.Writer, platforms []cli.Platform) error {
	output := make([]mcpListPlatformOutput, 0, len(platforms))

	for _, p := range platforms {
		servers, err := p.ListMCP()
		if err != nil {
			return fmt.Errorf("listing MCP servers for %s: %w", p.Name(), err)
		}

		infos := make([]mcpServerInfoJSON, len(servers))
		for i, s := range servers {
			infos[i] = mcpServerInfoJSON{
				Name:      s.Name,
				Transport: s.Transport,
				Command:   s.Command,
				URL:       s.URL,
				Disabled:  s.Disabled,
				Env:       maskSecrets(s.Env, mcpListShowSecrets),
			}
		}
		output = append(output, mcpListPlatformOutput{
			Platform: p.Name(),
			Servers:  infos,
		})
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

// outputMCPTabular outputs MCP servers in tabular format grouped by platform.
func outputMCPTabular(w io.Writer, platforms []cli.Platform) error {
	hasServers := false

	for i, p := range platforms {
		servers, err := p.ListMCP()
		if err != nil {
			return fmt.Errorf("listing MCP servers for %s: %w", p.Name(), err)
		}

		if len(servers) > 0 {
			hasServers = true
		}

		// Add blank line between platforms (but not before first)
		if i > 0 {
			fmt.Fprintln(w)
		}

		// Platform header
		fmt.Fprintf(w, "%sPlatform: %s%s\n", colorCyan+colorBold, p.DisplayName(), colorReset)

		if len(servers) == 0 {
			fmt.Fprintf(w, "  %s(no MCP servers configured)%s\n", colorGray, colorReset)
			continue
		}

		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		// Table headers
		fmt.Fprintf(tw, "  %sNAME%s\t%sTRANSPORT%s\t%sCOMMAND/URL%s\t%sSTATUS%s\n",
			colorBold, colorReset,
			colorBold, colorReset,
			colorBold, colorReset,
			colorBold, colorReset)

		for _, s := range servers {
			// Determine command/URL to display
			endpoint := s.Command
			if s.URL != "" {
				endpoint = s.URL
			}
			endpoint = truncate(endpoint, 50)

			// Determine status
			status := "enabled"
			statusColor := colorGreen
			if s.Disabled {
				status = "disabled"
				statusColor = colorGray
			}

			fmt.Fprintf(tw, "  %s%s%s\t%s\t%s\t%s%s%s\n",
				colorGreen, s.Name, colorReset,
				s.Transport,
				endpoint,
				statusColor, status, colorReset)
		}
		tw.Flush()
	}

	if !hasServers {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "No MCP servers configured")
	}

	return nil
}

// secretKeyPatterns contains substrings that indicate a key likely contains a secret.
var secretKeyPatterns = []string{
	"TOKEN",
	"KEY",
	"SECRET",
	"PASSWORD",
	"AUTH",
	"CREDENTIAL",
	"API_KEY",
}

// maskSecrets masks secret values in environment variables.
// If showSecrets is true, returns the original map unchanged.
// Secret detection is based on key names containing common secret indicators.
func maskSecrets(env map[string]string, showSecrets bool) map[string]string {
	if env == nil {
		return nil
	}
	if showSecrets {
		return env
	}

	masked := make(map[string]string, len(env))
	for k, v := range env {
		upper := strings.ToUpper(k)
		isSecret := false
		for _, pattern := range secretKeyPatterns {
			if strings.Contains(upper, pattern) {
				isSecret = true
				break
			}
		}
		if isSecret && len(v) > 4 {
			masked[k] = "****" + v[len(v)-4:]
		} else if isSecret {
			masked[k] = "********"
		} else {
			masked[k] = v
		}
	}
	return masked
}
