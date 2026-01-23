package repo

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/config"
	"github.com/thoreinstein/aix/internal/repo"
)

// Package-level flag variables for repo add command.
var nameFlag string

func init() {
	addCmd.Flags().StringVar(&nameFlag, "name", "", "custom name for the repository")
	Cmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:   "add <url>",
	Short: "Add a repository source",
	Long: `Add a Git repository as a source for skills, commands, and agents.

The repository is shallow cloned to the local cache. The repository name
is derived from the URL unless overridden with --name.`,
	Example: `  # Add from GitHub
  aix repo add https://github.com/example/community-skills.git

  # Add with custom name
  aix repo add https://github.com/example/skills.git --name my-skills

  # Add from private repo (SSH)
  aix repo add git@github.com:org/private-skills.git`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

func runAdd(_ *cobra.Command, args []string) error {
	url := args[0]

	// Get config path
	configPath := config.DefaultConfigPath()

	// Create manager
	manager := repo.NewManager(configPath)

	// Build options
	var opts []repo.Option
	if nameFlag != "" {
		opts = append(opts, repo.WithName(nameFlag))
	}

	// Add the repository
	repoConfig, err := manager.Add(url, opts...)
	if err != nil {
		return handleAddError(err)
	}

	// Print success message
	fmt.Printf("✓ Repository '%s' added from %s\n", repoConfig.Name, url)
	fmt.Printf("  Cached at: %s\n", repoConfig.Path)

	return nil
}

// Sentinel errors for repo add command.
var (
	errInvalidGitURL   = errors.New("invalid Git URL")
	errInvalidRepoName = errors.New("invalid repository name")
)

// handleAddError returns a user-friendly error message for known error types.
func handleAddError(err error) error {
	switch {
	case errors.Is(err, repo.ErrInvalidURL):
		return errInvalidGitURL
	case errors.Is(err, repo.ErrNameCollision):
		return fmt.Errorf("repository '%s' already exists", nameFlag)
	case errors.Is(err, repo.ErrInvalidName):
		return errInvalidRepoName
	default:
		return err
	}
}
