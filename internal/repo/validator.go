// Package repo provides repository management for skill repositories.
package repo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/thoreinstein/aix/internal/mcp"
	"github.com/thoreinstein/aix/pkg/frontmatter"
)

// ValidationWarning represents a non-fatal issue found during repository validation.
// Warnings are informational and do not block operations.
type ValidationWarning struct {
	// Path is the relative path within the repository where the issue was found.
	Path string

	// Message describes the validation issue.
	Message string
}

// expectedDirs lists the standard resource directories in an aix repository.
var expectedDirs = []string{"skills", "commands", "agents", "mcp"}

// ValidateRepoContent checks a repository for structural issues and invalid resources.
// It returns warnings for missing directories and unparseable resource files.
// This function never returns an error - all issues are reported as warnings
// to avoid blocking operations.
func ValidateRepoContent(repoPath string) []ValidationWarning {
	var warnings []ValidationWarning

	// Check for expected directories
	for _, dir := range expectedDirs {
		dirPath := filepath.Join(repoPath, dir)
		info, err := os.Stat(dirPath)
		if os.IsNotExist(err) {
			warnings = append(warnings, ValidationWarning{
				Path:    dir,
				Message: "directory not found",
			})
			continue
		}
		if err != nil {
			warnings = append(warnings, ValidationWarning{
				Path:    dir,
				Message: "cannot access directory: " + err.Error(),
			})
			continue
		}
		if !info.IsDir() {
			warnings = append(warnings, ValidationWarning{
				Path:    dir,
				Message: "expected directory, found file",
			})
			continue
		}

		// Validate resources in each directory
		dirWarnings := validateDirectory(repoPath, dir)
		warnings = append(warnings, dirWarnings...)
	}

	return warnings
}

// validateDirectory validates resources within a specific directory type.
func validateDirectory(repoPath, dirType string) []ValidationWarning {
	switch dirType {
	case "skills":
		return validateSkills(repoPath)
	case "commands":
		return validateCommands(repoPath)
	case "agents":
		return validateAgents(repoPath)
	case "mcp":
		return validateMCP(repoPath)
	default:
		return nil
	}
}

// resourceMeta is a minimal struct for parsing frontmatter headers.
type resourceMeta struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// validateSkills checks all skill directories for valid SKILL.md files.
func validateSkills(repoPath string) []ValidationWarning {
	var warnings []ValidationWarning
	skillsDir := filepath.Join(repoPath, "skills")

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		// Directory access issues are already reported by ValidateRepoContent
		return nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillPath := filepath.Join(skillsDir, entry.Name(), "SKILL.md")
		relPath := filepath.Join("skills", entry.Name(), "SKILL.md")

		if _, err := os.Stat(skillPath); os.IsNotExist(err) {
			warnings = append(warnings, ValidationWarning{
				Path:    filepath.Join("skills", entry.Name()),
				Message: "skill directory missing SKILL.md",
			})
			continue
		}

		if err := validateFrontmatter(skillPath); err != nil {
			warnings = append(warnings, ValidationWarning{
				Path:    relPath,
				Message: "invalid frontmatter: " + err.Error(),
			})
		}
	}

	return warnings
}

// validateCommands checks all command resources for valid frontmatter.
func validateCommands(repoPath string) []ValidationWarning {
	var warnings []ValidationWarning
	cmdsDir := filepath.Join(repoPath, "commands")

	entries, err := os.ReadDir(cmdsDir)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Directory-style command: commands/name/command.md
			cmdPath := filepath.Join(cmdsDir, entry.Name(), "command.md")
			relPath := filepath.Join("commands", entry.Name(), "command.md")

			if _, err := os.Stat(cmdPath); os.IsNotExist(err) {
				warnings = append(warnings, ValidationWarning{
					Path:    filepath.Join("commands", entry.Name()),
					Message: "command directory missing command.md",
				})
				continue
			}

			if err := validateFrontmatter(cmdPath); err != nil {
				warnings = append(warnings, ValidationWarning{
					Path:    relPath,
					Message: "invalid frontmatter: " + err.Error(),
				})
			}
		} else if strings.HasSuffix(entry.Name(), ".md") {
			// Direct file command: commands/name.md
			cmdPath := filepath.Join(cmdsDir, entry.Name())
			relPath := filepath.Join("commands", entry.Name())

			if err := validateFrontmatter(cmdPath); err != nil {
				warnings = append(warnings, ValidationWarning{
					Path:    relPath,
					Message: "invalid frontmatter: " + err.Error(),
				})
			}
		}
	}

	return warnings
}

// validateAgents checks all agent resources for valid frontmatter.
func validateAgents(repoPath string) []ValidationWarning {
	var warnings []ValidationWarning
	agentsDir := filepath.Join(repoPath, "agents")

	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Directory-style agent: agents/name/AGENT.md
			agentPath := filepath.Join(agentsDir, entry.Name(), "AGENT.md")
			relPath := filepath.Join("agents", entry.Name(), "AGENT.md")

			if _, err := os.Stat(agentPath); os.IsNotExist(err) {
				warnings = append(warnings, ValidationWarning{
					Path:    filepath.Join("agents", entry.Name()),
					Message: "agent directory missing AGENT.md",
				})
				continue
			}

			if err := validateFrontmatter(agentPath); err != nil {
				warnings = append(warnings, ValidationWarning{
					Path:    relPath,
					Message: "invalid frontmatter: " + err.Error(),
				})
			}
		} else if strings.HasSuffix(entry.Name(), ".md") {
			// Direct file agent: agents/name.md
			agentPath := filepath.Join(agentsDir, entry.Name())
			relPath := filepath.Join("agents", entry.Name())

			if err := validateFrontmatter(agentPath); err != nil {
				warnings = append(warnings, ValidationWarning{
					Path:    relPath,
					Message: "invalid frontmatter: " + err.Error(),
				})
			}
		}
	}

	return warnings
}

// validateMCP checks all MCP server files for valid JSON.
func validateMCP(repoPath string) []ValidationWarning {
	var warnings []ValidationWarning
	mcpDir := filepath.Join(repoPath, "mcp")

	entries, err := os.ReadDir(mcpDir)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		mcpPath := filepath.Join(mcpDir, entry.Name())
		relPath := filepath.Join("mcp", entry.Name())

		data, err := os.ReadFile(mcpPath)
		if err != nil {
			warnings = append(warnings, ValidationWarning{
				Path:    relPath,
				Message: "cannot read file: " + err.Error(),
			})
			continue
		}

		var server mcp.Server
		if err := json.Unmarshal(data, &server); err != nil {
			warnings = append(warnings, ValidationWarning{
				Path:    relPath,
				Message: "invalid JSON: " + err.Error(),
			})
		}
	}

	return warnings
}

// validateFrontmatter attempts to parse the frontmatter from a markdown file.
// It returns an error if the frontmatter is malformed.
func validateFrontmatter(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var meta resourceMeta
	return frontmatter.ParseHeader(file, &meta)
}
