package commands

import "github.com/thoreinstein/aix/cmd/aix/commands/mcp"

func init() {
	rootCmd.AddCommand(mcp.Cmd)
}
