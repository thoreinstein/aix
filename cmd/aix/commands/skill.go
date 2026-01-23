package commands

import "github.com/thoreinstein/aix/cmd/aix/commands/skill"

func init() {
	rootCmd.AddCommand(skill.Cmd)
}
