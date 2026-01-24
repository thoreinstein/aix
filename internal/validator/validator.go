// Package validator provides a unified validation framework for aix.
package validator

import (
	"fmt"
	"strings"
)

// Severity represents the impact of a validation issue.
type Severity int

const (
	// SeverityError indicates a blocking validation failure.
	SeverityError Severity = iota
	// SeverityWarning indicates a recommended but non-blocking issue.
	SeverityWarning
	// SeverityInfo indicates an informational note.
	SeverityInfo
)

func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	case SeverityInfo:
		return "info"
	default:
		return "unknown"
	}
}

// Issue represents a single validation problem.
type Issue struct {
	// Severity indicates the impact of the issue.
	Severity Severity
	// Field identifies the field with the issue (optional).
	Field string
	// Message is a human-readable description of the problem.
	Message string
	// Value is the actual value that failed validation (optional).
	Value any
	// Context is additional platform-specific or domain-specific context.
	Context map[string]string
}

// Error implements the error interface.
func (i Issue) Error() string {
	var sb strings.Builder
	sb.WriteString(i.Severity.String())
	sb.WriteString(": ")
	if i.Field != "" {
		sb.WriteString("field \"")
		sb.WriteString(i.Field)
		sb.WriteString("\": ")
	}
	sb.WriteString(i.Message)
	if i.Value != nil {
		fmt.Fprintf(&sb, " (got %v)", i.Value)
	}
	return sb.String()
}

// Result aggregates validation issues.
type Result struct {
	Issues []Issue
}

// HasErrors returns true if any issue has SeverityError.
func (r *Result) HasErrors() bool {
	if r == nil {
		return false
	}
	for _, i := range r.Issues {
		if i.Severity == SeverityError {
			return true
		}
	}
	return false
}

// HasWarnings returns true if any issue has SeverityWarning.
func (r *Result) HasWarnings() bool {
	if r == nil {
		return false
	}
	for _, i := range r.Issues {
		if i.Severity == SeverityWarning {
			return true
		}
	}
	return false
}

// AddError adds an error issue to the result.
func (r *Result) AddError(field, message string, value any) {
	r.Issues = append(r.Issues, Issue{
		Severity: SeverityError,
		Field:    field,
		Message:  message,
		Value:    value,
	})
}

// AddWarning adds a warning issue to the result.
func (r *Result) AddWarning(field, message string, value any) {
	r.Issues = append(r.Issues, Issue{
		Severity: SeverityWarning,
		Field:    field,
		Message:  message,
		Value:    value,
	})
}

// AddInfo adds an info issue to the result.
func (r *Result) AddInfo(field, message string, value any) {
	r.Issues = append(r.Issues, Issue{
		Severity: SeverityInfo,
		Field:    field,
		Message:  message,
		Value:    value,
	})
}

// Errors returns a slice of all issues with SeverityError.
func (r *Result) Errors() []Issue {
	if r == nil {
		return nil
	}
	var res []Issue
	for _, i := range r.Issues {
		if i.Severity == SeverityError {
			res = append(res, i)
		}
	}
	return res
}

// Warnings returns a slice of all issues with SeverityWarning.
func (r *Result) Warnings() []Issue {
	if r == nil {
		return nil
	}
	var res []Issue
	for _, i := range r.Issues {
		if i.Severity == SeverityWarning {
			res = append(res, i)
		}
	}
	return res
}
