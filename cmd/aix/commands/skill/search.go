package skill

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/config"
	"github.com/thoreinstein/aix/internal/paths"
	"github.com/thoreinstein/aix/internal/repo"
	"github.com/thoreinstein/aix/internal/resource"
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
func runSearchWithWriter(w io.Writer, args []string) error {
	query := args[0]

	// Get repo list
	configPath := filepath.Join(paths.ConfigHome(), config.AppName, "config.yaml")
	mgr := repo.NewManager(configPath)

	repos, err := mgr.List()
	if err != nil {
		return errors.Wrap(err, "listing repositories")
	}

	if len(repos) == 0 {
		fmt.Fprintln(w, "No repositories configured.")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Add a repository with:")
		fmt.Fprintln(w, "  aix repo add <url>")
		return nil
	}

	// Scan all repos
	scanner := resource.NewScanner()
	resources, err := scanner.ScanAll(repos)
	if err != nil {
		return errors.Wrap(err, "scanning repositories")
	}

	// Build search options - filter to skills only
	opts := resource.SearchOptions{
		Type: resource.TypeSkill,
	}
	if searchRepo != "" {
		opts.RepoName = searchRepo
	}

	// Search
	results := resource.Search(resources, query, opts)

	// Phase 2: just print the count of results found (formatting comes in Phase 3)
	fmt.Fprintf(w, "Found %d skill(s) matching %q\n", len(results), query)
	return nil
}
