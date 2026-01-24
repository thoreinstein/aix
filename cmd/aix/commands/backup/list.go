package backup

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/cmd/aix/commands/flags"
	"github.com/thoreinstein/aix/internal/backup"
	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/errors"
)

var listJSON bool

func init() {
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output in JSON format")
	Cmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available backups",
	Long: `List all available configuration backups grouped by platform.

By default, lists backups for all detected platforms. Use the --platform flag
to limit to a specific platform. Backups are shown in chronological order
with the most recent first.`,
	Example: `  # List all backups
  aix backup list

  # List backups for a specific platform
  aix backup list --platform claude

  # Output as JSON
  aix backup list --json

  See Also:
    aix backup restore - Restore from a backup
    aix backup create  - Create a new backup`,
	RunE: runList,
}

// listOutput represents the JSON output for backup list.
type listOutput struct {
	Platform string       `json:"platform"`
	Backups  []infoOutput `json:"backups"`
}

// infoOutput represents a single backup in JSON output.
type infoOutput struct {
	ID         string    `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	FileCount  int       `json:"file_count"`
	AIXVersion string    `json:"aix_version"`
}

func runList(_ *cobra.Command, _ []string) error {
	return runListWithWriter(os.Stdout)
}

func runListWithWriter(w io.Writer) error {
	platforms, err := cli.ResolvePlatforms(flags.GetPlatformFlag())
	if err != nil {
		return errors.Wrap(err, "resolving platforms")
	}

	mgr := backup.NewManager()

	if listJSON {
		return outputListJSON(w, platforms, mgr)
	}
	return outputListTabular(w, platforms, mgr)
}

func outputListJSON(w io.Writer, platforms []cli.Platform, mgr *backup.Manager) error {
	output := make([]listOutput, 0, len(platforms))

	for _, p := range platforms {
		manifests, err := mgr.List(p.Name())
		if err != nil && !errors.Is(err, backup.ErrNoBackupsFound) {
			return errors.Wrapf(err, "listing backups for %s", p.Name())
		}

		backups := make([]infoOutput, len(manifests))
		for i, m := range manifests {
			backups[i] = infoOutput{
				ID:         m.ID,
				CreatedAt:  m.CreatedAt,
				FileCount:  len(m.Files),
				AIXVersion: m.AIXVersion,
			}
		}

		output = append(output, listOutput{
			Platform: p.Name(),
			Backups:  backups,
		})
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return errors.Wrap(enc.Encode(output), "encoding output")
}

func outputListTabular(w io.Writer, platforms []cli.Platform, mgr *backup.Manager) error {
	hasBackups := false

	for i, p := range platforms {
		manifests, err := mgr.List(p.Name())
		if err != nil && !errors.Is(err, backup.ErrNoBackupsFound) {
			return errors.Wrapf(err, "listing backups for %s", p.Name())
		}

		if len(manifests) > 0 {
			hasBackups = true
		}

		// Add blank line between platforms (but not before first)
		if i > 0 {
			fmt.Fprintln(w)
		}

		// Platform header
		fmt.Fprintf(w, "%sPlatform: %s%s\n", colorCyan+colorBold, p.DisplayName(), colorReset)

		if len(manifests) == 0 {
			fmt.Fprintf(w, "  %s(no backups available)%s\n", colorGray, colorReset)
			continue
		}

		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintf(tw, "  %sID%s\t%sCREATED%s\t%sFILES%s\t%sVERSION%s\n",
			colorBold, colorReset,
			colorBold, colorReset,
			colorBold, colorReset,
			colorBold, colorReset)

		for _, m := range manifests {
			fmt.Fprintf(tw, "  %s%s%s\t%s\t%d\t%s\n",
				colorGreen, m.ID, colorReset,
				m.CreatedAt.Local().Format("2006-01-02 15:04:05"),
				len(m.Files),
				m.AIXVersion)
		}
		tw.Flush()
	}

	if !hasBackups {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "No backups available")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Backups are created automatically before aix modifies configurations.")
		fmt.Fprintln(w, "You can also create a backup manually with: aix backup create")
	}

	return nil
}
