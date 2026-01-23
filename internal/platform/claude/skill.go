package claude

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/errors"

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
		return nil, errors.Wrap(err, "reading skill directory")
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

		name := entry.Name()
		skillPath := m.paths.SkillPath(name)
		if skillPath == "" {
			continue
		}

		// Check if SKILL.md exists
		if _, err := os.Stat(skillPath); os.IsNotExist(err) {
			continue
		}

		f, err := os.Open(skillPath)
		if err != nil {
			return nil, errors.Wrapf(err, "opening skill file %q", name)
		}

		skill := &Skill{Name: name}
		if err := frontmatter.ParseHeader(f, skill); err != nil {
			f.Close()
			return nil, errors.Wrapf(err, "parsing skill header %q", name)
		}
		f.Close()

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
		return nil, errors.Wrap(err, "reading skill file")
	}

	skill, err := parseSkillFile(data)
	if err != nil {
		return nil, errors.Wrap(err, "parsing skill file")
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
		return errors.Wrap(err, "creating skill directory")
	}

	// Generate skill file content
	content, err := formatSkillFile(s)
	if err != nil {
		return errors.Wrap(err, "formatting skill file")
	}

	// Write skill file
	if err := os.WriteFile(skillPath, content, 0o644); err != nil {
		return errors.Wrap(err, "writing skill file")
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
		return errors.Wrap(err, "removing skill directory")
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
		return nil, errors.Wrap(err, "parsing frontmatter")
	}

	// Set body content, trimming leading/trailing whitespace
	skill.Instructions = strings.TrimSpace(string(body))

	return &skill, nil
}

// formatSkillFile formats a Skill as a SKILL.md file with YAML frontmatter.
func formatSkillFile(s *Skill) ([]byte, error) {
	// Build frontmatter struct (only include non-empty optional fields)
	meta := struct {
		Name          string            `yaml:"name"`
		Description   string            `yaml:"description"`
		License       string            `yaml:"license,omitempty"`
		Compatibility []string          `yaml:"compatibility,omitempty"`
		Metadata      map[string]string `yaml:"metadata,omitempty"`
		AllowedTools  string            `yaml:"allowed-tools,omitempty"`
	}{
		Name:          s.Name,
		Description:   s.Description,
		License:       s.License,
		Compatibility: s.Compatibility,
		Metadata:      s.Metadata,
		AllowedTools:  s.AllowedTools.String(),
	}

	return frontmatter.Format(meta, s.Instructions)
}
