package agent

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/cmd/aix/commands/flags"
	"github.com/thoreinstein/aix/internal/backup"
	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/internal/install"
	"github.com/thoreinstein/aix/internal/platform/claude"
	"github.com/thoreinstein/aix/internal/platform/opencode"
	"github.com/thoreinstein/aix/internal/resource"
	"github.com/thoreinstein/aix/pkg/frontmatter"
)

// Sentinel errors for agent install operations.
var (
	errAgentInstallFailed  = errors.New("failed to install agent to any platform")
	errAgentNameRequired   = errors.New("agent name is required")
	errAgentCollision      = errors.New("agent collision detected")
	errAgentInstallPartial = errors.New("partial installation failure")
)

// installForce enables overwriting existing agents without confirmation.
var installForce bool

// installFile forces treating the argument as a file path instead of searching repos.
var installFile bool

// installAllFromRepo installs all agents from a specific repository.
var installAllFromRepo string

var installer *install.Installer

func init() {
	installCmd.Flags().BoolVar(&installForce, "force", false,
		"overwrite existing agent without confirmation")
	installCmd.Flags().BoolVarP(&installFile, "file", "f", false,
		"treat argument as a file path instead of searching repos")
	installCmd.Flags().StringVar(&installAllFromRepo, "all-from-repo", "",
		"install all agents from a specific repository")
	flags.AddScopeFlag(installCmd)
	Cmd.AddCommand(installCmd)

	installer = install.NewInstaller(resource.TypeAgent, "agent", installFromLocal)
}

var installCmd = &cobra.Command{
	Use:   "install <source>",
	Short: "Install an agent from a repository or local path",
	Long: `Install an AI coding agent from a configured repository or local AGENT.md file.

The source can be:
  - An agent name to search in configured repositories
  - A path to an AGENT.md file
  - A directory containing an AGENT.md file

When given a name (not a path), aix searches configured repositories first.
If the agent exists in multiple repositories, you will be prompted to select one.
Use --file to skip repo search and treat the argument as a file path.

The AGENT.md file should contain YAML frontmatter with at least a 'name' field,
followed by the agent's instructions in markdown format.

Example AGENT.md:
  ---
  name: code-reviewer
  description: Reviews code for quality and best practices
  ---

  You are a code review expert. When reviewing code...`,
	Example: `  # Install by name from configured repos
  aix agent install code-reviewer

  # Install from a file
  aix agent install ./my-agent/AGENT.md
  aix agent install --file my-agent  # Force file path interpretation

  # Install from a directory
  aix agent install ./my-agent/

  # Install to specific platform
  aix agent install code-reviewer --platform claude

  # Force overwrite existing agent
  aix agent install code-reviewer --force

  # Install all agents from a specific repo
  aix agent install --all-from-repo official`,
	Args: func(cmd *cobra.Command, args []string) error {
		if installAllFromRepo != "" {
			if len(args) > 0 {
				return errors.New("cannot specify both --all-from-repo and a source argument")
			}
			return nil
		}
		if len(args) != 1 {
			return errors.New("requires exactly one argument (source)")
		}
		return nil
	},
	RunE: runInstall,
}

func runInstall(_ *cobra.Command, args []string) error {
	// Determine configuration scope
	scope, err := cli.DetermineScope(flags.GetScopeFlag())
	if err != nil {
		return fmt.Errorf("determining configuration scope: %w", err)
	}

	if installAllFromRepo != "" {
		if err := installer.InstallAllFromRepo(installAllFromRepo, scope); err != nil {
			return errors.Wrap(err, "installing all from repo")
		}
		return nil
	}

	source := args[0]

	// If --file flag is set, treat argument as file path (old behavior)
	if installFile {
		return installFromLocal(source, scope)
	}

	// If source is clearly a path, use direct install
	if install.LooksLikePath(source) {
		return installFromLocal(source, scope)
	}

	// Try repo lookup first
	matches, err := resource.FindByName(source, resource.TypeAgent)
	if err != nil {
		if errors.Is(err, resource.ErrNoReposConfigured) {
			return errors.New("no repositories configured. Run 'aix repo add <url>' to add one")
		}
		return fmt.Errorf("searching repositories: %w", err)
	}

	if len(matches) > 0 {
		if err := installer.InstallFromRepo(source, matches, scope); err != nil {
			return errors.Wrap(err, "installing from repo")
		}
		return nil
	}

	// No matches in repos - check if input might be a forgotten path
	if install.MightBePath(source, "agent") {
		if _, statErr := os.Stat(source); statErr == nil {
			return fmt.Errorf("agent %q not found in repositories, but a local file exists at that path.\nDid you mean: aix agent install --file %s", source, source)
		}
	}

	// Check if it's a local path that exists
	if _, err := os.Stat(source); err == nil {
		return installFromLocal(source, scope)
	}

	return fmt.Errorf("agent %q not found in any configured repository", source)
}

