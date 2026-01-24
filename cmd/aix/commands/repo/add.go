package repo

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/config"
	"github.com/thoreinstein/aix/internal/errors"
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
	return runAddWithIO(args, os.Stdout)
}

// runAddWithIO allows injecting a writer for testing.
func runAddWithIO(args []string, w io.Writer) error {
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

	// Add the repository with progress indicator
	fmt.Fprintf(w, "Cloning %s... ", url)
	repoConfig, err := manager.Add(url, opts...)
	if err != nil {
		fmt.Fprintln(w, "failed")
		return handleAddError(err)
	}
	fmt.Fprintln(w, "done")

	// Print success message
	fmt.Fprintf(w, "[OK] Repository '%s' added from %s\n", repoConfig.Name, url)
	fmt.Fprintf(w, "  Cached at: %s\n", repoConfig.Path)

	// Validate repository content and show warnings
	warnings := repo.ValidateRepoContent(repoConfig.Path)
	printValidationWarnings(w, warnings)

	return nil
}

// handleAddError returns a user-friendly error message for known error types.
func handleAddError(err error) error {
	switch {
	case errors.Is(err, repo.ErrInvalidURL):
		return errors.NewUserError(
			errors.New("invalid Git URL"),
			"Use HTTPS, SSH, or git:// protocol (e.g., https://github.com/org/repo.git)",
		)
	case errors.Is(err, repo.ErrNameCollision):
		return errors.NewUserError(
			err,
			"Run: aix repo list to see existing repositories\n       Use: --name <alternate-name> to specify a different name",
		)
	case errors.Is(err, repo.ErrInvalidName):
		return errors.NewUserError(
			errors.New("invalid repository name"),
			"Names must be lowercase alphanumeric with hyphens, starting with a letter (e.g., 'my-skills')",
		)
	default:
		return errors.NewSystemError(
			errors.Wrap(err, "failed to add repository"),
			"Check your network connection and Git credentials",
		)
	}
}
