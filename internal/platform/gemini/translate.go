package gemini

import (
	"regexp"
	"strings"

	"github.com/thoreinstein/aix/internal/errors"
)

// Variables supported by Gemini CLI.
const (
	VarArguments = "$ARGUMENTS"
	VarSelection = "$SELECTION"
)

// platformVars maps canonical variables to Gemini CLI format.
var platformVars = map[string]string{
	VarArguments: "{{argument}}",
	VarSelection: "{{selection}}",
}

// canonicalVars maps Gemini CLI variables back to canonical format.
var canonicalVars = map[string]string{
	"{{argument}}":  VarArguments,
	"{{args}}":      VarArguments, // Support both for compatibility
	"{{selection}}": VarSelection,
}

// varPattern matches variable syntax: $ followed by 2+ uppercase letters/underscores.
var varPattern = regexp.MustCompile(`\$[A-Z][A-Z_]+\b`)

// ErrUnsupportedVariable indicates content contains variables not supported by Gemini CLI.
var ErrUnsupportedVariable = errors.New("unsupported variable")

// TranslateVariables converts canonical variable syntax to Gemini CLI format.
// $ARGUMENTS -> {{argument}}
func TranslateVariables(content string) string {
	result := content
	for can, plat := range platformVars {
		result = strings.ReplaceAll(result, can, plat)
	}
	return result
}

// TranslateToCanonical converts Gemini CLI variable syntax to canonical format.
// {{args}} -> $ARGUMENTS
// {{argument}} -> $ARGUMENTS
func TranslateToCanonical(content string) string {
	result := content
	for plat, can := range canonicalVars {
		result = strings.ReplaceAll(result, plat, can)
	}
	return result
}

// ValidateVariables checks if content contains only supported variables.
func ValidateVariables(content string) error {
	vars := varPattern.FindAllString(content, -1)
	if len(vars) == 0 {
		return nil
	}

	var unsupported []string
	seen := make(map[string]struct{})

	for _, v := range vars {
		if _, ok := platformVars[v]; !ok {
			if _, alreadySeen := seen[v]; !alreadySeen {
				unsupported = append(unsupported, v)
				seen[v] = struct{}{}
			}
		}
	}

	if len(unsupported) == 0 {
		return nil
	}

	return errors.Wrapf(ErrUnsupportedVariable, "%s", strings.Join(unsupported, ", "))
}

// ListVariables returns all canonical variables found in the content.
func ListVariables(content string) []string {
	matches := varPattern.FindAllString(content, -1)
	if len(matches) == 0 {
		return []string{}
	}

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
