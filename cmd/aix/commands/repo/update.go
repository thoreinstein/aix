package repo

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/config"
	aixerrors "github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/internal/repo"
)

func init() {
	Cmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "Update repository sources",
	Long: `Update repository sources by pulling latest changes.

If a name is provided, only that repository is updated.
If no name is provided, all repositories are updated.`,
	Example: `  # Update all repositories
  aix repo update

  # Update specific repository
  aix repo update community-skills`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUpdate,
}

func runUpdate(_ *cobra.Command, args []string) error {
	return runUpdateWithIO(args, os.Stdout)
}

// runUpdateWithIO allows injecting a writer for testing.
func runUpdateWithIO(args []string, w io.Writer) error {
	var name string
	if len(args) > 0 {
		name = args[0]
	}

	// Get config path
	configPath := config.DefaultConfigPath()

	// Create manager
	manager := repo.NewManager(configPath)

	// Update specific repository
	if name != "" {
		fmt.Fprintf(w, "Updating %s... ", name)
		if err := manager.Update(name); err != nil {
			fmt.Fprintln(w, "\u2717 failed")
			return handleUpdateError(name, err)
		}
		fmt.Fprintln(w, "\u2713 done")

		// Validate repository content and show warnings
		repoConfig, err := manager.Get(name)
		if err == nil {
			warnings := repo.ValidateRepoContent(repoConfig.Path)
			printValidationWarnings(w, warnings)
		}
		return nil
	}

	// Update all repositories
	repos, err := manager.List()
	if err != nil {
		return fmt.Errorf("listing repositories: %w", err)
	}

	if len(repos) == 0 {
		fmt.Fprintln(w, "No repositories configured.")
		return nil
	}

	var failed []string
	var allWarnings []repo.ValidationWarning
	for _, r := range repos {
		fmt.Fprintf(w, "Updating %s... ", r.Name)
		// Use UpdateByPath to avoid redundant config reload
		if err := manager.UpdateByPath(r.Path); err != nil {
			fmt.Fprintln(w, "\u2717 failed")
			failed = append(failed, fmt.Sprintf("%s: %v", r.Name, err))
			continue
		}
		fmt.Fprintln(w, "\u2713 done")

		// Collect validation warnings
		warnings := repo.ValidateRepoContent(r.Path)
		allWarnings = append(allWarnings, warnings...)
	}

	// Print all validation warnings at the end
	printValidationWarnings(w, allWarnings)

	if len(failed) > 0 {
		return fmt.Errorf("some repositories failed to update:\n  %s", joinErrors(failed))
	}

	return nil
}

// handleUpdateError returns a user-friendly error message for known error types.
func handleUpdateError(name string, err error) error {
	if errors.Is(err, repo.ErrNotFound) {
		return aixerrors.NewUserError(
			errors.Newf("repository '%s' not found", name),
			"Run: aix repo list to see available repositories",
		)
	}
	return aixerrors.NewSystemError(
		errors.Wrapf(err, "updating '%s'", name),
		"Check your network connection and repository access",
	)
}

// joinErrors joins error strings with newline and indentation.
func joinErrors(errs []string) string {
	return strings.Join(errs, "\n  ")
}
