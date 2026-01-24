package gemini

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/pkg/fileutil"
	"github.com/thoreinstein/aix/pkg/frontmatter"
)

// Sentinel errors for skill operations.
var (
	ErrSkillNotFound = errors.New("skill not found")
	ErrInvalidSkill  = errors.New("invalid skill: name required")
)

// SkillManager handles CRUD operations for Gemini CLI skills.
type SkillManager struct {
	paths *GeminiPaths
}

// NewSkillManager creates a new SkillManager with the given paths configuration.
func NewSkillManager(paths *GeminiPaths) *SkillManager {
	return &SkillManager{
		paths: paths,
	}
}

// List returns all skills in the skill directory.
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

	skill.Name = name
	return skill, nil
}

// Install creates or overwrites a skill.
func (m *SkillManager) Install(s *Skill) error {
	if s == nil || s.Name == "" {
		return ErrInvalidSkill
	}

	skillPath := m.paths.SkillPath(s.Name)
	if skillPath == "" {
		return errors.New("cannot determine skill path")
	}

	skillDir := filepath.Dir(skillPath)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		return errors.Wrap(err, "creating skill directory")
	}

	// Translate variables to Gemini format before formatting
	s.Instructions = TranslateVariables(s.Instructions)

	content, err := formatSkillFile(s)
	if err != nil {
		return errors.Wrap(err, "formatting skill file")
	}

	if err := fileutil.AtomicWriteFile(skillPath, content, 0o644); err != nil {
		return errors.Wrap(err, "writing skill file")
	}

	return nil
}

// Uninstall removes a skill by name.
func (m *SkillManager) Uninstall(name string) error {
	if name == "" {
		return nil
	}

	skillPath := m.paths.SkillPath(name)
	if skillPath == "" {
		return nil
	}

	skillDir := filepath.Dir(skillPath)
	if err := os.RemoveAll(skillDir); err != nil {
		return errors.Wrap(err, "removing skill directory")
	}

	return nil
}

func parseSkillFile(data []byte) (*Skill, error) {
	var skill Skill
	body, err := frontmatter.MustParse(bytes.NewReader(data), &skill)
	if err != nil {
		return nil, errors.Wrap(err, "parsing frontmatter")
	}

	// Translate variables back to canonical format when reading
	skill.Instructions = TranslateToCanonical(strings.TrimSpace(string(body)))

	return &skill, nil
}

func formatSkillFile(s *Skill) ([]byte, error) {
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

	data, err := frontmatter.Format(meta, s.Instructions)
	if err != nil {
		return nil, errors.Wrap(err, "formatting skill content")
	}
	return data, nil
}
