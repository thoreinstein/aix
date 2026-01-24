package command

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/config"
	"github.com/thoreinstein/aix/internal/repo"
	"github.com/thoreinstein/aix/internal/resource"
)

// Package-level flag variables for search command.
var (
	searchRepo string
	searchJSON bool
)

// searchResultJSON represents a command search result for JSON output.
type searchResultJSON struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Repository  string `json:"repository"`
	Path        string `json:"path"`
}

// searchCmd is the command search subcommand.
var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for commands in repositories",
	Long:  `Search for commands across all configured repositories by name or description.`,
	Example: `  # Search for commands matching "test"
  aix command search test

  # Search within a specific repository
  aix command search test --repo=my-repo

  # Output results as JSON
  aix command search test --json`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

func init() {
	searchCmd.Flags().StringVar(&searchRepo, "repo", "", "Filter by repository name")
	searchCmd.Flags().BoolVar(&searchJSON, "json", false, "Output in JSON format")
	Cmd.AddCommand(searchCmd)
}

// runSearch executes the command search command.
func runSearch(_ *cobra.Command, args []string) error {
	return runSearchWithWriter(os.Stdout, args)
}

// runSearchWithWriter allows injecting a writer for testing.
func runSearchWithWriter(w io.Writer, args []string) error {
	query := args[0]

	// Get repo list
	configPath := config.DefaultConfigPath()
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

	// Validate --repo filter against known repositories.
	if searchRepo != "" {
		found := false
		for _, r := range repos {
			if r.Name == searchRepo {
				found = true
				break
			}
		}
		if !found {
			return errors.Errorf("repository %q not found; run 'aix repo list' to see available repositories", searchRepo)
		}
	}

	// Scan all repos
	scanner := resource.NewScanner()
	resources, err := scanner.ScanAll(repos)
	if err != nil {
		return errors.Wrap(err, "scanning repositories")
	}

	// Build search options - filter to commands only
	opts := resource.SearchOptions{
		Type: resource.TypeCommand,
	}
	if searchRepo != "" {
		opts.RepoName = searchRepo
	}

	// Search
	results := resource.Search(resources, query, opts)

	if len(results) == 0 {
		fmt.Fprintf(w, "No commands found matching %q\n", query)
		return nil
	}

	// Output results
	if searchJSON {
		return outputSearchJSON(w, results)
	}
	return outputSearchTable(w, results)
}

// outputSearchJSON writes search results as JSON.
func outputSearchJSON(w io.Writer, results []resource.Resource) error {
	jsonResults := make([]searchResultJSON, 0, len(results))
	for _, r := range results {
		jsonResults = append(jsonResults, searchResultJSON{
			Name:        r.Name,
			Description: r.Description,
			Repository:  r.RepoName,
			Path:        r.Path,
		})
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(jsonResults)
}

// outputSearchTable writes search results as a formatted table.
func outputSearchTable(w io.Writer, results []resource.Resource) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "%sNAME%s\t%sDESCRIPTION%s\t%sREPOSITORY%s\n",
		colorBold, colorReset, colorBold, colorReset, colorBold, colorReset)

	for _, r := range results {
		fmt.Fprintf(tw, "%s%s%s\t%s\t%s\n",
			colorGreen, r.Name, colorReset,
			truncate(r.Description, 60),
			r.RepoName)
	}

	return tw.Flush()
}
