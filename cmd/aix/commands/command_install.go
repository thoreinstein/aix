package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/command/parser"
	"github.com/thoreinstein/aix/internal/command/validator"
	"github.com/thoreinstein/aix/internal/platform/claude"
	"github.com/thoreinstein/aix/internal/platform/opencode"
)

// commandInstallForce enables overwriting existing commands without confirmation.
var commandInstallForce bool

// Sentinel errors for command install operations.
var errCommandInstallFailed = errors.New("command installation failed")

func init() {
	commandInstallCmd.Flags().BoolVarP(&commandInstallForce, "force", "f", false,
		"overwrite existing command without confirmation")
	commandCmd.AddCommand(commandInstallCmd)
}

var commandInstallCmd = &cobra.Command{
	Use:   "install <source>",
	Short: "Install a slash command from a local path or git URL",
	Long: `Install a slash command from a local file, directory, or git repository.

The source can be:
  - A local .md file containing the command definition
  - A directory containing a command.md file or any .md file
  - A git URL (https://, git@, or .git suffix)

For git URLs, the repository is cloned to a temporary directory, the command
is installed, and the temporary directory is cleaned up.`,
	Example: `  # Install from a local file
  aix command install ./review.md

  # Install from a directory
  aix command install ./my-command/

  # Install from a git repository
  aix command install https://github.com/user/my-command.git
  aix command install git@github.com:user/my-command.git

  # Force overwrite existing command
  aix command install ./review.md --force

  # Install to a specific platform
  aix command install ./review.md --platform claude

  See Also:
    aix command remove   - Remove an installed command
    aix command init     - Create a new command`,
	Args: cobra.ExactArgs(1),
	RunE: runCommandInstall,
}

func runCommandInstall(_ *cobra.Command, args []string) error {
	source := args[0]

	// Determine if source is a git URL or local path
	if isCommandGitURL(source) {
		return installCommandFromGit(source)
	}

	return installCommandFromLocal(source)
}

// isCommandGitURL returns true if the source looks like a git repository URL.
func isCommandGitURL(source string) bool {
	if strings.Contains(source, "://") {
		return true
	}
	if strings.HasSuffix(source, ".git") {
		return true
	}
	if strings.HasPrefix(source, "git@") {
		return true
	}
	return false
}

// installCommandFromGit clones a git repository and installs the command from it.
func installCommandFromGit(url string) error {
	fmt.Println("Cloning repository...")

	// Create temp directory for clone
	tempDir, err := os.MkdirTemp("", "aix-command-*")
	if err != nil {
		return errors.Wrap(err, "creating temp directory")
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to clean up temp dir: %v\n", removeErr)
		}
	}()

	// Clone the repository
	cmd := exec.Command("git", "clone", "--depth=1", url, tempDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "cloning repository")
	}

	return installCommandFromLocal(tempDir)
}

// installCommandFromLocal installs a command from a local file or directory.
func installCommandFromLocal(source string) error {
	// Resolve to absolute path for consistent error messages
	absPath, err := filepath.Abs(source)
	if err != nil {
		return errors.Wrap(err, "resolving path")
	}

	// Determine command file path
	commandPath := absPath
	info, err := os.Stat(absPath)
	if err != nil {
		return errors.Wrap(err, "accessing source")
	}

	if info.IsDir() {
		// Look for command.md in directory first
		candidatePath := filepath.Join(absPath, "command.md")
		if _, err := os.Stat(candidatePath); err == nil {
			commandPath = candidatePath
		} else {
			// Fall back to finding any .md file that doesn't start with _
			commandPath = ""
			entries, readErr := os.ReadDir(absPath)
			if readErr != nil {
				return errors.Wrap(readErr, "reading directory")
			}
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				name := e.Name()
				if strings.HasSuffix(name, ".md") && !strings.HasPrefix(name, "_") {
					commandPath = filepath.Join(absPath, name)
					break
				}
			}
			if commandPath == "" {
				return errors.Newf("no command file found in %s (expected command.md or any .md file)", absPath)
			}
		}
	}

	// Verify file exists
	if _, err := os.Stat(commandPath); err != nil {
		return errors.Newf("command file not found: %s", commandPath)
	}

	fmt.Println("Validating command...")

	// Parse command using claude.Command as the canonical type
	p := parser.New[*claude.Command]()
	cmd, err := p.ParseFile(commandPath)
	if err != nil {
		return errors.Wrap(err, "parsing command")
	}

	// Validate command
	v := validator.New()
	result := v.Validate(*cmd, commandPath)
	if result.HasErrors() {
		fmt.Println("Command validation failed:")
		for _, e := range result.Errors {
			fmt.Printf("  - %v\n", e)
		}
		return errCommandInstallFailed
	}

	// Print warnings (but don't fail)
	for _, w := range result.Warnings {
		fmt.Printf("  ⚠ %s\n", w.Message)
	}

	// Get target platforms
	platforms, err := cli.ResolvePlatforms(GetPlatformFlag())
	if err != nil {
		return err
	}

	// Check for existing commands (unless --force)
	if !commandInstallForce {
		for _, plat := range platforms {
			if _, err := plat.GetCommand((*cmd).Name); err == nil {
				return errors.Newf("command %q already exists on %s (use --force to overwrite)",
					(*cmd).Name, plat.DisplayName())
			}
		}
	}

	// Install to each platform
	var installedCount int
	for _, plat := range platforms {
		fmt.Printf("Installing '%s' to %s... ", (*cmd).Name, plat.DisplayName())

		// Convert command to platform-specific type
		platformCmd := convertCommandForPlatform(*cmd, plat.Name())

		if err := plat.InstallCommand(platformCmd); err != nil {
			fmt.Println("failed")
			return errors.Wrapf(err, "failed to install to %s", plat.DisplayName())
		}

		fmt.Println("done")
		installedCount++
	}

	// Print summary
	platformWord := "platform"
	if installedCount != 1 {
		platformWord = "platforms"
	}
	fmt.Printf("✓ Command '%s' installed to %d %s\n", (*cmd).Name, installedCount, platformWord)

	return nil
}

// convertCommandForPlatform converts a canonical claude.Command to the appropriate
// platform-specific command type.
func convertCommandForPlatform(cmd *claude.Command, platformName string) any {
	switch platformName {
	case "claude":
		// Claude uses the canonical format, return as-is
		return cmd
	case "opencode":
		// Convert to OpenCode command format
		return convertToOpenCodeCommand(cmd)
	default:
		// Unknown platform, return as-is and let the adapter handle it
		return cmd
	}
}

// convertToOpenCodeCommand converts a Claude command to an OpenCode command.
// Note: OpenCode has a simpler command model, so some fields are lost in translation:
//   - ArgumentHint: not supported
//   - DisableModelInvocation: not supported
//   - UserInvocable: not supported
//   - AllowedTools: not supported
//   - Context: not supported
//   - Hooks: not supported
func convertToOpenCodeCommand(c *claude.Command) *opencode.Command {
	return &opencode.Command{
		Name:         c.Name,
		Description:  c.Description,
		Agent:        c.Agent,
		Model:        c.Model,
		Instructions: c.Instructions,
	}
}
