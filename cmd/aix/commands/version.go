package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is the semantic version of the build.
	Version = "dev"
	// Commit is the git commit SHA of the build.
	Commit = "none"
	// Date is the build date.
	Date = "unknown"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Print the version, commit, and build date of aix.`,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("aix version %s\n", Version)
		fmt.Printf("  commit: %s\n", Commit)
		fmt.Printf("  built:  %s\n", Date)
	},
}
