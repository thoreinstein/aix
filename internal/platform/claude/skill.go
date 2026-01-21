package claude

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/thoreinstein/aix/pkg/frontmatter"
)

// Sentinel errors for skill operations.
var (
	ErrSkillNotFound = errors.New("skill not found")
	ErrInvalidSkill  = errors.New("invalid skill: name required")
)

// SkillManager handles CRUD operations for Claude Code skills.
type SkillManager struct {
	paths *ClaudePaths
}

// NewSkillManager creates a new SkillManager with the given paths configuration.
func NewSkillManager(paths *ClaudePaths) *SkillManager {
	return &SkillManager{
		paths: paths,
	}
}

// List returns all skills in the skill directory.
// Returns an empty slice if the directory doesn't exist or is empty.
func (m *SkillManager) List() ([]*Skill, error) {
	skillDir := m.paths.SkillDir()
	if skillDir == "" {
		return nil, nil
	}

	entries, err := os.ReadDir(skillDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading skill directory: %w", err)
	}

	// Count directories for pre-allocation
	dirCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			dirCount++
		}
	}

	skills := make([]*Skill, 0, dirCount)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skill, err := m.Get(entry.Name())
		if err != nil {
			if errors.Is(err, ErrSkillNotFound) {
				// Skip directories without valid SKILL.md
				continue
			}
			return nil, fmt.Errorf("reading skill %s: %w", entry.Name(), err)
		}
		skills = append(skills, skill)
	}

	return skills, nil
}

// Get retrieves a skill by name.
// Returns ErrSkillNotFound if the skill doesn't exist.
func (m *SkillManager) Get(name string) (*Skill, error) {
	if name == "" {
		return nil, ErrInvalidSkill
	}

	skillPath := m.paths.SkillPath(name)
	if skillPath == "" {
		return nil, ErrSkillNotFound
	}

	data, err := os.ReadFile(skillPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrSkillNotFound
		}
		return nil, fmt.Errorf("reading skill file: %w", err)
	}

	skill, err := parseSkillFile(data)
	if err != nil {
		return nil, fmt.Errorf("parsing skill file: %w", err)
	}

	// Ensure name matches directory name
	skill.Name = name

	return skill, nil
}

// Install creates or overwrites a skill.
// Creates the skill directory if it doesn't exist.
func (m *SkillManager) Install(s *Skill) error {
	if s == nil || s.Name == "" {
		return ErrInvalidSkill
	}

	skillPath := m.paths.SkillPath(s.Name)
	if skillPath == "" {
		return errors.New("cannot determine skill path")
	}

	// Create skill directory
	skillDir := filepath.Dir(skillPath)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		return fmt.Errorf("creating skill directory: %w", err)
	}

	// Generate skill file content
	content, err := formatSkillFile(s)
	if err != nil {
		return fmt.Errorf("formatting skill file: %w", err)
	}

	// Write skill file
	if err := os.WriteFile(skillPath, content, 0o644); err != nil {
		return fmt.Errorf("writing skill file: %w", err)
	}

	return nil
}

// Uninstall removes a skill by name.
// This operation is idempotent - returns nil if skill doesn't exist.
func (m *SkillManager) Uninstall(name string) error {
	if name == "" {
		return nil
	}

	skillPath := m.paths.SkillPath(name)
	if skillPath == "" {
		return nil
	}

	// Remove the skill directory (parent of SKILL.md)
	skillDir := filepath.Dir(skillPath)
	if err := os.RemoveAll(skillDir); err != nil {
		return fmt.Errorf("removing skill directory: %w", err)
	}

	return nil
}

// parseSkillFile parses a SKILL.md file with YAML frontmatter.
// The format is:
//
//	---
//	name: skill-name
//	description: What this skill does
//	...
//	---
//
//	Instructions go here as the body content.
func parseSkillFile(data []byte) (*Skill, error) {
	var skill Skill

	// Skills require frontmatter, so use MustParse
	body, err := frontmatter.MustParse(bytes.NewReader(data), &skill)
	if err != nil {
		return nil, fmt.Errorf("parsing frontmatter: %w", err)
	}

	// Set body content, trimming leading/trailing whitespace
	skill.Instructions = strings.TrimSpace(string(body))

	return &skill, nil
}

// formatSkillFile formats a Skill as a SKILL.md file with YAML frontmatter.
func formatSkillFile(s *Skill) ([]byte, error) {
	// Build frontmatter struct (only include non-empty optional fields)
	meta := struct {
		Name        string   `yaml:"name"`
		Description string   `yaml:"description"`
		Version     string   `yaml:"version,omitempty"`
		Author      string   `yaml:"author,omitempty"`
		Tools       []string `yaml:"tools,omitempty"`
		Triggers    []string `yaml:"triggers,omitempty"`
	}{
		Name:        s.Name,
		Description: s.Description,
		Version:     s.Version,
		Author:      s.Author,
		Tools:       s.Tools,
		Triggers:    s.Triggers,
	}

	return frontmatter.Format(meta, s.Instructions)
}
