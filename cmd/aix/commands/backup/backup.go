// Package backup provides CLI commands for managing configuration backups.
package backup

import "github.com/spf13/cobra"

// Color constants for terminal output.
const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorGray   = "\033[90m"
)

// Cmd is the root backup command.
var Cmd = &cobra.Command{
	Use:   "backup",
	Short: "Manage configuration backups",
	Long: `Manage configuration backups for AI coding assistant platforms.

Before aix modifies platform configurations, it automatically creates backups.
This command group allows you to list, restore, create, and prune backups.

Backups are stored in ~/.config/aix/backups/ organized by platform.`,
	Example: `  # List all backups
  aix backup list

  # List backups for a specific platform
  aix backup list --platform claude

  # Restore from the most recent backup
  aix backup restore --platform claude

  # Restore from a specific backup
  aix backup restore 20260123T100712 --platform claude

  # Create a manual backup
  aix backup create --platform claude

  # Remove old backups, keeping the 3 most recent
  aix backup prune --keep 3

  See Also:
    aix backup list    - List available backups
    aix backup restore - Restore from a backup
    aix backup create  - Manually create a backup
    aix backup prune   - Remove old backups`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	},
}
