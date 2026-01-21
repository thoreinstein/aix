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

var commandListJSON bool

func init() {
	commandListCmd.Flags().BoolVar(&commandListJSON, "json", false, "Output in JSON format")
	commandCmd.AddCommand(commandListCmd)
}

var commandListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed slash commands",
	Long: `List all installed slash commands grouped by platform.

By default, lists commands for all detected platforms. Use the --platform flag
to limit to specific platforms.

Examples:
  # List all commands
  aix command list

  # List commands for a specific platform
  aix command list --platform claude

  # Output as JSON
  aix command list --json`,
	RunE: runCommandList,
}

// commandListOutput represents the JSON output format for command list.
type commandListOutput map[string][]commandInfoJSON

// commandInfoJSON represents a command in JSON output format.
type commandInfoJSON struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func runCommandList(_ *cobra.Command, _ []string) error {
	return runCommandListWithWriter(os.Stdout)
}

// runCommandListWithWriter allows injecting a writer for testing.
func runCommandListWithWriter(w io.Writer) error {
	platforms, err := cli.ResolvePlatforms(GetPlatformFlag())
	if err != nil {
		return err
	}

	if commandListJSON {
		return outputCommandsJSON(w, platforms)
	}
	return outputCommandsTabular(w, platforms)
}

// outputCommandsJSON outputs commands in JSON format.
func outputCommandsJSON(w io.Writer, platforms []cli.Platform) error {
	output := make(commandListOutput)

	for _, p := range platforms {
		commands, err := p.ListCommands()
		if err != nil {
			return fmt.Errorf("listing commands for %s: %w", p.Name(), err)
		}

		infos := make([]commandInfoJSON, len(commands))
		for i, c := range commands {
			infos[i] = commandInfoJSON{
				Name:        c.Name,
				Description: c.Description,
			}
		}
		output[p.Name()] = infos
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

// outputCommandsTabular outputs commands in tabular format grouped by platform.
func outputCommandsTabular(w io.Writer, platforms []cli.Platform) error {
	hasCommands := false

	for i, p := range platforms {
		commands, err := p.ListCommands()
		if err != nil {
			return fmt.Errorf("listing commands for %s: %w", p.Name(), err)
		}

		if len(commands) > 0 {
			hasCommands = true
		}

		// Add blank line between platforms (but not before first)
		if i > 0 {
			fmt.Fprintln(w)
		}

		// Platform header
		fmt.Fprintf(w, "%sPlatform: %s%s\n", colorCyan+colorBold, p.DisplayName(), colorReset)

		if len(commands) == 0 {
			fmt.Fprintf(w, "  %s(no commands installed)%s\n", colorGray, colorReset)
			continue
		}

		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		// Table headers
		fmt.Fprintf(tw, "  %sNAME%s\t%sDESCRIPTION%s\n", colorBold, colorReset, colorBold, colorReset)

		for _, c := range commands {
			desc := truncate(c.Description, 80)
			fmt.Fprintf(tw, "  %s/%s%s\t%s\n", colorGreen, c.Name, colorReset, desc)
		}
		tw.Flush()
	}

	if !hasCommands {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "No commands installed")
	}

	return nil
}
