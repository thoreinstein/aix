package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/backup"
	"github.com/thoreinstein/aix/internal/cli"
)

func init() {
	backupCmd.AddCommand(backupRestoreCmd)
}

var backupRestoreCmd = &cobra.Command{
	Use:   "restore [backup-id]",
	Short: "Restore from a backup",
	Long: `Restore platform configuration from a backup.

If no backup ID is provided, restores from the most recent backup for the
specified platform. The --platform flag is required to avoid accidental
restoration to the wrong platform.

All files in the backup are restored to their original locations, preserving
permissions. Existing files are overwritten.`,
	Example: `  # Restore from the most recent Claude backup
  aix backup restore --platform claude

  # Restore from a specific backup
  aix backup restore 20260123T100712 --platform claude

  # List available backups first
  aix backup list --platform claude

  See Also:
    aix backup list   - List available backups
    aix backup create - Create a new backup`,
	Args: cobra.MaximumNArgs(1),
	RunE: runBackupRestore,
}

func runBackupRestore(cmd *cobra.Command, args []string) error {
	return runBackupRestoreWithWriter(cmd, args, os.Stdout)
}

func runBackupRestoreWithWriter(_ *cobra.Command, args []string, w io.Writer) error {
	platformFlag := GetPlatformFlag()
	if len(platformFlag) == 0 {
		return errors.New("--platform is required for restore")
	}

	platforms, err := cli.ResolvePlatforms(platformFlag)
	if err != nil {
		return err
	}
	if len(platforms) != 1 {
		return errors.New("restore requires exactly one platform")
	}

	platform := platforms[0]
	mgr := backup.NewManager()

	// Determine backup ID
	var backupID string
	if len(args) > 0 {
		backupID = args[0]
	} else {
		// Get most recent backup
		manifests, err := mgr.List(platform.Name())
		if err != nil {
			if errors.Is(err, backup.ErrNoBackupsFound) {
				return errors.Errorf("no backups found for %s", platform.DisplayName())
			}
			return errors.Wrap(err, "listing backups")
		}
		backupID = manifests[0].ID
		fmt.Fprintf(w, "Using most recent backup: %s\n", backupID)
	}

	// Get backup details for confirmation message
	manifest, err := mgr.Get(platform.Name(), backupID)
	if err != nil {
		return errors.Wrapf(err, "getting backup %s", backupID)
	}

	fmt.Fprintf(w, "Restoring %d files from backup %s...\n", len(manifest.Files), backupID)

	// Perform restore
	if err := mgr.Restore(platform.Name(), backupID); err != nil {
		return errors.Wrap(err, "restoring backup")
	}

	fmt.Fprintf(w, "%s✓ Restored %s configuration from backup %s%s\n",
		colorGreen, platform.DisplayName(), backupID, colorReset)

	return nil
}
