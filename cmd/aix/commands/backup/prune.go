package backup

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/cmd/aix/commands/flags"
	"github.com/thoreinstein/aix/internal/backup"
	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/errors"
)

var pruneKeep int

func init() {
	pruneCmd.Flags().IntVar(&pruneKeep, "keep", backup.DefaultRetentionCount,
		"Number of backups to retain per platform")
	Cmd.AddCommand(pruneCmd)
}

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove old backups",
	Long: `Remove old backups beyond the retention count.

By default, keeps the 5 most recent backups per platform and removes older ones.
Use the --keep flag to specify a different retention count.

By default, prunes backups for all detected platforms. Use the --platform
flag to limit to specific platforms.`,
	Example: `  # Prune all platforms, keeping default (5) backups each
  aix backup prune

  # Keep only the 3 most recent backups
  aix backup prune --keep 3

  # Prune only Claude backups
  aix backup prune --platform claude

  # Remove all backups (keep 0)
  aix backup prune --keep 0

  See Also:
    aix backup list   - List available backups
    aix backup create - Create a new backup`,
	RunE: runPrune,
}

func runPrune(_ *cobra.Command, _ []string) error {
	return runPruneWithWriter(os.Stdout)
}

func runPruneWithWriter(w io.Writer) error {
	if pruneKeep < 0 {
		return errors.New("--keep must be non-negative")
	}

	platforms, err := cli.ResolvePlatforms(flags.GetPlatformFlag())
	if err != nil {
		return errors.Wrap(err, "resolving platforms")
	}

	mgr := backup.NewManager()
	pruned := 0

	for _, p := range platforms {
		// Get current backup count
		manifests, err := mgr.List(p.Name())
		if err != nil {
			if errors.Is(err, backup.ErrNoBackupsFound) {
				continue
			}
			return errors.Wrapf(err, "listing backups for %s", p.Name())
		}

		toRemove := len(manifests) - pruneKeep
		if toRemove <= 0 {
			continue
		}

		if err := mgr.Prune(p.Name(), pruneKeep); err != nil {
			return errors.Wrapf(err, "pruning backups for %s", p.Name())
		}

		fmt.Fprintf(w, "%s✓ %s: removed %d old backup(s)%s\n",
			colorGreen, p.DisplayName(), toRemove, colorReset)
		pruned += toRemove
	}

	if pruned == 0 {
		fmt.Fprintln(w, "No backups to prune")
	} else {
		fmt.Fprintf(w, "\nTotal: removed %d backup(s)\n", pruned)
	}

	return nil
}
