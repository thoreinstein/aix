// Package commands provides CLI commands for the aix tool.
package commands

import "github.com/thoreinstein/aix/cmd/aix/commands/backup"

func init() {
	rootCmd.AddCommand(backup.Cmd)
}
