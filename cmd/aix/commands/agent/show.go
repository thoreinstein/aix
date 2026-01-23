package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/cmd/aix/commands/flags"
	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/platform/claude"
	"github.com/thoreinstein/aix/internal/platform/opencode"
)

const defaultInstructionsPreviewLength = 200

var (
	showJSON bool
	showFull bool
)

func init() {
	showCmd.Flags().BoolVar(&showJSON, "json", false, "Output as JSON")
	showCmd.Flags().BoolVar(&showFull, "full", false, "Show complete instructions (default truncated)")
	Cmd.AddCommand(showCmd)
}

var showCmd = &cobra.Command{
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
	RunE: runShow,
}

// showDetail holds unified agent information for display.
type showDetail struct {
	Name          string            `json:"name"`
	Description   string            `json:"description,omitempty"`
	Instructions  string            `json:"instructions,omitempty"`
	Installations []installLocation `json:"installations"`

	// OpenCode-specific fields
	Mode        string  `json:"mode,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

// installLocation describes where an agent is installed.
type installLocation struct {
	Platform string `json:"platform"`
	Path     string `json:"path"`
}

func runShow(_ *cobra.Command, args []string) error {
	return runShowWithWriter(args[0], os.Stdout)
}

// runShowWithWriter allows injecting a writer for testing.
func runShowWithWriter(name string, w io.Writer) error {
	platforms, err := cli.ResolvePlatforms(flags.GetPlatformFlag())
	if err != nil {
		return err
	}

	// Collect agent info from all platforms where it exists
	var detail *showDetail
	installations := make([]installLocation, 0, len(platforms))

	for _, p := range platforms {
		agentAny, err := p.GetAgent(name)
		if err != nil {
			// Agent not found on this platform is expected - try next platform
			if errors.Is(err, claude.ErrAgentNotFound) || errors.Is(err, opencode.ErrAgentNotFound) {
				continue
			}
			// Other errors (permission, parse) should be reported
			return fmt.Errorf("reading agent from %s: %w", p.DisplayName(), err)
		}

		// Build installation location
		agentPath := filepath.Join(p.AgentDir(), name+".md")
		installations = append(installations, installLocation{
			Platform: p.DisplayName(),
			Path:     agentPath,
		})

		// Extract agent details (use first found as canonical)
		if detail == nil {
			detail = extractDetail(agentAny)
		}
	}

	if detail == nil {
		return fmt.Errorf("agent %q not found on any platform", name)
	}

	detail.Installations = installations

	// Truncate instructions unless --full is specified
	if !showFull && len(detail.Instructions) > defaultInstructionsPreviewLength {
		detail.Instructions = detail.Instructions[:defaultInstructionsPreviewLength]
	}

	if showJSON {
		return outputShowJSON(w, detail)
	}

	return outputShowText(w, detail)
}

// extractDetail converts a platform-specific agent to the unified detail struct.
func extractDetail(agent any) *showDetail {
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
func extractClaudeAgent(a *claude.Agent) *showDetail {
	return &showDetail{
		Name:         a.Name,
		Description:  a.Description,
		Instructions: a.Instructions,
	}
}

// extractOpenCodeAgent extracts details from an OpenCode agent.
func extractOpenCodeAgent(a *opencode.Agent) *showDetail {
	return &showDetail{
		Name:         a.Name,
		Description:  a.Description,
		Mode:         a.Mode,
		Temperature:  a.Temperature,
		Instructions: a.Instructions,
	}
}

func outputShowJSON(w io.Writer, detail *showDetail) error {
	data, err := json.MarshalIndent(detail, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	fmt.Fprintln(w, string(data))
	return nil
}

func outputShowText(w io.Writer, detail *showDetail) error {
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
		if !showFull {
			fmt.Fprintln(w, "  [truncated, use --full for complete output]")
		}
	}

	return nil
}
