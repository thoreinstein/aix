package mcp

import "github.com/thoreinstein/aix/internal/errors"

// Translator converts between canonical and platform-specific MCP formats.
//
// Each platform adapter (Claude Code, OpenCode, Codex, Gemini CLI) implements
// this interface to enable bidirectional translation of MCP server configurations.
//
// # Translation Flow
//
// When reading from a platform config:
//
//	platformJSON -> Translator.ToCanonical() -> *Config
//
// When writing to a platform config:
//
//	*Config -> Translator.FromCanonical() -> platformJSON
//
// # Unknown Field Preservation
//
// Implementations MUST preserve unknown fields during translation. Platform
// configs may contain fields not defined in the canonical format, and these
// must be retained through the round-trip to avoid data loss.
//
// # Error Handling
//
// Implementations should return [ErrFieldNotSupported] when encountering a
// canonical field that cannot be represented in the platform format. This
// allows callers to warn users about potential data loss.
//
// Implementations should return [ErrRequiredFieldMissing] when the platform
// data is missing a field required for valid canonical representation.
//
// # Example Implementation
//
//	type ClaudeTranslator struct{}
//
//	func (t *ClaudeTranslator) ToCanonical(platformData []byte) (*Config, error) {
//	    var claude ClaudeMCPConfig
//	    if err := json.Unmarshal(platformData, &claude); err != nil {
//	        return nil, err
//	    }
//	    // Convert ClaudeMCPConfig -> canonical *Config
//	    return convertToCanonical(claude), nil
//	}
//
//	func (t *ClaudeTranslator) FromCanonical(cfg *Config) ([]byte, error) {
//	    claude := convertFromCanonical(cfg)
//	    return json.MarshalIndent(claude, "", "  ")
//	}
//
//	func (t *ClaudeTranslator) Platform() string {
//	    return "claude"
//	}
type Translator interface {
	// ToCanonical converts platform-specific MCP configuration to canonical format.
	//
	// The platformData parameter contains raw JSON from the platform's config file.
	// For example, Claude Code uses a "mcpServers" key with server configs, while
	// OpenCode uses an "mcp" key with a different structure.
	//
	// Returns the canonical Config representation, or an error if the platform
	// data is malformed or cannot be converted.
	//
	// Unknown fields in the platform data should be captured and stored in the
	// canonical types' unknownFields for preservation during round-trip.
	ToCanonical(platformData []byte) (*Config, error)

	// FromCanonical converts canonical MCP configuration to platform-specific format.
	//
	// Returns JSON bytes formatted according to the platform's expected structure.
	// The output should be suitable for writing directly to the platform's config file.
	//
	// Returns [ErrFieldNotSupported] if the canonical config contains fields that
	// cannot be represented in the platform format (caller should warn the user).
	//
	// Unknown fields stored in the canonical types should be included in the output
	// if the platform format supports them.
	FromCanonical(cfg *Config) ([]byte, error)

	// Platform returns the name of the platform this translator handles.
	//
	// This is used for error messages, logging, and registry lookups.
	// Expected values: "claude", "opencode", "codex", "gemini"
	Platform() string
}

// Sentinel errors for translation operations.
var (
	// ErrFieldNotSupported indicates a canonical field cannot be represented
	// in the target platform format. Callers should warn users about potential
	// data loss when this error is encountered.
	//
	// Example: A canonical server has both Command and URL, but the target
	// platform only supports one transport type.
	ErrFieldNotSupported = errors.New("field not supported by platform")

	// ErrRequiredFieldMissing indicates the platform data is missing a field
	// required to construct a valid canonical representation.
	//
	// Example: A platform server config has neither "command" nor "url",
	// making it impossible to determine the transport type.
	ErrRequiredFieldMissing = errors.New("required field missing from platform data")
)
