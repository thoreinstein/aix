// Package cli provides CLI-specific types and utilities for the aix command.
package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/thoreinstein/aix/internal/paths"
	"github.com/thoreinstein/aix/internal/platform"
	"github.com/thoreinstein/aix/internal/platform/claude"
	"github.com/thoreinstein/aix/internal/platform/opencode"
)

// Sentinel errors for platform operations.
var (
	// ErrUnknownPlatform is returned when an unknown platform name is provided.
	ErrUnknownPlatform = errors.New("unknown platform")

	// ErrNoPlatformsAvailable is returned when no platforms are detected.
	ErrNoPlatformsAvailable = errors.New("no platforms available")
)

// SkillInfo provides a simplified view of a skill for CLI display.
// This is a platform-agnostic representation used for listing.
type SkillInfo struct {
	// Name is the skill's unique identifier.
	Name string

	// Description explains what the skill does.
	Description string

	// Source indicates where the skill came from: "local" or a git URL.
	Source string
}

// Platform defines the interface that platform adapters must implement
// for CLI operations. This is the consumer interface used by CLI commands.
type Platform interface {
	// Name returns the platform identifier (e.g., "claude", "opencode").
	Name() string

	// DisplayName returns a human-readable platform name (e.g., "Claude Code").
	DisplayName() string

	// IsAvailable checks if the platform is installed on this system.
	IsAvailable() bool

	// SkillDir returns the skills directory for the platform.
	SkillDir() string

	// InstallSkill installs a skill to the platform.
	// The skill parameter is platform-specific.
	InstallSkill(skill any) error

	// UninstallSkill removes a skill by name.
	UninstallSkill(name string) error

	// ListSkills returns information about all installed skills.
	ListSkills() ([]SkillInfo, error)

	// GetSkill retrieves a skill by name.
	// Returns the platform-specific skill type.
	GetSkill(name string) (any, error)
}

// claudeAdapter wraps ClaudePlatform to implement the Platform interface.
type claudeAdapter struct {
	p *claude.ClaudePlatform
}

func (a *claudeAdapter) Name() string {
	return a.p.Name()
}

func (a *claudeAdapter) DisplayName() string {
	return a.p.DisplayName()
}

func (a *claudeAdapter) IsAvailable() bool {
	return a.p.IsAvailable()
}

func (a *claudeAdapter) InstallSkill(skill any) error {
	s, ok := skill.(*claude.Skill)
	if !ok {
		return fmt.Errorf("expected *claude.Skill, got %T", skill)
	}
	return a.p.InstallSkill(s)
}

func (a *claudeAdapter) UninstallSkill(name string) error {
	return a.p.UninstallSkill(name)
}

func (a *claudeAdapter) ListSkills() ([]SkillInfo, error) {
	skills, err := a.p.ListSkills()
	if err != nil {
		return nil, err
	}

	infos := make([]SkillInfo, len(skills))
	for i, s := range skills {
		infos[i] = SkillInfo{
			Name:        s.Name,
			Description: s.Description,
			Source:      "local", // TODO: track source in skill metadata
		}
	}
	return infos, nil
}

func (a *claudeAdapter) GetSkill(name string) (any, error) {
	return a.p.GetSkill(name)
}

func (a *claudeAdapter) SkillDir() string {
	return a.p.SkillDir()
}

// opencodeAdapter wraps OpenCodePlatform to implement the Platform interface.
type opencodeAdapter struct {
	p *opencode.OpenCodePlatform
}

func (a *opencodeAdapter) Name() string {
	return a.p.Name()
}

func (a *opencodeAdapter) DisplayName() string {
	return a.p.DisplayName()
}

func (a *opencodeAdapter) IsAvailable() bool {
	return a.p.IsAvailable()
}

func (a *opencodeAdapter) InstallSkill(skill any) error {
	s, ok := skill.(*opencode.Skill)
	if !ok {
		return fmt.Errorf("expected *opencode.Skill, got %T", skill)
	}
	return a.p.InstallSkill(s)
}

func (a *opencodeAdapter) UninstallSkill(name string) error {
	return a.p.UninstallSkill(name)
}

func (a *opencodeAdapter) ListSkills() ([]SkillInfo, error) {
	skills, err := a.p.ListSkills()
	if err != nil {
		return nil, err
	}

	infos := make([]SkillInfo, len(skills))
	for i, s := range skills {
		infos[i] = SkillInfo{
			Name:        s.Name,
			Description: s.Description,
			Source:      "local", // TODO: track source in skill metadata
		}
	}
	return infos, nil
}

func (a *opencodeAdapter) GetSkill(name string) (any, error) {
	return a.p.GetSkill(name)
}

func (a *opencodeAdapter) SkillDir() string {
	return a.p.SkillDir()
}

// NewPlatform creates a Platform adapter for the given platform name.
// Returns ErrUnknownPlatform if the platform name is not recognized.
func NewPlatform(name string) (Platform, error) {
	switch name {
	case paths.PlatformClaude:
		return &claudeAdapter{p: claude.NewClaudePlatform()}, nil
	case paths.PlatformOpenCode:
		return &opencodeAdapter{p: opencode.NewOpenCodePlatform()}, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownPlatform, name)
	}
}

// ResolvePlatforms returns Platform instances for the given platform names.
// If names is empty, returns all detected/installed platforms.
// Returns an error if any platform name is invalid or if no platforms are available.
func ResolvePlatforms(names []string) ([]Platform, error) {
	// If no names specified, use all detected platforms
	if len(names) == 0 {
		detected := platform.DetectInstalled()
		if len(detected) == 0 {
			return nil, ErrNoPlatformsAvailable
		}

		platforms := make([]Platform, 0, len(detected))
		for _, d := range detected {
			// Only include platforms we have adapters for
			p, err := NewPlatform(d.Name)
			if err != nil {
				continue // Skip platforms without adapters
			}
			platforms = append(platforms, p)
		}

		if len(platforms) == 0 {
			return nil, ErrNoPlatformsAvailable
		}
		return platforms, nil
	}

	// Validate and create platforms for the specified names
	var invalid []string
	platforms := make([]Platform, 0, len(names))

	for _, name := range names {
		if !paths.ValidPlatform(name) {
			invalid = append(invalid, name)
			continue
		}

		p, err := NewPlatform(name)
		if err != nil {
			invalid = append(invalid, name)
			continue
		}
		platforms = append(platforms, p)
	}

	if len(invalid) > 0 {
		return nil, fmt.Errorf("%w: %s (valid: %s)",
			ErrUnknownPlatform,
			strings.Join(invalid, ", "),
			strings.Join(paths.Platforms(), ", "))
	}

	return platforms, nil
}
