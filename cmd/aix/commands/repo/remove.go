package repo

import (
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/config"
	aixerrors "github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/internal/repo"
)

func init() {
	Cmd.AddCommand(removeCmd)
}

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a repository source",
	Long: `Remove a repository from the configured sources.

This removes both the configuration entry and the cached clone.`,
	Example: `  aix repo remove community-skills`,
	Args:    cobra.ExactArgs(1),
	RunE:    runRemove,
}

func runRemove(_ *cobra.Command, args []string) error {
	name := args[0]

	configPath := config.DefaultConfigPath()
	manager := repo.NewManager(configPath)

	if err := manager.Remove(name); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return aixerrors.NewUserError(
				errors.Newf("repository %q not found", name),
				"Run: aix repo list to see available repositories",
			)
		}
		// Cache cleanup failure is a warning, not a fatal error
		if errors.Is(err, repo.ErrCacheCleanupFailed) {
			fmt.Printf("✓ Repository %q removed\n", name)
			fmt.Printf("⚠ Warning: %v\n", err)
			return nil
		}
		return aixerrors.NewSystemError(
			errors.Wrapf(err, "removing repository %q", name),
			"Check file permissions on the cache directory",
		)
	}

	fmt.Printf("✓ Repository %q removed\n", name)
	return nil
}
