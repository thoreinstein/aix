package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/platform/claude"
	"github.com/thoreinstein/aix/internal/platform/opencode"
)

const defaultAgentInstructionsPreviewLength = 200

var (
	agentShowJSON bool
	agentShowFull bool
)

func init() {
	agentShowCmd.Flags().BoolVar(&agentShowJSON, "json", false, "Output as JSON")
	agentShowCmd.Flags().BoolVar(&agentShowFull, "full", false, "Show complete instructions (default truncated)")
	agentCmd.AddCommand(agentShowCmd)
}

var agentShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Display detailed agent information",
	Long: `Display detailed information about an installed agent.

Searches for the agent across all detected platforms (or only the specified
--platform). Shows metadata, installation locations, and an instructions preview.

Examples:
  aix agent show code-reviewer
  aix agent show code-reviewer --full
  aix agent show code-reviewer --json
  aix agent show code-reviewer --platform claude`,
	Args: cobra.ExactArgs(1),
	RunE: runAgentShow,
}

// showAgentDetail holds unified agent information for display.
type showAgentDetail struct {
	Name          string                 `json:"name"`
	Description   string                 `json:"description,omitempty"`
	Instructions  string                 `json:"instructions,omitempty"`
	Installations []agentInstallLocation `json:"installations"`

	// OpenCode-specific fields
	Mode        string  `json:"mode,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

// agentInstallLocation describes where an agent is installed.
type agentInstallLocation struct {
	Platform string `json:"platform"`
	Path     string `json:"path"`
}

func runAgentShow(_ *cobra.Command, args []string) error {
	return runAgentShowWithWriter(args[0], os.Stdout)
}

// runAgentShowWithWriter allows injecting a writer for testing.
func runAgentShowWithWriter(name string, w io.Writer) error {
	platforms, err := cli.ResolvePlatforms(GetPlatformFlag())
	if err != nil {
		return err
	}

	// Collect agent info from all platforms where it exists
	var detail *showAgentDetail
	installations := make([]agentInstallLocation, 0, len(platforms))

	for _, p := range platforms {
		agentAny, err := p.GetAgent(name)
		if err != nil {
			// Agent not found on this platform, continue to next
			continue
		}

		// Build installation location
		agentPath := filepath.Join(p.AgentDir(), name+".md")
		installations = append(installations, agentInstallLocation{
			Platform: p.DisplayName(),
			Path:     agentPath,
		})

		// Extract agent details (use first found as canonical)
		if detail == nil {
			detail = extractAgentDetail(agentAny)
		}
	}

	if detail == nil {
		return fmt.Errorf("agent %q not found on any platform", name)
	}

	detail.Installations = installations

	// Truncate instructions unless --full is specified
	if !agentShowFull && len(detail.Instructions) > defaultAgentInstructionsPreviewLength {
		detail.Instructions = detail.Instructions[:defaultAgentInstructionsPreviewLength]
	}

	if agentShowJSON {
		return outputAgentShowJSON(w, detail)
	}

	return outputAgentShowText(w, detail)
}

// extractAgentDetail converts a platform-specific agent to the unified detail struct.
func extractAgentDetail(agent any) *showAgentDetail {
	switch a := agent.(type) {
	case *claude.Agent:
		return extractClaudeAgent(a)
	case *opencode.Agent:
		return extractOpenCodeAgent(a)
	default:
		return nil
	}
}

// extractClaudeAgent extracts details from a Claude agent.
func extractClaudeAgent(a *claude.Agent) *showAgentDetail {
	return &showAgentDetail{
		Name:         a.Name,
		Description:  a.Description,
		Instructions: a.Instructions,
	}
}

// extractOpenCodeAgent extracts details from an OpenCode agent.
func extractOpenCodeAgent(a *opencode.Agent) *showAgentDetail {
	return &showAgentDetail{
		Name:         a.Name,
		Description:  a.Description,
		Mode:         a.Mode,
		Temperature:  a.Temperature,
		Instructions: a.Instructions,
	}
}

func outputAgentShowJSON(w io.Writer, detail *showAgentDetail) error {
	data, err := json.MarshalIndent(detail, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	fmt.Fprintln(w, string(data))
	return nil
}

func outputAgentShowText(w io.Writer, detail *showAgentDetail) error {
	fmt.Fprintf(w, "Agent: %s\n", detail.Name)

	if detail.Description != "" {
		fmt.Fprintf(w, "Description: %s\n", detail.Description)
	}

	// OpenCode-specific fields
	if detail.Mode != "" {
		fmt.Fprintf(w, "Mode: %s\n", detail.Mode)
	}
	if detail.Temperature != 0 {
		fmt.Fprintf(w, "Temperature: %.2f\n", detail.Temperature)
	}

	if len(detail.Installations) > 0 {
		fmt.Fprintln(w, "\nInstalled On:")
		for _, loc := range detail.Installations {
			fmt.Fprintf(w, "  - %s (%s)\n", loc.Platform, loc.Path)
		}
	}

	if detail.Instructions != "" {
		fmt.Fprintln(w, "\nInstructions Preview:")
		fmt.Fprintf(w, "  %s\n", detail.Instructions)
		if !agentShowFull {
			fmt.Fprintln(w, "  [truncated, use --full for complete output]")
		}
	}

	return nil
}
