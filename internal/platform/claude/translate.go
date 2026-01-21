package claude

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Variables supported by Claude Code.
const (
	VarArguments = "$ARGUMENTS"
	VarSelection = "$SELECTION"
)

// supportedVars is the set of variables recognized by Claude Code.
var supportedVars = map[string]struct{}{
	VarArguments: {},
	VarSelection: {},
}

// varPattern matches variable syntax: $ followed by 2+ uppercase letters/underscores.
// Uses negative lookahead simulation via atomic group behavior - the pattern stops matching
// when it hits lowercase letters (thanks to the character class).
// Examples: $ARGUMENTS matches, $VAR matches in "$VAR123", $Arguments doesn't match.
var varPattern = regexp.MustCompile(`\$[A-Z][A-Z_]+`)

// ErrUnsupportedVariable indicates content contains variables not supported by Claude Code.
var ErrUnsupportedVariable = errors.New("unsupported variable")

// TranslateVariables converts canonical variable syntax to Claude Code format.
// Since Claude Code uses the canonical format ($ARGUMENTS, $SELECTION),
// this is essentially a pass-through that preserves the content unchanged.
func TranslateVariables(content string) string {
	return content
}

// TranslateToCanonical converts Claude Code variable syntax to canonical format.
// Since Claude Code uses the canonical format, this is a pass-through.
func TranslateToCanonical(content string) string {
	return content
}

// ValidateVariables checks if content contains only supported variables.
// Returns nil if valid, or an error listing unsupported variables.
func ValidateVariables(content string) error {
	vars := varPattern.FindAllString(content, -1)
	if len(vars) == 0 {
		return nil
	}

	var unsupported []string
	seen := make(map[string]struct{})

	for _, v := range vars {
		if _, ok := supportedVars[v]; !ok {
			// Deduplicate unsupported variables
			if _, alreadySeen := seen[v]; !alreadySeen {
				unsupported = append(unsupported, v)
				seen[v] = struct{}{}
			}
		}
	}

	if len(unsupported) == 0 {
		return nil
	}

	return fmt.Errorf("%w: %s", ErrUnsupportedVariable, strings.Join(unsupported, ", "))
}

// ListVariables returns all variables found in the content.
// Returns an empty slice if no variables are found.
// The returned slice contains unique variables in the order they first appear.
func ListVariables(content string) []string {
	matches := varPattern.FindAllString(content, -1)
	if len(matches) == 0 {
		return []string{}
	}

	// Deduplicate while preserving order
	seen := make(map[string]struct{})
	result := make([]string, 0, len(matches))

	for _, v := range matches {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}

	return result
}
