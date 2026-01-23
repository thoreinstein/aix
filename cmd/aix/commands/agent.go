package commands

import "github.com/thoreinstein/aix/cmd/aix/commands/agent"

func init() {
	rootCmd.AddCommand(agent.Cmd)
}
