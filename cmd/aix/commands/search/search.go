// Package search provides the search command for finding resources across cached repositories.
package search

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/config"
	"github.com/thoreinstein/aix/internal/paths"
	"github.com/thoreinstein/aix/internal/repo"
	"github.com/thoreinstein/aix/internal/resource"
)

// ANSI color codes for terminal output.
const (
	colorReset = "\033[0m"
	colorBold  = "\033[1m"
	colorGreen = "\033[32m"
	colorGray  = "\033[90m"
)

var (
	typeFilter string
	repoFilter string
	jsonOutput bool
)

func init() {
	Cmd.Flags().StringVar(&typeFilter, "type", "", "Filter by resource type (skill, command, agent, mcp)")
	Cmd.Flags().StringVar(&repoFilter, "repo", "", "Filter by repository name")
	Cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
}

// Cmd is the search command.
var Cmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for resources across cached repositories",
	Long: `Search for skills, commands, agents, and MCP servers across all cached repositories.

The search is case-insensitive and matches against resource names and descriptions.
Results are sorted by match quality: exact name matches first, then prefix matches,
then substring matches, then description-only matches.

If no query is provided, all resources are listed (subject to filters).`,
	Example: `  # Search for resources containing "deploy"
  aix search deploy

  # Search for skills only
  aix search --type=skill

  # Search in a specific repository
  aix search --repo=official deploy

  # Output as JSON
  aix search deploy --json

  # List all resources
  aix search`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSearch,
}

func runSearch(_ *cobra.Command, args []string) error {
	return runSearchWithWriter(os.Stdout, args)
}

// runSearchWithWriter allows injecting a writer for testing.
func runSearchWithWriter(w io.Writer, args []string) error {
	// Get the query (optional)
	var query string
	if len(args) > 0 {
		query = args[0]
	}

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

	// Build search options
	opts := resource.SearchOptions{
		Type:     resource.ResourceType(typeFilter),
		RepoName: repoFilter,
	}

	// Search
	results := resource.Search(resources, query, opts)

	// Output
	if jsonOutput {
		return outputJSON(w, results)
	}
	return outputTabular(w, results)
}

// outputJSON outputs resources in JSON format.
func outputJSON(w io.Writer, resources []resource.Resource) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(resources)
}

// outputTabular outputs resources in a human-readable table format.
func outputTabular(w io.Writer, resources []resource.Resource) error {
	if len(resources) == 0 {
		fmt.Fprintln(w, "No resources found.")
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "%sTYPE%s\t%sREPO%s\t%sNAME%s\t%sDESCRIPTION%s\n",
		colorBold, colorReset,
		colorBold, colorReset,
		colorBold, colorReset,
		colorBold, colorReset)

	for _, r := range resources {
		fmt.Fprintf(tw, "%s\t%s\t%s%s%s\t%s%s%s\n",
			r.Type,
			r.RepoName,
			colorGreen, r.Name, colorReset,
			colorGray, truncate(r.Description, 50), colorReset)
	}

	return tw.Flush()
}

// truncate shortens a string to maxLen characters, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
