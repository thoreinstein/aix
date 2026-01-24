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

func init() {
	Cmd.AddCommand(createCmd)
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a manual backup",
	Long: `Create a backup of platform configuration files.

Backups are created automatically before aix modifies configurations.
This command allows you to create additional backups manually.

By default, creates backups for all detected platforms. Use the --platform
flag to limit to specific platforms.`,
	Example: `  # Create backup for all platforms
  aix backup create

  # Create backup for a specific platform
  aix backup create --platform claude

  See Also:
    aix backup list    - List available backups
    aix backup restore - Restore from a backup`,
	RunE: runCreate,
}

func runCreate(_ *cobra.Command, _ []string) error {
	return runCreateWithWriter(os.Stdout)
}

func runCreateWithWriter(w io.Writer) error {
	platforms, err := cli.ResolvePlatforms(flags.GetPlatformFlag())
	if err != nil {
		return errors.Wrap(err, "resolving platforms")
	}

	mgr := backup.NewManager()
	created := 0

	for _, p := range platforms {
		paths := p.BackupPaths()
		if len(paths) == 0 {
			fmt.Fprintf(w, "%s%s: no paths configured for backup%s\n",
				colorYellow, p.DisplayName(), colorReset)
			continue
		}

		manifest, err := mgr.Backup(p.Name(), paths)
		if err != nil {
			if errors.Is(err, errors.New("no files to back up")) {
				fmt.Fprintf(w, "%s%s: no files found to back up%s\n",
					colorYellow, p.DisplayName(), colorReset)
				continue
			}
			return errors.Wrapf(err, "backing up %s", p.Name())
		}

		fmt.Fprintf(w, "%s✓ %s: created backup %s (%d files)%s\n",
			colorGreen, p.DisplayName(), manifest.ID, len(manifest.Files), colorReset)
		created++
	}

	if created == 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "No backups created. Configurations may not exist yet.")
	}

	return nil
}
