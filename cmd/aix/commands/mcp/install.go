package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/cmd/aix/commands/flags"
	"github.com/thoreinstein/aix/internal/backup"
	"github.com/thoreinstein/aix/internal/cli"
	cliprompt "github.com/thoreinstein/aix/internal/cli/prompt"
	"github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/internal/git"
	"github.com/thoreinstein/aix/internal/mcp"
	mcpvalidator "github.com/thoreinstein/aix/internal/mcp/validator"
	"github.com/thoreinstein/aix/internal/resource"
	"github.com/thoreinstein/aix/internal/validator"
)

var (
	installForce bool
	installFile  bool
)

func init() {
	installCmd.Flags().BoolVar(&installForce, "force", false,
		"overwrite existing MCP server without confirmation")
	installCmd.Flags().BoolVarP(&installFile, "file", "f", false,
		"treat argument as a file path instead of searching repos")
	Cmd.AddCommand(installCmd)
}

var installCmd = &cobra.Command{
	Use:   "install <source>",
	Short: "Install an MCP server from a repository, local file, or git URL",
	Long: `Install an MCP server configuration from a configured repository, local JSON file, or git URL.

The source can be:
  - An MCP server name to search in configured repositories
  - A local path to a .json MCP configuration file
  - A git URL (https://, git@, or .git suffix) containing mcp/*.json files

When given a name (not a path), aix searches configured repositories first.
If the server exists in multiple repositories, you will be prompted to select one.
Use --file to skip repo search and treat the argument as a file path.

For git URLs, the repository is cloned to a temporary directory, MCP servers
are discovered in the mcp/ directory, and you select which to install.`,
	Example: `  # Install by name from configured repos
  aix mcp install github-mcp

  # Install from local JSON file
  aix mcp install ./servers/github.json
  aix mcp install --file github.json  # Force file path interpretation

  # Install from absolute path
  aix mcp install /path/to/server.json

  # Install from git repository
  aix mcp install https://github.com/user/mcp-servers.git

  # Force overwrite existing server
  aix mcp install github-mcp --force`,
	Args: cobra.ExactArgs(1),
	RunE: runInstall,
}

// Sentinel errors for install operations.
var errInstallFailed = errors.New("installation failed")

func runInstall(_ *cobra.Command, args []string) error {
	source := args[0]

	// If --file flag is set, treat argument as file path or URL (old behavior)
	if installFile {
		if git.IsURL(source) {
			return installFromGit(source)
		}
		return installFromLocal(source)
	}

	// If source is clearly a path or URL, use direct install
	if git.IsURL(source) || looksLikePath(source) {
		if git.IsURL(source) {
			return installFromGit(source)
		}
		return installFromLocal(source)
	}

	// Try repo lookup first
	matches, err := resource.FindByName(source, resource.TypeMCP)
	if err != nil {
		if errors.Is(err, resource.ErrNoReposConfigured) {
			return errors.New("no repositories configured. Run 'aix repo add <url>' to add one")
		}
		return errors.Wrap(err, "searching repositories")
	}

	if len(matches) > 0 {
		return installFromRepo(source, matches)
	}

	// No matches in repos - check if input might be a forgotten path
	if mightBePath(source) {
		if _, statErr := os.Stat(source); statErr == nil {
			return errors.Newf("MCP server %q not found in repositories, but a local file exists at that path.\nDid you mean: aix mcp install --file %s", source, source)
		}
	}

	// Check if it's a local path that exists
	if _, err := os.Stat(source); err == nil {
		return installFromLocal(source)
	}

	return errors.Newf("MCP server %q not found in any configured repository", source)
}

// looksLikePath returns true if the source appears to be a file path.
func looksLikePath(source string) bool {
	// Starts with ./ or ../ or /
	if strings.HasPrefix(source, "./") || strings.HasPrefix(source, "../") || strings.HasPrefix(source, "/") {
		return true
	}
	// Contains path separator
	if strings.Contains(source, string(filepath.Separator)) {
		return true
	}
	// On Windows, also check for backslash
	if filepath.Separator != '/' && strings.Contains(source, "/") {
		return true
	}
	return false
}

// mightBePath returns true if the input might be a path the user forgot the --file flag for.
func mightBePath(s string) bool {
	// Ends with .json extension
	if strings.HasSuffix(strings.ToLower(s), ".json") {
		return true
	}
	// Contains backslash (Windows paths, even on Unix)
	if strings.Contains(s, "\\") {
		return true
	}
	return false
}

// installFromRepo installs an MCP server from a configured repository.
func installFromRepo(name string, matches []resource.Resource) error {
	var selected *resource.Resource

	if len(matches) == 1 {
		selected = &matches[0]
	} else {
		// Multiple matches - prompt user to select
		choice, err := cliprompt.SelectResourceDefault(name, matches)
		if err != nil {
			return errors.Wrap(err, "selecting resource")
		}
		selected = choice
	}

	// Copy from cache to temp directory
	tempDir, err := resource.CopyToTemp(selected)
	if err != nil {
		return errors.Wrap(err, "copying from repository cache")
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to clean up temp dir: %v\n", removeErr)
		}
	}()

	fmt.Printf("Installing from repository: %s\n", selected.RepoName)

	// For MCP resources (flat files), the file is copied to tempDir with its original name
	// We need to find the JSON file in the temp directory
	tempFile := filepath.Join(tempDir, filepath.Base(selected.Path))
	return installFromLocal(tempFile)
}

