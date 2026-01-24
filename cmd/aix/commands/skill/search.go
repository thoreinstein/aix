package skill

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// Package-level flag variables for search command.
var (
	searchRepo string
	searchJSON bool
)

// searchCmd is the skill search subcommand.
var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for skills in repositories",
	Long:  `Search for skills across all configured repositories by name or description.`,
	Example: `  # Search for skills matching "test"
  aix skill search test

  # Search within a specific repository
  aix skill search test --repo=my-repo

  # Output results as JSON
  aix skill search test --json`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

func init() {
	searchCmd.Flags().StringVar(&searchRepo, "repo", "", "Filter by repository name")
	searchCmd.Flags().BoolVar(&searchJSON, "json", false, "Output in JSON format")
	Cmd.AddCommand(searchCmd)
}

// runSearch executes the skill search command.
func runSearch(_ *cobra.Command, args []string) error {
	return runSearchWithWriter(os.Stdout, args)
}

// runSearchWithWriter allows injecting a writer for testing.
func runSearchWithWriter(w io.Writer, _ []string) error {
	fmt.Fprintln(w, "skill search: not yet implemented")
	return nil
}
