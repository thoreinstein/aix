package skill

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/cmd/aix/commands/flags"
	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/errors"
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
	Short: "Display detailed skill information",
	Long: `Display detailed information about an installed skill.

Searches for the skill across all detected platforms (or only the specified
--platform). Shows metadata, allowed tools, installation locations, and an
instructions preview.`,
	Example: `  # Show details for 'debug' skill
  aix skill show debug

  # Show full instructions
  aix skill show debug --full

  # Show details as JSON
  aix skill show debug --json

  # Show details for specific platform
  aix skill show debug --platform claude

  See Also:
    aix skill list     - List installed skills
    aix skill edit     - Edit a skill`,
	Args: cobra.ExactArgs(1),
	RunE: runShow,
}

// showDetail holds unified skill information for display.
type showDetail struct {
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	License       string            `json:"license,omitempty"`
	Version       string            `json:"version,omitempty"`
	Author        string            `json:"author,omitempty"`
	Compatibility []string          `json:"compatibility,omitempty"`
	AllowedTools  []string          `json:"allowed_tools,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	Instructions  string            `json:"instructions,omitempty"`
	Installations []installLocation `json:"installations"`
}

// installLocation describes where a skill is installed.
type installLocation struct {
	Platform string `json:"platform"`
	Path     string `json:"path"`
}

func runShow(_ *cobra.Command, args []string) error {
	name := args[0]

	platforms, err := cli.ResolvePlatforms(flags.GetPlatformFlag())
	if err != nil {
		return errors.Wrap(err, "resolving platforms")
	}

	// Collect skill info from all platforms where it exists
	var detail *showDetail
	installations := make([]installLocation, 0, len(platforms))

	for _, p := range platforms {
		skillAny, err := p.GetSkill(name)
		if err != nil {
			// Skill not found on this platform, continue to next
			continue
		}

		// Build installation location
		skillPath := filepath.Join(p.SkillDir(), name, "SKILL.md")
		installations = append(installations, installLocation{
			Platform: p.DisplayName(),
			Path:     skillPath,
		})

		// Extract skill details (use first found as canonical)
		if detail == nil {
			detail = extractDetail(skillAny)
		}
	}

	if detail == nil {
		return errors.Newf("skill %q not found on any platform", name)
	}

	detail.Installations = installations

	// Truncate instructions unless --full is specified
	if !showFull && len(detail.Instructions) > defaultInstructionsPreviewLength {
		detail.Instructions = detail.Instructions[:defaultInstructionsPreviewLength]
	}

	if showJSON {
		return outputShowAsJSON(detail)
	}

	return outputShowAsText(detail)
}

// extractDetail converts a platform-specific skill to the unified detail struct.
func extractDetail(skill any) *showDetail {
	switch s := skill.(type) {
	case *claude.Skill:
		return extractClaudeDetail(s)
	case *opencode.Skill:
		return extractOpenCodeDetail(s)
	default:
		return nil
	}
}

// extractClaudeDetail extracts details from a Claude skill.
func extractClaudeDetail(s *claude.Skill) *showDetail {
	allowedTools := []string(s.AllowedTools)

	return &showDetail{
		Name:          s.Name,
		Description:   s.Description,
		License:       s.License,
		Compatibility: s.Compatibility,
		AllowedTools:  allowedTools,
		Metadata:      s.Metadata,
		Instructions:  s.Instructions,
	}
}

// extractOpenCodeDetail extracts details from an OpenCode skill.
func extractOpenCodeDetail(s *opencode.Skill) *showDetail {
	// Convert OpenCode's map[string]any metadata to map[string]string
	var metadata map[string]string
	if len(s.Metadata) > 0 {
		metadata = make(map[string]string, len(s.Metadata))
		for k, v := range s.Metadata {
			metadata[k] = fmt.Sprint(v)
		}
	}

	// Convert compatibility map to string slice for display
	var compatibility []string
	for platform, version := range s.Compatibility {
		if version != "" {
			compatibility = append(compatibility, fmt.Sprintf("%s %s", platform, version))
		} else {
			compatibility = append(compatibility, platform)
		}
	}

	return &showDetail{
		Name:          s.Name,
		Description:   s.Description,
		Version:       s.Version,
		Author:        s.Author,
		Compatibility: compatibility,
		AllowedTools:  s.AllowedTools,
		Metadata:      metadata,
		Instructions:  s.Instructions,
	}
}

func outputShowAsJSON(detail *showDetail) error {
	data, err := json.MarshalIndent(detail, "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshaling JSON")
	}
	fmt.Println(string(data))
	return nil
}

func outputShowAsText(detail *showDetail) error {
	fmt.Printf("Skill: %s\n", detail.Name)
	fmt.Printf("Description: %s\n", detail.Description)

	if detail.License != "" {
		fmt.Printf("License: %s\n", detail.License)
	}
	if detail.Version != "" {
		fmt.Printf("Version: %s\n", detail.Version)
	}
	if detail.Author != "" {
		fmt.Printf("Author: %s\n", detail.Author)
	}
	if len(detail.Compatibility) > 0 {
		fmt.Printf("Compatibility: %s\n", strings.Join(detail.Compatibility, ", "))
	}

	if len(detail.AllowedTools) > 0 {
		fmt.Println("\nAllowed Tools:")
		for _, tool := range detail.AllowedTools {
			fmt.Printf("  - %s\n", tool)
		}
	}

	if len(detail.Installations) > 0 {
		fmt.Println("\nInstalled On:")
		for _, loc := range detail.Installations {
			fmt.Printf("  - %s (%s)\n", loc.Platform, loc.Path)
		}
	}

	if detail.Instructions != "" {
		fmt.Println("\nInstructions Preview:")
		fmt.Printf("  %s\n", detail.Instructions)
		if !showFull {
			fmt.Println("  [truncated, use --full for complete output]")
		}
	}

	return nil
}
