// Package toolperm provides parsing and validation for tool permission syntax
// as defined in the Agent Skills Specification.
package toolperm

import (
	"regexp"
	"strings"
)

// Permission represents a parsed tool permission.
type Permission struct {
	// Name is the tool name (e.g., "Read", "Bash", "Write")
	Name string

	// Scope is the optional scope specification (e.g., "git:*" from "Bash(git:*)")
	// Empty string if no scope is specified.
	Scope string
}

// String returns the permission in its canonical string form.
func (p Permission) String() string {
	if p.Scope == "" {
		return p.Name
	}
	return p.Name + "(" + p.Scope + ")"
}

// toolRegex matches tool permission syntax: ToolName or ToolName(scope)
// Tool names must be PascalCase per the Agent Skills Specification:
// - Must start with an uppercase letter [A-Z]
// - Followed by zero or more alphanumeric characters [a-zA-Z0-9]
// Captures: group 1 = tool name, group 2 = scope (optional, without parens)
var toolRegex = regexp.MustCompile(`^([A-Z][a-zA-Z0-9]*)(?:\(([^)]+)\))?$`)

// Parser handles tool permission string parsing.
type Parser struct{}

// New creates a new Parser instance.
func New() *Parser {
	return &Parser{}
}

// Parse parses a space-delimited allowed-tools string into individual permissions.
// Returns an empty slice for empty input.
func (p *Parser) Parse(allowedTools string) ([]Permission, error) {
	allowedTools = strings.TrimSpace(allowedTools)
	if allowedTools == "" {
		return []Permission{}, nil
	}

	tokens := strings.Fields(allowedTools)
	perms := make([]Permission, 0, len(tokens))

	for _, token := range tokens {
		perm, err := p.ParseSingle(token)
		if err != nil {
			return nil, err
		}
		perms = append(perms, perm)
	}

	return perms, nil
}

// ParseSingle parses a single tool permission token.
func (p *Parser) ParseSingle(token string) (Permission, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return Permission{}, &ToolPermError{Token: token, Message: "empty tool permission"}
	}

	matches := toolRegex.FindStringSubmatch(token)
	if matches == nil {
		return Permission{}, &ToolPermError{
			Token:   token,
			Message: "invalid tool permission syntax: tool name must be PascalCase (start with uppercase letter, e.g., Read, Write, Bash)",
		}
	}

	return Permission{
		Name:  matches[1],
		Scope: matches[2], // Will be empty string if no capture
	}, nil
}

// Format converts a slice of permissions back to space-delimited string.
func (p *Parser) Format(perms []Permission) string {
	if len(perms) == 0 {
		return ""
	}

	parts := make([]string, len(perms))
	for i, perm := range perms {
		parts[i] = perm.String()
	}
	return strings.Join(parts, " ")
}