// installFromLocal installs an agent from a local file or directory.
func installFromLocal(source string, scope cli.Scope) error {
	platforms, err := cli.ResolvePlatforms(flags.GetPlatformFlag())
	if err != nil {
		return fmt.Errorf("resolving platforms: %w", err)
	}

	// Resolve AGENT.md path
	agentPath, err := resolveAgentPath(source)
	if err != nil {
		return err
	}

	// Calculate default name from filename/directory
	defaultName := strings.TrimSuffix(filepath.Base(agentPath), filepath.Ext(agentPath))
	if strings.ToUpper(defaultName) == "AGENT" {
		defaultName = filepath.Base(filepath.Dir(agentPath))
	}

	// Read and parse the AGENT.md file
	content, err := os.ReadFile(agentPath)
	if err != nil {
		return fmt.Errorf("reading agent file: %w", err)
	}

	// Track the agent name once successfully parsed (same for all platforms)
	var agentName string

	// Track results for each platform
	type installResult struct {
		platform   string
		installed  bool
		collision  bool
		targetPath string
		errMsg     string
	}
	results := make([]installResult, 0, len(platforms))

	// Install to each platform
	for _, p := range platforms {
		// Ensure backup exists before modifying
		if err := backup.EnsureBackedUp(p.Name(), p.BackupPaths()); err != nil {
			return fmt.Errorf("backing up %s before install: %w", p.DisplayName(), err)
		}

		result := installResult{platform: p.Name()}

		agent, parseErr := parseAgentForPlatform(p.Name(), content, defaultName)
		if parseErr != nil {
			result.errMsg = fmt.Sprintf("could not parse agent: %v", parseErr)
			results = append(results, result)
			continue
		}

		// Get agent name for collision check
		parsedName := getAgentName(agent)
		if parsedName == "" {
			result.errMsg = "agent name is required"
			results = append(results, result)
			continue
		}
		// Capture agent name on first successful parse (same for all platforms)
		if agentName == "" {
			agentName = parsedName
		}

		// Determine target path for error messages
		result.targetPath = filepath.Join(p.AgentDir(), parsedName+".md")

		// Check for collision unless --force is set
		existingAgent, getErr := p.GetAgent(parsedName, cli.ScopeDefault)
		if getErr == nil && existingAgent != nil {
			// Agent exists - check for idempotency
			if agentsAreIdentical(agent, existingAgent) {
				// Content is identical - succeed silently (idempotent)
				result.installed = true
				results = append(results, result)
				continue
			}

			if !installForce {
				// Collision detected and no --force
				result.collision = true
				results = append(results, result)
				continue
			}

			// --force flag set, will overwrite
			fmt.Fprintf(os.Stderr, "Warning: overwriting existing agent %q on %s\n", parsedName, p.DisplayName())
		}

		// Perform installation
		if installErr := p.InstallAgent(agent, scope); installErr != nil {
			result.errMsg = fmt.Sprintf("could not install agent: %v", installErr)
			results = append(results, result)
			continue
		}

		result.installed = true
		results = append(results, result)
	}

	// Collect results
	var installed, collisions []string
	var otherErrors []string

	for _, r := range results {
		switch {
		case r.installed:
			installed = append(installed, r.platform)
		case r.collision:
			collisions = append(collisions, fmt.Sprintf("%s: %s", r.platform, r.targetPath))
		case r.errMsg != "":
			otherErrors = append(otherErrors, fmt.Sprintf("%s: %s", r.platform, r.errMsg))
		}
	}

	// Report successful installations
	if len(installed) > 0 {
		fmt.Printf("Installed %s to %s\n", agentName, strings.Join(installed, ", "))
	}

	// Report other errors as warnings (they don't block collision errors)
	for _, e := range otherErrors {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", e)
	}

	// Handle collision errors
	if len(collisions) > 0 {
		fmt.Fprintf(os.Stderr, "\nAgent collision detected:\n")
		for _, c := range collisions {
			fmt.Fprintf(os.Stderr, "  - %s\n", c)
		}
		fmt.Fprintf(os.Stderr, "\nUse --force to overwrite existing agents.\n")

		// If some platforms succeeded but others had collisions, return partial error
		if len(installed) > 0 {
			return errAgentInstallPartial
		}
		return errAgentCollision
	}

	// If nothing was installed
	if len(installed) == 0 {
		return errAgentInstallFailed
	}

	return nil
}

