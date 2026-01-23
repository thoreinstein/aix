// Package commands provides CLI commands for the aix tool.
package commands

import "github.com/thoreinstein/aix/cmd/aix/commands/repo"

func init() {
	rootCmd.AddCommand(repo.Cmd)
}
