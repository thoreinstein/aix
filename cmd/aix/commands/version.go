package commands

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/paths"
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
	Long:  `Print the version, commit, build date, Go version, and detected platforms.`,
	Example: `  # Print version info
  aix version

See Also: aix status, aix doctor`,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("aix version %s\n", Version)
		fmt.Printf("  commit:    %s\n", Commit)
		fmt.Printf("  built:     %s\n", Date)
		fmt.Printf("  go:        %s\n", runtime.Version())
		fmt.Println("  platforms:")
		for _, p := range paths.Platforms() {
			status := "not installed"
			if platformInstalled(p) {
				status = "installed"
			}
			fmt.Printf("    %-9s %s\n", p+":", status)
		}
	},
}

// platformInstalled checks if a platform's config directory exists.
func platformInstalled(platform string) bool {
	dir := paths.GlobalConfigDir(platform)
	if dir == "" {
		return false
	}
	_, err := os.Stat(dir)
	return err == nil
}
