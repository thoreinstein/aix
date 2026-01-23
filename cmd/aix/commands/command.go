package commands

import "github.com/thoreinstein/aix/cmd/aix/commands/command"

func init() {
	rootCmd.AddCommand(command.Cmd)
}
