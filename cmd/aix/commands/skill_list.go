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

var skillListJSON bool

func init() {
	skillListCmd.Flags().BoolVar(&skillListJSON, "json", false, "Output in JSON format")
	skillCmd.AddCommand(skillListCmd)
}

var skillListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed skills",
	Long: `List all installed skills grouped by platform.

By default, lists skills for all detected platforms. Use the --platform flag
to limit to specific platforms.

Examples:
  # List all skills
  aix skill list

  # List skills for a specific platform
  aix skill list --platform claude

  # Output as JSON
  aix skill list --json`,
	RunE: runSkillList,
}

// skillListOutput represents the JSON output format for skill list.
type skillListOutput map[string][]skillInfoJSON

// skillInfoJSON represents a skill in JSON output format.
type skillInfoJSON struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func runSkillList(_ *cobra.Command, _ []string) error {
	return runSkillListWithWriter(os.Stdout)
}

// runSkillListWithWriter allows injecting a writer for testing.
func runSkillListWithWriter(w io.Writer) error {
	platforms, err := cli.ResolvePlatforms(GetPlatformFlag())
	if err != nil {
		return err
	}

	if skillListJSON {
		return outputSkillsJSON(w, platforms)
	}
	return outputSkillsTabular(w, platforms)
}

// outputSkillsJSON outputs skills in JSON format.
func outputSkillsJSON(w io.Writer, platforms []cli.Platform) error {
	output := make(skillListOutput)

	for _, p := range platforms {
		skills, err := p.ListSkills()
		if err != nil {
			return fmt.Errorf("listing skills for %s: %w", p.Name(), err)
		}

		infos := make([]skillInfoJSON, len(skills))
		for i, s := range skills {
			infos[i] = skillInfoJSON{
				Name:        s.Name,
				Description: s.Description,
			}
		}
		output[p.Name()] = infos
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

// ANSI color codes
const (
	colorReset = "\033[0m"
	colorBold  = "\033[1m"
	colorCyan  = "\033[36m"
	colorGreen = "\033[32m"
	colorGray  = "\033[90m"
)

// outputSkillsTabular outputs skills in tabular format grouped by platform.
func outputSkillsTabular(w io.Writer, platforms []cli.Platform) error {
	hasSkills := false

	for i, p := range platforms {
		skills, err := p.ListSkills()
		if err != nil {
			return fmt.Errorf("listing skills for %s: %w", p.Name(), err)
		}

		if len(skills) > 0 {
			hasSkills = true
		}

		// Add blank line between platforms (but not before first)
		if i > 0 {
			fmt.Fprintln(w)
		}

		// Platform header
		fmt.Fprintf(w, "%sPlatform: %s%s\n", colorCyan+colorBold, p.DisplayName(), colorReset)

		if len(skills) == 0 {
			fmt.Fprintf(w, "  %s(no skills installed)%s\n", colorGray, colorReset)
			continue
		}

		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		// Table headers
		fmt.Fprintf(tw, "  %sNAME%s\t%sDESCRIPTION%s\n", colorBold, colorReset, colorBold, colorReset)

		for _, s := range skills {
			desc := truncate(s.Description, 80)
			fmt.Fprintf(tw, "  %s%s%s\t%s\n", colorGreen, s.Name, colorReset, desc)
		}
		tw.Flush()
	}

	if !hasSkills {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "No skills installed")
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
