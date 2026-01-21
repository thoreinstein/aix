// Package validator provides validation for Skill structs per the Agent Skills Specification.
package validator

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/thoreinstein/aix/internal/platform/claude"
	"github.com/thoreinstein/aix/internal/skill/toolperm"
)

const (
	// maxNameLength is the maximum allowed length for skill names.
	maxNameLength = 64
)

// nameRegex validates skill names: lowercase alphanumeric, single hyphens allowed
// between segments, no start/end hyphen, no consecutive hyphens.
var nameRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

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
// Returns a slice of validation errors, or nil if valid.
func (v *Validator) Validate(s *claude.Skill) []error {
	var errs []error

	errs = append(errs, v.validateName(s.Name)...)
	errs = append(errs, v.validateDescription(s.Description)...)

	if v.strict && s.AllowedTools != "" {
		errs = append(errs, v.validateAllowedTools(s.AllowedTools)...)
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

// ValidateWithPath validates a Skill and additionally checks that the skill name
// matches the containing directory name. The path should be the path to the skill file.
func (v *Validator) ValidateWithPath(s *claude.Skill, path string) []error {
	errs := v.Validate(s)

	if s.Name != "" {
		dir := filepath.Base(filepath.Dir(path))
		if dir != s.Name {
			errs = append(errs, &ValidationError{
				Field:   "name",
				Message: "skill name must match directory name",
				Value:   s.Name,
				Context: map[string]string{
					"directory": dir,
					"path":      path,
				},
			})
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

// validateName checks the name field for compliance.
func (v *Validator) validateName(name string) []error {
	var errs []error

	if name == "" {
		errs = append(errs, &ValidationError{
			Field:   "name",
			Message: "name is required",
		})
		return errs
	}

	if len(name) > maxNameLength {
		errs = append(errs, &ValidationError{
			Field:   "name",
			Message: "name exceeds maximum length of 64 characters",
			Value:   name,
		})
	}

	if !nameRegex.MatchString(name) {
		msg := "name must be lowercase alphanumeric with single hyphens between segments"
		if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
			msg = "name cannot start or end with a hyphen"
		} else if strings.Contains(name, "--") {
			msg = "name cannot contain consecutive hyphens"
		} else if strings.ToLower(name) != name {
			msg = "name must be lowercase"
		}
		errs = append(errs, &ValidationError{
			Field:   "name",
			Message: msg,
			Value:   name,
		})
	}

	return errs
}

// validateDescription checks the description field for compliance.
func (v *Validator) validateDescription(description string) []error {
	var errs []error

	if description == "" {
		errs = append(errs, &ValidationError{
			Field:   "description",
			Message: "description is required",
		})
		return errs
	}

	if strings.TrimSpace(description) == "" {
		errs = append(errs, &ValidationError{
			Field:   "description",
			Message: "description cannot be only whitespace",
			Value:   description,
		})
	}

	return errs
}

// validateAllowedTools validates the AllowedTools syntax using the toolperm parser.
func (v *Validator) validateAllowedTools(allowedTools string) []error {
	var errs []error

	_, err := v.toolParser.Parse(allowedTools)
	if err != nil {
		errs = append(errs, &ValidationError{
			Field:   "allowed-tools",
			Message: err.Error(),
			Value:   allowedTools,
		})
	}

	return errs
}
