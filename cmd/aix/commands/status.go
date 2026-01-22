package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/doctor"
	"github.com/thoreinstein/aix/internal/paths"
)

var (
	statusJSON    bool
	statusQuiet   bool
	statusVerbose bool
)

func init() {
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "output as JSON")
	statusCmd.Flags().BoolVar(&statusQuiet, "quiet", false, "summary counts only")
	statusCmd.Flags().BoolVar(&statusVerbose, "verbose", false, "show detailed item information")
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show configuration overview",
	Long: `Show an overview of what's configured across all platforms.

Displays counts of skills, slash commands, and MCP servers for each
detected platform.

Output modes (mutually exclusive):
  (default)   Compact tabular view with counts per platform
  --quiet     Summary counts only, one line per platform
  --verbose   Show detailed item information (descriptions, env vars)
  --json      Machine-readable JSON output

In verbose mode, sensitive environment variable values (TOKEN, SECRET,
KEY, PASSWORD, CREDENTIAL, AUTH) are redacted for security.

Examples:
  # Show status for all platforms
  aix status

  # Quick summary
  aix status --quiet

  # Detailed view with item information
  aix status --verbose

  # JSON output for scripting
  aix status --json`,
	PreRunE: validateStatusFlags,
	RunE:    runStatus,
}

// validateStatusFlags ensures output flags are mutually exclusive.
func validateStatusFlags(_ *cobra.Command, _ []string) error {
	count := 0
	if statusJSON {
		count++
	}
	if statusQuiet {
		count++
	}
	if statusVerbose {
		count++
	}

	if count > 1 {
		return errors.New("flags --json, --quiet, and --verbose are mutually exclusive")
	}

	return nil
}

func runStatus(_ *cobra.Command, _ []string) error {
	return runStatusWithWriter(os.Stdout)
}

// runStatusWithWriter allows injecting a writer for testing.
func runStatusWithWriter(w io.Writer) error {
	// Get all known platforms, not just detected ones
	allPlatformNames := paths.Platforms()
	platforms := make([]cli.Platform, 0, len(allPlatformNames))

	for _, name := range allPlatformNames {
		p, err := cli.NewPlatform(name)
		if err != nil {
			continue // Skip platforms without adapters
		}
		platforms = append(platforms, p)
	}

	// Apply platform filter if specified
	filterPlatforms := GetPlatformFlag()
	if len(filterPlatforms) > 0 {
		filtered := make([]cli.Platform, 0, len(filterPlatforms))
		for _, p := range platforms {
			for _, name := range filterPlatforms {
				if p.Name() == name {
					filtered = append(filtered, p)
					break
				}
			}
		}
		platforms = filtered
	}

	if statusJSON {
		return outputStatusJSON(w, platforms)
	}
	if statusQuiet {
		return outputStatusQuiet(w, platforms)
	}
	if statusVerbose {
		return outputStatusVerbose(w, platforms)
	}
	return outputStatusCompact(w, platforms)
}

// platformStatus holds the collected status for a single platform.
type platformStatus struct {
	Platform    cli.Platform
	Available   bool
	Skills      []cli.SkillInfo
	SkillsErr   error
	Commands    []cli.CommandInfo
	CommandsErr error
	MCP         []cli.MCPInfo
	MCPErr      error
}

// collectPlatformStatus gathers all status information for a platform.
func collectPlatformStatus(p cli.Platform) platformStatus {
	status := platformStatus{
		Platform:  p,
		Available: p.IsAvailable(),
	}

	if !status.Available {
		return status
	}

	status.Skills, status.SkillsErr = p.ListSkills()
	status.Commands, status.CommandsErr = p.ListCommands()
	status.MCP, status.MCPErr = p.ListMCP()

	return status
}

// mcpCounts returns the total, enabled, and disabled MCP server counts.
func mcpCounts(servers []cli.MCPInfo) (total, enabled, disabled int) {
	total = len(servers)
	for _, s := range servers {
		if s.Disabled {
			disabled++
		} else {
			enabled++
		}
	}
	return total, enabled, disabled
}

// JSON output types

type statusJSONOutput struct {
	Version   string                       `json:"version"`
	Platforms map[string]platformJSONEntry `json:"platforms"`
}

type platformJSONEntry struct {
	Available bool            `json:"available"`
	Skills    *itemsJSONEntry `json:"skills,omitempty"`
	Commands  *itemsJSONEntry `json:"commands,omitempty"`
	MCP       *mcpJSONEntry   `json:"mcp,omitempty"`
	Error     string          `json:"error,omitempty"`
}

type itemsJSONEntry struct {
	Count int        `json:"count"`
	Items []itemJSON `json:"items,omitempty"`
	Error string     `json:"error,omitempty"`
}

