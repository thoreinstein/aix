// Package repo provides CLI commands for managing skill repositories.
package repo

import "github.com/spf13/cobra"

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
