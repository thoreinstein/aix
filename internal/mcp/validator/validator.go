package validator

import (
	"slices"

	"github.com/thoreinstein/aix/internal/mcp"
	"github.com/thoreinstein/aix/internal/validator"
)

// validPlatforms is the set of valid platform identifiers.
var validPlatforms = []string{"darwin", "linux", "windows"}

// validTransports is the set of valid transport values.
var validTransports = []string{mcp.TransportStdio, mcp.TransportSSE, ""}

// Option configures a Validator.
type Option func(*Validator)

// Validator validates canonical MCP configurations.
type Validator struct {
	// allowEmpty permits configs with no servers.
	// Default is false (at least one server required).
	allowEmpty bool
}

// New creates a new Validator with the given options.
func New(opts ...Option) *Validator {
	v := &Validator{
		allowEmpty: false,
	}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

// WithAllowEmpty configures whether empty configs (no servers) are allowed.
// Default is false, meaning at least one server is required.
func WithAllowEmpty(allow bool) Option {
	return func(v *Validator) {
		v.allowEmpty = allow
	}
}

// Validate checks a Config for issues.
// Returns a Result containing errors and warnings.
func (v *Validator) Validate(cfg *mcp.Config) *validator.Result {
	result := &validator.Result{}
	if cfg == nil {
		result.AddError("", "config is nil", nil)
		return result
	}

	// Check for empty config
	if !v.allowEmpty && len(cfg.Servers) == 0 {
		result.AddError("", "config has no servers", nil)
	}

	// Validate each server
	for name, server := range cfg.Servers {
		v.validateServer(name, server, result)
	}

	return result
}

// validateServer validates a single server configuration.
func (v *Validator) validateServer(name string, server *mcp.Server, result *validator.Result) {
	context := map[string]string{"server": name}

	// Server name validation
	if server.Name == "" {
		result.Issues = append(result.Issues, validator.Issue{
			Severity: validator.SeverityError,
			Field:    "name",
			Message:  "server name is required",
			Context:  context,
		})
	}

	// Validate transport value
	if !slices.Contains(validTransports, server.Transport) {
		result.Issues = append(result.Issues, validator.Issue{
			Severity: validator.SeverityError,
			Field:    "transport",
			Message:  "transport must be 'stdio', 'sse', or empty",
			Context:  context,
		})
	}

	// Determine effective transport and validate required fields
	v.validateTransportFields(name, server, result)

	// Validate platforms
	v.validatePlatforms(name, server, result)

	// Validate env keys
	v.validateEnv(name, server, result)

	// Validate header keys
	v.validateHeaders(name, server, result)
}

// validateTransportFields validates that the server has the required fields
// for its transport type.
func (v *Validator) validateTransportFields(name string, server *mcp.Server, result *validator.Result) {
	context := map[string]string{"server": name}

	// Determine effective transport
	isLocal := server.IsLocal()
	isRemote := server.IsRemote()

	// If explicit transport is set, validate required fields
	switch server.Transport {
	case mcp.TransportStdio:
		if server.Command == "" {
			result.Issues = append(result.Issues, validator.Issue{
				Severity: validator.SeverityError,
				Field:    "command",
				Message:  "stdio transport requires command",
				Context:  context,
			})
		}
	case mcp.TransportSSE:
		if server.URL == "" {
			result.Issues = append(result.Issues, validator.Issue{
				Severity: validator.SeverityError,
				Field:    "url",
				Message:  "sse transport requires URL",
				Context:  context,
			})
		}
	case "":
		// No explicit transport - infer from fields
		if server.Command == "" && server.URL == "" {
			result.Issues = append(result.Issues, validator.Issue{
				Severity: validator.SeverityError,
				Field:    "command/url",
				Message:  "server must have command (for local) or URL (for remote)",
				Context:  context,
			})
		}
	}

	// Warn about ambiguous configuration
	if server.Command != "" && server.URL != "" {
		msg := "server has both command and URL"
		if isLocal {
			msg += "; transport=stdio means command will be used"
		} else if isRemote {
			msg += "; transport=sse means URL will be used"
		} else {
			msg += "; without explicit transport, command takes precedence"
		}
		result.Issues = append(result.Issues, validator.Issue{
			Severity: validator.SeverityWarning,
			Message:  msg,
			Context:  context,
		})
	}
}

// validatePlatforms validates that all platform values are recognized.
func (v *Validator) validatePlatforms(name string, server *mcp.Server, result *validator.Result) {
	context := map[string]string{"server": name}
	for _, platform := range server.Platforms {
		if !slices.Contains(validPlatforms, platform) {
			result.Issues = append(result.Issues, validator.Issue{
				Severity: validator.SeverityError,
				Field:    "platforms",
				Message:  "invalid platform: " + platform + " (valid: darwin, linux, windows)",
				Value:    platform,
				Context:  context,
			})
		}
	}
}

// validateEnv validates that environment variable keys are non-empty.
func (v *Validator) validateEnv(name string, server *mcp.Server, result *validator.Result) {
	context := map[string]string{"server": name}
	for key := range server.Env {
		if key == "" {
			result.Issues = append(result.Issues, validator.Issue{
				Severity: validator.SeverityError,
				Field:    "env",
				Message:  "environment variable key cannot be empty",
				Context:  context,
			})
			break // Only report once
		}
	}
}

// validateHeaders validates that HTTP header keys are non-empty.
func (v *Validator) validateHeaders(name string, server *mcp.Server, result *validator.Result) {
	context := map[string]string{"server": name}
	for key := range server.Headers {
		if key == "" {
			result.Issues = append(result.Issues, validator.Issue{
				Severity: validator.SeverityError,
				Field:    "headers",
				Message:  "header key cannot be empty",
				Context:  context,
			})
			break // Only report once
		}
	}
}