type itemJSON struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type mcpJSONEntry struct {
	Count    int           `json:"count"`
	Enabled  int           `json:"enabled"`
	Disabled int           `json:"disabled"`
	Items    []mcpItemJSON `json:"items,omitempty"`
	Error    string        `json:"error,omitempty"`
}

type mcpItemJSON struct {
	Name      string            `json:"name"`
	Transport string            `json:"transport"`
	Command   string            `json:"command,omitempty"`
	URL       string            `json:"url,omitempty"`
	Disabled  bool              `json:"disabled"`
	Env       map[string]string `json:"env,omitempty"`
}

func outputStatusJSON(w io.Writer, platforms []cli.Platform) error {
	output := statusJSONOutput{
		Version:   Version,
		Platforms: make(map[string]platformJSONEntry),
	}

	for _, p := range platforms {
		status := collectPlatformStatus(p)
		entry := platformJSONEntry{
			Available: status.Available,
		}

		if !status.Available {
			output.Platforms[p.Name()] = entry
			continue
		}

		// Skills
		if status.SkillsErr != nil {
			entry.Skills = &itemsJSONEntry{Error: status.SkillsErr.Error()}
		} else {
			items := make([]itemJSON, len(status.Skills))
			for i, s := range status.Skills {
				items[i] = itemJSON{Name: s.Name, Description: s.Description}
			}
			entry.Skills = &itemsJSONEntry{Count: len(status.Skills), Items: items}
		}

		// Commands
		if status.CommandsErr != nil {
			entry.Commands = &itemsJSONEntry{Error: status.CommandsErr.Error()}
		} else {
			items := make([]itemJSON, len(status.Commands))
			for i, c := range status.Commands {
				items[i] = itemJSON{Name: c.Name, Description: c.Description}
			}
			entry.Commands = &itemsJSONEntry{Count: len(status.Commands), Items: items}
		}

		// MCP
		if status.MCPErr != nil {
			entry.MCP = &mcpJSONEntry{Error: status.MCPErr.Error()}
		} else {
			total, enabled, disabled := mcpCounts(status.MCP)
			items := make([]mcpItemJSON, len(status.MCP))
			for i, m := range status.MCP {
				items[i] = mcpItemJSON{
					Name:      m.Name,
					Transport: m.Transport,
					Command:   m.Command,
					URL:       m.URL,
					Disabled:  m.Disabled,
					Env:       doctor.MaskSecrets(m.Env),
				}
			}
			entry.MCP = &mcpJSONEntry{
				Count:    total,
				Enabled:  enabled,
				Disabled: disabled,
				Items:    items,
			}
		}

		output.Platforms[p.Name()] = entry
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

func outputStatusQuiet(w io.Writer, platforms []cli.Platform) error {
	for _, p := range platforms {
		status := collectPlatformStatus(p)

		if !status.Available {
			fmt.Fprintf(w, "%s: (not installed)\n", p.Name())
			continue
		}

		// Build parts, handling errors gracefully
		parts := make([]string, 0, 3)

		if status.SkillsErr != nil {
			parts = append(parts, "skills: error")
		} else {
			parts = append(parts, fmt.Sprintf("%d skills", len(status.Skills)))
		}

		if status.CommandsErr != nil {
			parts = append(parts, "commands: error")
		} else {
			parts = append(parts, fmt.Sprintf("%d commands", len(status.Commands)))
		}

		if status.MCPErr != nil {
			parts = append(parts, "mcp: error")
		} else {
			parts = append(parts, fmt.Sprintf("%d mcp", len(status.MCP)))
		}

		fmt.Fprintf(w, "%s: %s\n", p.Name(), strings.Join(parts, ", "))
	}

	return nil
}

func outputStatusCompact(w io.Writer, platforms []cli.Platform) error {
	fmt.Fprintf(w, "aix version %s\n", Version)

	for _, p := range platforms {
		status := collectPlatformStatus(p)

		// Add blank line before each platform section
		fmt.Fprintln(w)

		fmt.Fprintf(w, "%sPlatform: %s%s", colorCyan+colorBold, p.DisplayName(), colorReset)
		if !status.Available {
			fmt.Fprintf(w, " %s(not installed)%s\n", colorGray, colorReset)
			continue
		}
		fmt.Fprintln(w)

		// Skills
		if status.SkillsErr != nil {
			fmt.Fprintf(w, "  %sSkills: error - %s%s\n", colorYellow, status.SkillsErr, colorReset)
		} else {
			fmt.Fprintf(w, "  Skills: %d\n", len(status.Skills))
		}

		// Commands
		if status.CommandsErr != nil {
			fmt.Fprintf(w, "  %sCommands: error - %s%s\n", colorYellow, status.CommandsErr, colorReset)
		} else {
			fmt.Fprintf(w, "  Commands: %d\n", len(status.Commands))
		}

		// MCP
		if status.MCPErr != nil {
			fmt.Fprintf(w, "  %sMCP Servers: error - %s%s\n", colorYellow, status.MCPErr, colorReset)
		} else {
			total, _, disabled := mcpCounts(status.MCP)
			if disabled > 0 {
				fmt.Fprintf(w, "  MCP Servers: %d (%d disabled)\n", total, disabled)
			} else {
				fmt.Fprintf(w, "  MCP Servers: %d\n", total)
			}
		}
	}

	return nil
}

func outputStatusVerbose(w io.Writer, platforms []cli.Platform) error {
	fmt.Fprintf(w, "aix version %s\n", Version)

	for _, p := range platforms {
		status := collectPlatformStatus(p)

		// Add blank line before each platform section
		fmt.Fprintln(w)

		fmt.Fprintf(w, "%sPlatform: %s%s", colorCyan+colorBold, p.DisplayName(), colorReset)
		if !status.Available {
			fmt.Fprintf(w, " %s(not installed)%s\n", colorGray, colorReset)
			continue
		}
		fmt.Fprintln(w)

		// Skills
		fmt.Fprintf(w, "\n  %sSkills:%s", colorBold, colorReset)
		if status.SkillsErr != nil {
			fmt.Fprintf(w, " %serror - %s%s\n", colorYellow, status.SkillsErr, colorReset)
		} else if len(status.Skills) == 0 {
			fmt.Fprintf(w, " %s(none)%s\n", colorGray, colorReset)
		} else {
			fmt.Fprintf(w, " %d\n", len(status.Skills))
			for _, s := range status.Skills {
				fmt.Fprintf(w, "    %s%s%s", colorGreen, s.Name, colorReset)
				if s.Description != "" {
					fmt.Fprintf(w, " - %s", truncate(s.Description, 60))
				}
				fmt.Fprintln(w)
			}
		}

		// Commands
		fmt.Fprintf(w, "\n  %sCommands:%s", colorBold, colorReset)
		if status.CommandsErr != nil {
			fmt.Fprintf(w, " %serror - %s%s\n", colorYellow, status.CommandsErr, colorReset)
		} else if len(status.Commands) == 0 {
			fmt.Fprintf(w, " %s(none)%s\n", colorGray, colorReset)
		} else {
			fmt.Fprintf(w, " %d\n", len(status.Commands))
			for _, c := range status.Commands {
				fmt.Fprintf(w, "    %s/%s%s", colorGreen, c.Name, colorReset)
				if c.Description != "" {
					fmt.Fprintf(w, " - %s", truncate(c.Description, 60))
				}
				fmt.Fprintln(w)
			}
		}

		// MCP
		fmt.Fprintf(w, "\n  %sMCP Servers:%s", colorBold, colorReset)
		if status.MCPErr != nil {
			fmt.Fprintf(w, " %serror - %s%s\n", colorYellow, status.MCPErr, colorReset)
		} else if len(status.MCP) == 0 {
			fmt.Fprintf(w, " %s(none)%s\n", colorGray, colorReset)
		} else {
			total, enabled, disabled := mcpCounts(status.MCP)
			if disabled > 0 {
				fmt.Fprintf(w, " %d (%d enabled, %d disabled)\n", total, enabled, disabled)
			} else {
				fmt.Fprintf(w, " %d\n", total)
			}
			for _, m := range status.MCP {
				statusStr := "enabled"
				statusColor := colorGreen
				if m.Disabled {
					statusStr = "disabled"
					statusColor = colorGray
				}
				fmt.Fprintf(w, "    %s%s%s [%s%s%s]\n", colorGreen, m.Name, colorReset, statusColor, statusStr, colorReset)
				fmt.Fprintf(w, "      Transport: %s\n", m.Transport)
				if m.Command != "" {
					fmt.Fprintf(w, "      Command: %s\n", m.Command)
				}
				if m.URL != "" {
					fmt.Fprintf(w, "      URL: %s\n", m.URL)
				}
				if len(m.Env) > 0 {
					fmt.Fprintf(w, "      Env:\n")
					maskedEnv := doctor.MaskSecrets(m.Env)
					for k, v := range maskedEnv {
						fmt.Fprintf(w, "        %s=%s\n", k, v)
					}
				}
			}
		}
	}

	return nil
}

// ANSI color codes (additional to those in skill_list.go)
const colorYellow = "\033[33m"