// installFromGit clones a git repository and installs MCP servers from it.
func installFromGit(url string) error {
	fmt.Println("Cloning repository...")

	// Create temp directory for clone
	tempDir, err := os.MkdirTemp("", "aix-mcp-*")
	if err != nil {
		return errors.Wrap(err, "creating temp directory")
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to clean up temp dir: %v\n", removeErr)
		}
	}()

	// Clone the repository
	if err := git.Clone(url, tempDir, 1); err != nil {
		return errors.Wrap(err, "cloning repository")
	}

	// Look for mcp/*.json files
	mcpDir := filepath.Join(tempDir, "mcp")
	entries, err := os.ReadDir(mcpDir)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("no mcp/ directory found in repository")
		}
		return errors.Wrap(err, "reading mcp directory")
	}

	var jsonFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			jsonFiles = append(jsonFiles, filepath.Join(mcpDir, entry.Name()))
		}
	}

	if len(jsonFiles) == 0 {
		return errors.New("no MCP server configurations (*.json) found in mcp/ directory")
	}

	// If single file, install it directly
	if len(jsonFiles) == 1 {
		return installFromLocal(jsonFiles[0])
	}

	// Multiple files - prompt user to select
	fmt.Printf("Found %d MCP server configurations:\n", len(jsonFiles))
	for i, f := range jsonFiles {
		fmt.Printf("  [%d] %s\n", i+1, filepath.Base(f))
	}
	fmt.Print("Select server to install (1-", len(jsonFiles), "): ")

	var choice int
	if _, err := fmt.Scanf("%d", &choice); err != nil || choice < 1 || choice > len(jsonFiles) {
		return errors.New("invalid selection")
	}

	return installFromLocal(jsonFiles[choice-1])
}

// installFromLocal installs an MCP server from a local JSON file.
func installFromLocal(serverPath string) error {
	// Resolve to absolute path for consistent error messages
	absPath, err := filepath.Abs(serverPath)
	if err != nil {
		absPath = serverPath
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return errors.Newf("MCP server file not found: %s", absPath)
	}

	// Check if it's a JSON file
	if !strings.HasSuffix(strings.ToLower(absPath), ".json") {
		return errors.Newf("expected .json file, got: %s", absPath)
	}

	fmt.Println("Validating MCP server configuration...")

	// Read and parse the JSON file
	data, err := os.ReadFile(absPath)
	if err != nil {
		return errors.Wrap(err, "reading server file")
	}

	var server mcp.Server
	if err := json.Unmarshal(data, &server); err != nil {
		return errors.Wrap(err, "parsing server JSON")
	}

	// Derive name from filename if not set in JSON
	if server.Name == "" {
		server.Name = strings.TrimSuffix(filepath.Base(absPath), ".json")
	}

	// Validate the server configuration by wrapping in a Config
	cfg := &mcp.Config{
		Servers: map[string]*mcp.Server{
			server.Name: &server,
		},
	}

	v := mcpvalidator.New()
	result := v.Validate(cfg)
	if result.HasErrors() {
		fmt.Println("MCP server validation failed:")
		reporter := validator.NewReporter(os.Stdout, validator.FormatText)
		_ = reporter.Report(result)
		return errInstallFailed
	}

	// Print warnings if any
	if result.HasWarnings() {
		for _, w := range result.Warnings() {
			fmt.Printf("  Warning: %s\n", w.Message)
		}
	}

	// Get target platforms
	platforms, err := cli.ResolvePlatforms(flags.GetPlatformFlag())
	if err != nil {
		return errors.Wrap(err, "resolving platforms")
	}

	// Check for existing servers (unless --force)
	if !installForce {
		for _, plat := range platforms {
			if _, err := plat.GetMCP(server.Name); err == nil {
				return errors.Newf("MCP server %q already exists on %s (use --force to overwrite)",
					server.Name, plat.DisplayName())
			}
		}
	}

	// Determine transport type
	transport := server.Transport
	if transport == "" {
		if server.URL != "" {
			transport = "sse"
		} else {
			transport = "stdio"
		}
	}

	// Install to each platform
	var installedCount int
	for _, plat := range platforms {
		// Ensure backup exists before modifying
		if err := backup.EnsureBackedUp(plat.Name(), plat.BackupPaths()); err != nil {
			return errors.Wrapf(err, "backing up %s before install", plat.DisplayName())
		}

		fmt.Printf("Installing '%s' to %s... ", server.Name, plat.DisplayName())

		// Set package-level variables that addMCPToPlatform expects
		mcpAddURL = server.URL
		mcpAddPlatforms = server.Platforms

		// Use the existing addMCPToPlatform function
		if err := addMCPToPlatform(plat, server.Name, server.Command, server.Args, transport, server.Env, server.Headers); err != nil {
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
	fmt.Printf("[OK] MCP server '%s' installed to %d %s\n", server.Name, installedCount, platformWord)

	return nil
}
