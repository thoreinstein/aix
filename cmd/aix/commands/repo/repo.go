// Package repo provides CLI commands for managing skill repositories.
package repo

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/repo"
)

// Cmd is the root repo command.
var Cmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage skill repositories",
	Long: `Manage skill repositories for discovering and installing resources.

Repositories are Git repositories that follow the standard aix directory structure,
containing skills/, commands/, agents/, or mcp/ subdirectories.

Repositories are shallow cloned to a local cache for efficient discovery and installation.`,
	Example: `  # Add a repository
  aix repo add https://github.com/example/skills-repo.git

  # List configured repositories
  aix repo list

  # Update all repositories
  aix repo update

  # Remove a repository
  aix repo remove community-skills

  See Also:
    aix repo add    - Add a repository source
    aix repo list   - List configured repositories
    aix repo update - Update repository caches
    aix repo remove - Remove a repository`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	},
}

// printValidationWarnings outputs validation warnings to the writer.
// It filters out "directory not found" warnings which are expected for repos
// that only contain certain resource types.
func printValidationWarnings(w io.Writer, warnings []repo.ValidationWarning) {
	// Filter to only show actionable warnings (not missing optional directories)
	var actionable []repo.ValidationWarning
	for _, warn := range warnings {
		if warn.Message != "directory not found" {
			actionable = append(actionable, warn)
		}
	}

	if len(actionable) == 0 {
		return
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "⚠ Validation warnings:")
	for _, warn := range actionable {
		fmt.Fprintf(w, "  %s: %s\n", warn.Path, warn.Message)
	}
}