// resolveAgentPath finds the AGENT.md file from the given source path.
func resolveAgentPath(source string) (string, error) {
	info, err := os.Stat(source)
	if err != nil {
		return "", fmt.Errorf("accessing source: %w", err)
	}

	if info.IsDir() {
		// Look for AGENT.md in directory
		agentPath := filepath.Join(source, "AGENT.md")
		if _, err := os.Stat(agentPath); err != nil {
			return "", fmt.Errorf("no AGENT.md found in %s", source)
		}
		return agentPath, nil
	}

	// Assume it's the AGENT.md file itself
	return source, nil
}

// parseAgentForPlatform parses AGENT.md content into platform-specific agent struct.
func parseAgentForPlatform(platform string, content []byte, defaultName string) (any, error) {
	switch platform {
	case "claude":
		var meta struct {
			Name        string `yaml:"name"`
			Description string `yaml:"description"`
		}
		body, err := frontmatter.Parse(bytes.NewReader(content), &meta)
		if err != nil {
			return nil, fmt.Errorf("parsing frontmatter: %w", err)
		}
		if meta.Name == "" {
			meta.Name = defaultName
		}
		if meta.Name == "" {
			return nil, errAgentNameRequired
		}
		return &claude.Agent{
				Name:         meta.Name,
				Description:  meta.Description,
				Instructions: string(body),
			},
			nil

	case "opencode":
		var meta struct {
			Name        string  `yaml:"name"`
			Description string  `yaml:"description"`
			Mode        string  `yaml:"mode"`
			Temperature float64 `yaml:"temperature"`
		}
		body, err := frontmatter.Parse(bytes.NewReader(content), &meta)
		if err != nil {
			return nil, fmt.Errorf("parsing frontmatter: %w", err)
		}
		if meta.Name == "" {
			meta.Name = defaultName
		}
		if meta.Name == "" {
			return nil, errAgentNameRequired
		}
		return &opencode.Agent{
				Name:         meta.Name,
				Description:  meta.Description,
				Mode:         meta.Mode,
				Temperature:  meta.Temperature,
				Instructions: string(body),
			},
			nil

	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}
}

// getAgentName extracts the name from a platform-specific agent struct.
func getAgentName(agent any) string {
	switch a := agent.(type) {
	case *claude.Agent:
		return a.Name
	case *opencode.Agent:
		return a.Name
	default:
		return ""
	}
}

// agentsAreIdentical compares two agents for content equality.
// Returns true if the agents have identical content (enabling idempotent installs).
func agentsAreIdentical(newAgent, existingAgent any) bool {
	switch n := newAgent.(type) {
	case *claude.Agent:
		existing, ok := existingAgent.(*claude.Agent)
		if !ok {
			return false
		}
		return n.Name == existing.Name &&
			n.Description == existing.Description &&
			normalizeInstructions(n.Instructions) == normalizeInstructions(existing.Instructions)

	case *opencode.Agent:
		existing, ok := existingAgent.(*opencode.Agent)
		if !ok {
			return false
		}
		return n.Name == existing.Name &&
			n.Description == existing.Description &&
			n.Mode == existing.Mode &&
			n.Temperature == existing.Temperature &&
			normalizeInstructions(n.Instructions) == normalizeInstructions(existing.Instructions)

	default:
		return false
	}
}

// normalizeInstructions normalizes whitespace in instructions for comparison.
// This handles minor formatting differences that shouldn't be considered collisions.
func normalizeInstructions(s string) string {
	return strings.TrimSpace(s)
}
