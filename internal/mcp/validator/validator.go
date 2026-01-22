package validator

import (
	"slices"

	"github.com/thoreinstein/aix/internal/mcp"
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
// Returns a slice of validation errors/warnings, or nil if valid.
// Use [HasErrors] to check if any errors (vs warnings) were found.
func (v *Validator) Validate(cfg *mcp.Config) []*ValidationError {
	if cfg == nil {
		return []*ValidationError{{
			Message:  "config is nil",
			Severity: SeverityError,
		}}
	}

	var errs []*ValidationError

	// Check for empty config
	if !v.allowEmpty && len(cfg.Servers) == 0 {
		errs = append(errs, &ValidationError{
			Message:  "config has no servers",
			Severity: SeverityError,
			Err:      ErrEmptyConfig,
		})
	}

	// Validate each server
	for name, server := range cfg.Servers {
		errs = append(errs, v.validateServer(name, server)...)
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

// validateServer validates a single server configuration.
func (v *Validator) validateServer(name string, server *mcp.Server) []*ValidationError {
	var errs []*ValidationError

	// Server name validation
	if server.Name == "" {
		errs = append(errs, &ValidationError{
			ServerName: name,
			Field:      "name",
			Message:    "server name is required",
			Severity:   SeverityError,
			Err:        ErrMissingServerName,
		})
	}

	// Validate transport value
	if !slices.Contains(validTransports, server.Transport) {
		errs = append(errs, &ValidationError{
			ServerName: name,
			Field:      "transport",
			Message:    "transport must be 'stdio', 'sse', or empty",
			Severity:   SeverityError,
			Err:        ErrInvalidTransport,
		})
	}

	// Determine effective transport and validate required fields
	errs = append(errs, v.validateTransportFields(name, server)...)

	// Validate platforms
	errs = append(errs, v.validatePlatforms(name, server)...)

	// Validate env keys
	errs = append(errs, v.validateEnv(name, server)...)

	// Validate header keys
	errs = append(errs, v.validateHeaders(name, server)...)

	return errs
}

// validateTransportFields validates that the server has the required fields
// for its transport type.
func (v *Validator) validateTransportFields(name string, server *mcp.Server) []*ValidationError {
	var errs []*ValidationError

	// Determine effective transport
	isLocal := server.IsLocal()
	isRemote := server.IsRemote()

	// If explicit transport is set, validate required fields
	switch server.Transport {
	case mcp.TransportStdio:
		if server.Command == "" {
			errs = append(errs, &ValidationError{
				ServerName: name,
				Field:      "command",
				Message:    "stdio transport requires command",
				Severity:   SeverityError,
				Err:        ErrMissingCommand,
			})
		}
	case mcp.TransportSSE:
		if server.URL == "" {
			errs = append(errs, &ValidationError{
				ServerName: name,
				Field:      "url",
				Message:    "sse transport requires URL",
				Severity:   SeverityError,
				Err:        ErrMissingURL,
			})
		}
	case "":
		// No explicit transport - infer from fields
		if server.Command == "" && server.URL == "" {
			errs = append(errs, &ValidationError{
				ServerName: name,
				Field:      "command/url",
				Message:    "server must have command (for local) or URL (for remote)",
				Severity:   SeverityError,
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
		errs = append(errs, &ValidationError{
			ServerName: name,
			Message:    msg,
			Severity:   SeverityWarning,
		})
	}

	return errs
}

// validatePlatforms validates that all platform values are recognized.
func (v *Validator) validatePlatforms(name string, server *mcp.Server) []*ValidationError {
	var errs []*ValidationError

	for _, platform := range server.Platforms {
		if !slices.Contains(validPlatforms, platform) {
			errs = append(errs, &ValidationError{
				ServerName: name,
				Field:      "platforms",
				Message:    "invalid platform: " + platform + " (valid: darwin, linux, windows)",
				Severity:   SeverityError,
				Err:        ErrInvalidPlatform,
			})
		}
	}

	return errs
}

// validateEnv validates that environment variable keys are non-empty.
func (v *Validator) validateEnv(name string, server *mcp.Server) []*ValidationError {
	var errs []*ValidationError

	for key := range server.Env {
		if key == "" {
			errs = append(errs, &ValidationError{
				ServerName: name,
				Field:      "env",
				Message:    "environment variable key cannot be empty",
				Severity:   SeverityError,
				Err:        ErrEmptyEnvKey,
			})
			break // Only report once
		}
	}

	return errs
}

// validateHeaders validates that HTTP header keys are non-empty.
func (v *Validator) validateHeaders(name string, server *mcp.Server) []*ValidationError {
	var errs []*ValidationError

	for key := range server.Headers {
		if key == "" {
			errs = append(errs, &ValidationError{
				ServerName: name,
				Field:      "headers",
				Message:    "header key cannot be empty",
				Severity:   SeverityError,
				Err:        ErrEmptyHeaderKey,
			})
			break // Only report once
		}
	}

	return errs
}
