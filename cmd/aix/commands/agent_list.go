package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/cli"
)

var agentListJSON bool

func init() {
	agentListCmd.Flags().BoolVar(&agentListJSON, "json", false, "Output in JSON format")
	agentCmd.AddCommand(agentListCmd)
}

var agentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed agents",
	Long: `List all installed agents grouped by platform.

By default, lists agents for all detected platforms. Use the --platform flag
to limit to specific platforms.

Examples:
  # List all agents
  aix agent list

  # List agents for a specific platform
  aix agent list --platform claude

  # Output as JSON
  aix agent list --json`,
	RunE: runAgentList,
}

// agentListOutput represents the JSON output format for agent list.
type agentListOutput map[string][]agentInfoJSON

// agentInfoJSON represents an agent in JSON output format.
type agentInfoJSON struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func runAgentList(_ *cobra.Command, _ []string) error {
	return runAgentListWithWriter(os.Stdout)
}

// runAgentListWithWriter allows injecting a writer for testing.
func runAgentListWithWriter(w io.Writer) error {
	platforms, err := cli.ResolvePlatforms(GetPlatformFlag())
	if err != nil {
		return err
	}

	if agentListJSON {
		return outputAgentsJSON(w, platforms)
	}
	return outputAgentsTabular(w, platforms)
}

// outputAgentsJSON outputs agents in JSON format.
func outputAgentsJSON(w io.Writer, platforms []cli.Platform) error {
	output := make(agentListOutput)

	for _, p := range platforms {
		agents, err := p.ListAgents()
		if err != nil {
			return fmt.Errorf("listing agents for %s: %w", p.Name(), err)
		}

		infos := make([]agentInfoJSON, len(agents))
		for i, a := range agents {
			infos[i] = agentInfoJSON{
				Name:        a.Name,
				Description: a.Description,
			}
		}
		output[p.Name()] = infos
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

// outputAgentsTabular outputs agents in tabular format grouped by platform.
func outputAgentsTabular(w io.Writer, platforms []cli.Platform) error {
	hasAgents := false

	for i, p := range platforms {
		agents, err := p.ListAgents()
		if err != nil {
			return fmt.Errorf("listing agents for %s: %w", p.Name(), err)
		}

		if len(agents) > 0 {
			hasAgents = true
		}

		// Add blank line between platforms (but not before first)
		if i > 0 {
			fmt.Fprintln(w)
		}

		// Platform header
		fmt.Fprintf(w, "%sPlatform: %s%s\n", colorCyan+colorBold, p.DisplayName(), colorReset)

		if len(agents) == 0 {
			fmt.Fprintf(w, "  %s(no agents installed)%s\n", colorGray, colorReset)
			continue
		}

		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		// Table headers
		fmt.Fprintf(tw, "  %sNAME%s\t%sDESCRIPTION%s\n", colorBold, colorReset, colorBold, colorReset)

		for _, a := range agents {
			desc := truncate(a.Description, 80)
			fmt.Fprintf(tw, "  %s%s%s\t%s\n", colorGreen, a.Name, colorReset, desc)
		}
		tw.Flush()
	}

	if !hasAgents {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "No agents installed")
	}

	return nil
}
