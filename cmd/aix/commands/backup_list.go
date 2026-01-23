package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/backup"
	"github.com/thoreinstein/aix/internal/cli"
)

var backupListJSON bool

func init() {
	backupListCmd.Flags().BoolVar(&backupListJSON, "json", false, "Output in JSON format")
	backupCmd.AddCommand(backupListCmd)
}

var backupListCmd = &cobra.Command{
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
	RunE: runBackupList,
}

// backupListOutput represents the JSON output for backup list.
type backupListOutput struct {
	Platform string             `json:"platform"`
	Backups  []backupInfoOutput `json:"backups"`
}

// backupInfoOutput represents a single backup in JSON output.
type backupInfoOutput struct {
	ID         string    `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	FileCount  int       `json:"file_count"`
	AIXVersion string    `json:"aix_version"`
}

func runBackupList(_ *cobra.Command, _ []string) error {
	return runBackupListWithWriter(os.Stdout)
}

func runBackupListWithWriter(w io.Writer) error {
	platforms, err := cli.ResolvePlatforms(GetPlatformFlag())
	if err != nil {
		return err
	}

	mgr := backup.NewManager()

	if backupListJSON {
		return outputBackupListJSON(w, platforms, mgr)
	}
	return outputBackupListTabular(w, platforms, mgr)
}

func outputBackupListJSON(w io.Writer, platforms []cli.Platform, mgr *backup.Manager) error {
	output := make([]backupListOutput, 0, len(platforms))

	for _, p := range platforms {
		manifests, err := mgr.List(p.Name())
		if err != nil && !errors.Is(err, backup.ErrNoBackupsFound) {
			return errors.Wrapf(err, "listing backups for %s", p.Name())
		}

		backups := make([]backupInfoOutput, len(manifests))
		for i, m := range manifests {
			backups[i] = backupInfoOutput{
				ID:         m.ID,
				CreatedAt:  m.CreatedAt,
				FileCount:  len(m.Files),
				AIXVersion: m.AIXVersion,
			}
		}

		output = append(output, backupListOutput{
			Platform: p.Name(),
			Backups:  backups,
		})
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

func outputBackupListTabular(w io.Writer, platforms []cli.Platform, mgr *backup.Manager) error {
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
