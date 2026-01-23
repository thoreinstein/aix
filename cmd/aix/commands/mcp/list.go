package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/cmd/aix/commands/flags"
	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/doctor"
)

// ANSI color codes for terminal output.
const (
	colorReset = "\033[0m"
	colorBold  = "\033[1m"
	colorCyan  = "\033[36m"
	colorGreen = "\033[32m"
	colorGray  = "\033[90m"
)

var (
	listJSON        bool
	listShowSecrets bool
)

func init() {
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output in JSON format")
	listCmd.Flags().BoolVar(&listShowSecrets, "show-secrets", false, "Reveal masked secrets in env values")
	Cmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured MCP servers",
	Long: `List all configured MCP servers grouped by platform.

By default, lists MCP servers for all detected platforms. Use the --platform flag
to limit to specific platforms.

Environment variables containing secrets (TOKEN, KEY, SECRET, PASSWORD, AUTH,
CREDENTIAL, API_KEY) are masked by default. Use --show-secrets to reveal them.`,
	Example: `  # List all MCP servers
  aix mcp list

  # List MCP servers for a specific platform
  aix mcp list --platform claude

  # Output as JSON
  aix mcp list --json

  # Show secret values in environment variables
  aix mcp list --show-secrets

  See Also:
    aix mcp show     - Show details of a specific server
    aix mcp add      - Add a new server`,
	RunE: runList,
}

// listPlatformOutput represents a single platform's MCP servers in JSON output.
type listPlatformOutput struct {
	Platform string           `json:"platform"`
	Servers  []serverInfoJSON `json:"servers"`
}

// serverInfoJSON represents an MCP server in JSON output format.
type serverInfoJSON struct {
	Name      string            `json:"name"`
	Transport string            `json:"transport"`
	Command   string            `json:"command,omitempty"`
	URL       string            `json:"url,omitempty"`
	Disabled  bool              `json:"disabled"`
	Env       map[string]string `json:"env,omitempty"`
}

func runList(_ *cobra.Command, _ []string) error {
	return runListWithWriter(os.Stdout)
}

// runListWithWriter allows injecting a writer for testing.
func runListWithWriter(w io.Writer) error {
	platforms, err := cli.ResolvePlatforms(flags.GetPlatformFlag())
	if err != nil {
		return err
	}

	if listJSON {
		return outputJSON(w, platforms)
	}
	return outputTabular(w, platforms)
}

// outputJSON outputs MCP servers in JSON format.
func outputJSON(w io.Writer, platforms []cli.Platform) error {
	output := make([]listPlatformOutput, 0, len(platforms))

	for _, p := range platforms {
		servers, err := p.ListMCP()
		if err != nil {
			return errors.Wrapf(err, "listing MCP servers for %s", p.Name())
		}

		infos := make([]serverInfoJSON, len(servers))
		for i, s := range servers {
			infos[i] = serverInfoJSON{
				Name:      s.Name,
				Transport: s.Transport,
				Command:   s.Command,
				URL:       s.URL,
				Disabled:  s.Disabled,
				Env:       maskSecretsIfNeeded(s.Env),
			}
		}
		output = append(output, listPlatformOutput{
			Platform: p.Name(),
			Servers:  infos,
		})
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

// outputTabular outputs MCP servers in tabular format grouped by platform.
func outputTabular(w io.Writer, platforms []cli.Platform) error {
	hasServers := false

	for i, p := range platforms {
		servers, err := p.ListMCP()
		if err != nil {
			return errors.Wrapf(err, "listing MCP servers for %s", p.Name())
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
		if err := tw.Flush(); err != nil {
			return errors.Wrap(err, "flushing tabwriter")
		}
	}

	if !hasServers {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "No MCP servers configured")
	}

	return nil
}

// maskSecretsIfNeeded conditionally masks secrets based on the --show-secrets flag.
func maskSecretsIfNeeded(env map[string]string) map[string]string {
	if listShowSecrets {
		return env
	}
	return doctor.MaskSecrets(env)
}

// truncate truncates a string to maxLen characters, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
