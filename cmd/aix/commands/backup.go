// Package commands provides CLI commands for the aix tool.
package commands

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(backupCmd)
}

var backupCmd = &cobra.Command{
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
