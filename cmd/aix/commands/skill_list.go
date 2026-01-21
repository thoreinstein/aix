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

		fmt.Fprintf(w, "Platform: %s\n", p.DisplayName())

		if len(skills) == 0 {
			fmt.Fprintln(w, "  (no skills installed)")
			continue
		}

		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "  NAME\tDESCRIPTION")
		for _, s := range skills {
			fmt.Fprintf(tw, "  %s\t%s\n", s.Name, s.Description)
		}
		tw.Flush()
	}

	if !hasSkills {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "No skills installed")
	}

	return nil
}
