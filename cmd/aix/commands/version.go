package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/cmd"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Print the version, commit, and build date of aix.`,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("aix version %s\n", cmd.Version)
		fmt.Printf("  commit: %s\n", cmd.Commit)
		fmt.Printf("  built:  %s\n", cmd.Date)
	},
}
