// Package doctor provides diagnostic checks for aix configuration.
package doctor

// Severity indicates the importance level of a check result.
type Severity int

const (
	// SeverityPass indicates the check passed without issues.
	SeverityPass Severity = iota

	// SeverityInfo indicates informational output, not a problem.
	SeverityInfo

	// SeverityWarning indicates a potential issue that doesn't prevent operation.
	SeverityWarning

	// SeverityError indicates a problem that prevents proper operation.
	SeverityError
)

// String returns the string representation of the severity level.
func (s Severity) String() string {
	switch s {
	case SeverityPass:
		return "pass"
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warning"
	case SeverityError:
		return "error"
	default:
		return "unknown"
	}
}

// CheckResult represents the outcome of a single diagnostic check.
type CheckResult struct {
	// Name is the identifier for this check.
	Name string `json:"name"`

	// Category groups related checks (e.g., "platform", "config", "mcp").
	Category string `json:"category"`

	// Status indicates the severity of the check result.
	Status Severity `json:"status"`

	// Message describes the check outcome.
	Message string `json:"message"`

	// Details contains additional context about the check result.
	// Keys and values depend on the specific check.
	Details map[string]any `json:"details,omitempty"`

	// Fixable indicates whether aix can automatically fix this issue.
	Fixable bool `json:"fixable,omitempty"`

	// FixHint provides guidance on how to resolve the issue.
	FixHint string `json:"fix_hint,omitempty"`
}

// Summary aggregates counts of check results by severity.
type Summary struct {
	// Passed is the count of checks with SeverityPass.
	Passed int `json:"passed"`

	// Info is the count of checks with SeverityInfo.
	Info int `json:"info"`

	// Warnings is the count of checks with SeverityWarning.
	Warnings int `json:"warnings"`

	// Errors is the count of checks with SeverityError.
	Errors int `json:"errors"`
}
