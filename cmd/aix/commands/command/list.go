package command

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
)

// ANSI color codes for terminal output.
const (
	colorReset = "\033[0m"
	colorBold  = "\033[1m"
	colorCyan  = "\033[36m"
	colorGreen = "\033[32m"
	colorGray  = "\033[90m"
)

var listJSON bool

func init() {
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output in JSON format")
	Cmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed slash commands",
	Long: `List all installed slash commands grouped by platform.

By default, lists commands for all detected platforms. Use the --platform flag
to limit to specific platforms.`,
	Example: `  # List all commands
  aix command list

  # List commands for a specific platform
  aix command list --platform claude

  # Output as JSON
  aix command list --json

  See Also:
    aix command show     - Show command details
    aix command edit     - Edit a command definition
    aix command install  - Install a new command`,
	RunE: runList,
}

// listOutput represents the JSON output format for command list.
type listOutput map[string][]infoJSON

// infoJSON represents a command in JSON output format.
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
		return err
	}

	if listJSON {
		return outputJSON(w, platforms)
	}
	return outputTabular(w, platforms)
}

// outputJSON outputs commands in JSON format.
func outputJSON(w io.Writer, platforms []cli.Platform) error {
	output := make(listOutput)

	for _, p := range platforms {
		commands, err := p.ListCommands()
		if err != nil {
			return errors.Wrapf(err, "listing commands for %s", p.Name())
		}

		infos := make([]infoJSON, len(commands))
		for i, c := range commands {
			infos[i] = infoJSON{
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

// outputTabular outputs commands in tabular format grouped by platform.
func outputTabular(w io.Writer, platforms []cli.Platform) error {
	hasCommands := false

	for i, p := range platforms {
		commands, err := p.ListCommands()
		if err != nil {
			return errors.Wrapf(err, "listing commands for %s", p.Name())
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

// truncate shortens a string to maxLen characters, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
