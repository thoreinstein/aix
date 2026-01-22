package commands

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/platform/claude"
	"github.com/thoreinstein/aix/internal/platform/opencode"
	"github.com/thoreinstein/aix/pkg/frontmatter"
)

// Sentinel errors for agent install operations.
var (
	errAgentInstallFailed = errors.New("failed to install agent to any platform")
	errAgentNameRequired  = errors.New("agent name is required")
)

func init() {
	agentCmd.AddCommand(agentInstallCmd)
}

var agentInstallCmd = &cobra.Command{
	Use:   "install <source>",
	Short: "Install an agent from a local path",
	Long: `Install an AI coding agent from a local AGENT.md file.

The source can be:
  - A path to an AGENT.md file
  - A directory containing an AGENT.md file

The AGENT.md file should contain YAML frontmatter with at least a 'name' field,
followed by the agent's instructions in markdown format.

Example AGENT.md:
  ---
  name: code-reviewer
  description: Reviews code for quality and best practices
  ---

  You are a code review expert. When reviewing code...`,
	Example: `  # Install from a file
  aix agent install ./my-agent/AGENT.md

  # Install from a directory
  aix agent install ./my-agent/

  # Install to specific platform
  aix agent install ./my-agent/ --platform claude`,
	Args: cobra.ExactArgs(1),
	RunE: runAgentInstall,
}

func runAgentInstall(_ *cobra.Command, args []string) error {
	source := args[0]
	platforms, err := cli.ResolvePlatforms(GetPlatformFlag())
	if err != nil {
		return fmt.Errorf("resolving platforms: %w", err)
	}

	// Resolve AGENT.md path
	agentPath, err := resolveAgentPath(source)
	if err != nil {
		return err
	}

	// Read and parse the AGENT.md file
	content, err := os.ReadFile(agentPath)
	if err != nil {
		return fmt.Errorf("reading agent file: %w", err)
	}

	// Install to each platform
	installed := make([]string, 0, len(platforms))
	for _, p := range platforms {
		agent, err := parseAgentForPlatform(p.Name(), content)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not parse agent for %s: %v\n", p.Name(), err)
			continue
		}

		if err := p.InstallAgent(agent); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not install agent to %s: %v\n", p.Name(), err)
			continue
		}
		installed = append(installed, p.Name())
	}

	if len(installed) == 0 {
		return errAgentInstallFailed
	}

	fmt.Printf("Agent installed to: %v\n", installed)
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
func parseAgentForPlatform(platform string, content []byte) (any, error) {
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
			return nil, errAgentNameRequired
		}
		return &claude.Agent{
			Name:         meta.Name,
			Description:  meta.Description,
			Instructions: string(body),
		}, nil

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
			return nil, errAgentNameRequired
		}
		return &opencode.Agent{
			Name:         meta.Name,
			Description:  meta.Description,
			Mode:         meta.Mode,
			Temperature:  meta.Temperature,
			Instructions: string(body),
		}, nil

	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}
}
