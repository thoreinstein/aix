// Package validator provides validation for Skill structs per the Agent Skills Specification.
package validator

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/thoreinstein/aix/internal/platform/claude"
	"github.com/thoreinstein/aix/internal/skill/toolperm"
	"github.com/thoreinstein/aix/internal/validator"
)

const (
	// maxNameLength is the maximum allowed length for skill names.
	maxNameLength = 64
)

// nameRegex validates skill names: must start with a letter, lowercase alphanumeric,
// single hyphens allowed between segments, no start/end hyphen, no consecutive hyphens.
var nameRegex = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

// Option configures a Validator.
type Option func(*Validator)

// Validator validates Skill structs against the Agent Skills Specification.
type Validator struct {
	toolParser *toolperm.Parser
	strict     bool
}

// New creates a new Validator with the given options.
func New(opts ...Option) *Validator {
	v := &Validator{
		toolParser: toolperm.New(),
		strict:     false,
	}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

// WithStrict enables strict validation mode.
// In strict mode, AllowedTools syntax is validated using the toolperm parser.
func WithStrict(strict bool) Option {
	return func(v *Validator) {
		v.strict = strict
	}
}

// Validate checks a Skill for compliance with the Agent Skills Specification.
// Returns a Result containing errors and warnings.
func (v *Validator) Validate(s *claude.Skill) *validator.Result {
	result := &validator.Result{}

	v.validateName(s.Name, result)
	v.validateDescription(s.Description, result)

	if v.strict && len(s.AllowedTools) > 0 {
		v.validateAllowedTools(s.AllowedTools.String(), result)
	}

	return result
}

// ValidateWithPath validates a Skill and additionally checks that the skill name
// matches the containing directory name. The path should be the path to the skill file.
func (v *Validator) ValidateWithPath(s *claude.Skill, path string) *validator.Result {
	result := v.Validate(s)

	if s.Name != "" {
		dir := filepath.Base(filepath.Dir(path))
		if dir != s.Name {
			result.Issues = append(result.Issues, validator.Issue{
				Severity: validator.SeverityError,
				Field:    "name",
				Message:  "skill name must match directory name",
				Value:    s.Name,
				Context: map[string]string{
					"directory": dir,
					"path":      path,
				},
			})
		}
	}

	return result
}

// validateName checks the name field for compliance.
func (v *Validator) validateName(name string, result *validator.Result) {
	if name == "" {
		result.AddError("name", "is required", "")
		return
	}

	if len(name) > maxNameLength {
		result.AddError("name", "exceeds maximum length of 64 characters", name)
	}

	if !nameRegex.MatchString(name) {
		msg := "name must start with a letter, be lowercase alphanumeric with single hyphens between segments"
		if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
			msg = "name cannot start or end with a hyphen"
		} else if strings.Contains(name, "--") {
			msg = "name cannot contain consecutive hyphens"
		} else if strings.ToLower(name) != name {
			msg = "name must be lowercase"
		}
		result.AddError("name", msg, name)
	}
}

// validateDescription checks the description field for compliance.
func (v *Validator) validateDescription(description string, result *validator.Result) {
	if description == "" {
		result.AddError("description", "is required", "")
		return
	}

	if strings.TrimSpace(description) == "" {
		result.AddError("description", "cannot be only whitespace", description)
	}
}

// validateAllowedTools validates the AllowedTools syntax using the toolperm parser.
func (v *Validator) validateAllowedTools(allowedTools string, result *validator.Result) {
	_, err := v.toolParser.Parse(allowedTools)
	if err != nil {
		result.AddError("allowed-tools", err.Error(), allowedTools)
	}
}
