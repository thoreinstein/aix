package command

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
	Short: "Display detailed command information",
	Long: `Display detailed information about an installed slash command.

Searches for the command across all detected platforms (or only the specified
--platform). Shows metadata, installation locations, and an instructions preview.`,
	Example: `  # Show command details
  aix command show review

  # Show full instructions (no truncation)
  aix command show review --full

  # Output as JSON
  aix command show review --json

  # Show details for a specific platform
  aix command show review --platform claude

  See Also:
    aix command list     - List installed commands
    aix command edit     - Edit a command definition
    aix command validate - Validate a command file`,
	Args: cobra.ExactArgs(1),
	RunE: runShow,
}

// showDetail holds unified command information for display.
type showDetail struct {
	Name          string            `json:"name"`
	Description   string            `json:"description,omitempty"`
	Model         string            `json:"model,omitempty"`
	Agent         string            `json:"agent,omitempty"`
	Instructions  string            `json:"instructions,omitempty"`
	Installations []installLocation `json:"installations"`

	// Claude-specific fields
	ArgumentHint           string   `json:"argument_hint,omitempty"`
	DisableModelInvocation bool     `json:"disable_model_invocation,omitempty"`
	UserInvocable          bool     `json:"user_invocable,omitempty"`
	AllowedTools           []string `json:"allowed_tools,omitempty"`
	Context                string   `json:"context,omitempty"`
	Hooks                  []string `json:"hooks,omitempty"`

	// OpenCode-specific fields
	Subtask  bool   `json:"subtask,omitempty"`
	Template string `json:"template,omitempty"`
}

// installLocation describes where a command is installed.
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

	// Collect command info from all platforms where it exists
	var detail *showDetail
	installations := make([]installLocation, 0, len(platforms))

	for _, p := range platforms {
		cmdAny, err := p.GetCommand(name)
		if err != nil {
			// Command not found on this platform, continue to next
			continue
		}

		// Build installation location
		cmdPath := filepath.Join(p.CommandDir(), name+".md")
		installations = append(installations, installLocation{
			Platform: p.DisplayName(),
			Path:     cmdPath,
		})

		// Extract command details (use first found as canonical)
		if detail == nil {
			detail = extractDetail(cmdAny)
		}
	}

	if detail == nil {
		return errors.Newf("command %q not found on any platform", name)
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

// extractDetail converts a platform-specific command to the unified detail struct.
func extractDetail(cmd any) *showDetail {
	switch c := cmd.(type) {
	case *claude.Command:
		return extractClaudeDetail(c)
	case *opencode.Command:
		return extractOpenCodeDetail(c)
	default:
		return nil
	}
}

// extractClaudeDetail extracts details from a Claude command.
func extractClaudeDetail(c *claude.Command) *showDetail {
	var allowedTools []string
	if len(c.AllowedTools) > 0 {
		allowedTools = []string(c.AllowedTools)
	}

	return &showDetail{
		Name:                   c.Name,
		Description:            c.Description,
		Model:                  c.Model,
		Agent:                  c.Agent,
		ArgumentHint:           c.ArgumentHint,
		DisableModelInvocation: c.DisableModelInvocation,
		UserInvocable:          c.UserInvocable,
		AllowedTools:           allowedTools,
		Context:                c.Context,
		Hooks:                  c.Hooks,
		Instructions:           c.Instructions,
	}
}

// extractOpenCodeDetail extracts details from an OpenCode command.
func extractOpenCodeDetail(c *opencode.Command) *showDetail {
	return &showDetail{
		Name:         c.Name,
		Description:  c.Description,
		Model:        c.Model,
		Agent:        c.Agent,
		Subtask:      c.Subtask,
		Template:     c.Template,
		Instructions: c.Instructions,
	}
}

func outputShowJSON(w io.Writer, detail *showDetail) error {
	data, err := json.MarshalIndent(detail, "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshaling JSON")
	}
	fmt.Fprintln(w, string(data))
	return nil
}

func outputShowText(w io.Writer, detail *showDetail) error {
	fmt.Fprintf(w, "Command: /%s\n", detail.Name)
	fmt.Fprintf(w, "Description: %s\n", detail.Description)

	if detail.Model != "" {
		fmt.Fprintf(w, "Model: %s\n", detail.Model)
	}
	if detail.Agent != "" {
		fmt.Fprintf(w, "Agent: %s\n", detail.Agent)
	}

	// Claude-specific fields
	if detail.ArgumentHint != "" {
		fmt.Fprintf(w, "Argument Hint: %s\n", detail.ArgumentHint)
	}
	if detail.Context != "" {
		fmt.Fprintf(w, "Context: %s\n", detail.Context)
	}
	if detail.DisableModelInvocation {
		fmt.Fprintf(w, "Model Invocation: disabled\n")
	}
	if detail.UserInvocable {
		fmt.Fprintf(w, "User Invocable: true\n")
	}
	if len(detail.Hooks) > 0 {
		fmt.Fprintf(w, "Hooks: %s\n", strings.Join(detail.Hooks, ", "))
	}

	// OpenCode-specific fields
	if detail.Subtask {
		fmt.Fprintf(w, "Subtask: true\n")
	}
	if detail.Template != "" {
		fmt.Fprintf(w, "Template: %s\n", detail.Template)
	}

	if len(detail.AllowedTools) > 0 {
		fmt.Println("\nAllowed Tools:")
		for _, tool := range detail.AllowedTools {
			fmt.Fprintf(w, "  - %s\n", tool)
		}
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
