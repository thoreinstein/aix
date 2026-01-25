package repo

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/config"
	"github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/internal/repo"
)

var listJSON bool

func init() {
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output in JSON format")
	Cmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured repository sources",
	Long:  `List all Git repositories configured as sources for skills, commands, and agents.`,
	Example: `  # List all repositories
  aix repo list

  # Output as JSON
  aix repo list --json

  See Also:
    aix repo add    - Add a repository source
    aix repo remove - Remove a repository`,
	Args: cobra.NoArgs,
	RunE: runList,
}

// repoJSON represents a repository in JSON output format.
type repoJSON struct {
	Name    string    `json:"name"`
	URL     string    `json:"url"`
	Path    string    `json:"path"`
	AddedAt time.Time `json:"added_at"`
}

func runList(_ *cobra.Command, _ []string) error {
	return runListWithWriter(os.Stdout, config.DefaultConfigPath())
}

// runListWithWriter allows injecting a writer for testing.
func runListWithWriter(w io.Writer, configPath string) error {
	mgr := repo.NewManager(configPath)

	repos, err := mgr.List()
	if err != nil {
		return errors.Wrap(err, "listing repositories")
	}

	if listJSON {
		return outputListJSON(w, repos)
	}
	return outputListTabular(w, repos)
}

// outputListJSON outputs repositories in JSON format.
func outputListJSON(w io.Writer, repos []config.RepoConfig) error {
	output := make([]repoJSON, len(repos))
	for i, r := range repos {
		output[i] = repoJSON{
			Name:    r.Name,
			URL:     r.URL,
			Path:    r.Path,
			AddedAt: r.AddedAt,
		}
	}

	// Sort by name for consistent output
	sort.Slice(output, func(i, j int) bool {
		return output[i].Name < output[j].Name
	})

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return errors.Wrap(enc.Encode(output), "encoding output")
}

// ANSI color codes.
const (
	colorReset = "\033[0m"
	colorBold  = "\033[1m"
	colorGreen = "\033[32m"
	colorGray  = "\033[90m"
)

// outputListTabular outputs repositories in tabular format.
func outputListTabular(w io.Writer, repos []config.RepoConfig) error {
	if len(repos) == 0 {
		fmt.Fprintln(w, "No repositories configured.")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Add a repository with:")
		fmt.Fprintln(w, "  aix repo add <url>")
		return nil
	}

	// Sort by name for consistent output
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Name < repos[j].Name
	})

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "%sNAME%s\t%sURL%s\t%sADDED%s\n",
		colorBold, colorReset,
		colorBold, colorReset,
		colorBold, colorReset)

	for _, r := range repos {
		fmt.Fprintf(tw, "%s%s%s\t%s\t%s%s%s\n",
			colorGreen, r.Name, colorReset,
			r.URL,
			colorGray, formatRelativeTime(r.AddedAt), colorReset)
	}

	return errors.Wrap(tw.Flush(), "flushing tabwriter")
}

// formatRelativeTime formats a time.Time as a human-readable relative time.
func formatRelativeTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}

	d := time.Since(t)

	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		mins := int(d.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case d < 24*time.Hour:
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case d < 30*24*time.Hour:
		weeks := int(d.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case d < 365*24*time.Hour:
		months := int(d.Hours() / (24 * 30))
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := int(d.Hours() / (24 * 365))
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}
