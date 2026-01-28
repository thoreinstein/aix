package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/cmd/aix/commands/flags"
	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/errors"
)

var listJSON bool

func init() {
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output in JSON format")
	flags.AddScopeFlag(listCmd)
	Cmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
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
	RunE: runList,
}

// listOutput represents the JSON output format for agent list.
type listOutput map[string][]infoJSON

// infoJSON represents an agent in JSON output format.
type infoJSON struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func runList(_ *cobra.Command, _ []string) error {
	return runListWithWriter(os.Stdout)
}

// runListWithWriter allows injecting a writer for testing.
func runListWithWriter(w io.Writer) error {
	platforms, err := cli.ResolvePlatforms(flags.GetPlatformFlag())
	if err != nil {
		return errors.Wrap(err, "resolving platforms")
	}

	scope := cli.ParseScope(flags.GetScopeFlag())

	if listJSON {
		return outputListJSON(w, platforms, scope)
	}
	return outputListTabular(w, platforms, scope)
}

// outputListJSON outputs agents in JSON format.
func outputListJSON(w io.Writer, platforms []cli.Platform, scope cli.Scope) error {
	output := make(listOutput)

	for _, p := range platforms {
		agents, err := p.ListAgents(scope)
		if err != nil {
			return fmt.Errorf("listing agents for %s: %w", p.Name(), err)
		}

		infos := make([]infoJSON, len(agents))
		for i, a := range agents {
			infos[i] = infoJSON{
				Name:        a.Name,
				Description: a.Description,
			}
		}
		output[p.Name()] = infos
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return errors.Wrap(enc.Encode(output), "encoding output")
}

// ANSI color codes.
const (
	colorReset = "\033[0m"
	colorBold  = "\033[1m"
	colorCyan  = "\033[36m"
	colorGreen = "\033[32m"
	colorGray  = "\033[90m"
)

// outputListTabular outputs agents in tabular format grouped by platform.
func outputListTabular(w io.Writer, platforms []cli.Platform, scope cli.Scope) error {
	hasAgents := false

	for i, p := range platforms {
		agents, err := p.ListAgents(scope)
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

// truncate shortens a string to maxLen runes, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen < 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}
