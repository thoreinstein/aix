// Package commands implements the CLI commands for aix.
package commands

import (
	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/config"
)

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	config.Init()
	// Ignore load errors - defaults will be used if no config file
	_ = config.Load()
}

var rootCmd = &cobra.Command{
	Use:   "aix",
	Short: "Unified CLI for AI coding assistant configurations",
	Long: `aix is a unified CLI for managing AI coding assistant configurations
across multiple platforms including Claude Code, OpenCode, Codex CLI,
and Gemini CLI.

Write once, deploy everywhere. Define your configurations in a
platform-agnostic format and let aix handle the translation to each
platform's native format.`,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
