package validator

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/fatih/color"
)

// Format specifies the output format for validation reports.
type Format string

const (
	// FormatText produces human-readable text output.
	FormatText Format = "text"
	// FormatJSON produces machine-readable JSON output.
	FormatJSON Format = "json"
)

// Reporter formats and writes validation results.
type Reporter struct {
	out    io.Writer
	format Format
}

// NewReporter creates a new Reporter.
func NewReporter(out io.Writer, format Format) *Reporter {
	return &Reporter{
		out:    out,
		format: format,
	}
}

// Report writes the validation result to the output.
func (r *Reporter) Report(result *Result) error {
	if result == nil {
		return nil
	}

	switch r.format {
	case FormatJSON:
		return r.reportJSON(result)
	default:
		return r.reportText(result)
	}
}

// reportJSON writes the result as JSON.
func (r *Reporter) reportJSON(result *Result) error {
	encoder := json.NewEncoder(r.out)
	encoder.SetIndent("", "  ")
	return errors.Wrap(encoder.Encode(result), "encoding JSON report")
}

// reportText writes the result as human-readable text.
func (r *Reporter) reportText(result *Result) error {
	if !result.HasErrors() && !result.HasWarnings() {
		fmt.Fprintln(r.out, color.GreenString("✓ Validation passed"))
		return nil
	}

	// Group issues by severity
	errors := result.Errors()
	warnings := result.Warnings()

	// Print Summary
	summary := []string{}
	if len(errors) > 0 {
		summary = append(summary, color.RedString("%d error(s)", len(errors)))
	}
	if len(warnings) > 0 {
		summary = append(summary, color.YellowString("%d warning(s)", len(warnings)))
	}
	fmt.Fprintf(r.out, "Validation failed: %s\n\n", strings.Join(summary, ", "))

	// Print Errors
	if len(errors) > 0 {
		fmt.Fprintln(r.out, "Errors:")
		for _, err := range errors {
			r.printIssue(err, color.FgRed)
		}
		fmt.Fprintln(r.out)
	}

	// Print Warnings
	if len(warnings) > 0 {
		fmt.Fprintln(r.out, "Warnings:")
		for _, warn := range warnings {
			r.printIssue(warn, color.FgYellow)
		}
		fmt.Fprintln(r.out)
	}

	return nil
}

func (r *Reporter) printIssue(i Issue, c color.Attribute) {
	printer := color.New(c).SprintFunc()

	// Format:  • [field] message (context)

	var sb strings.Builder
	sb.WriteString("  • ")

	if i.Field != "" {
		sb.WriteString(printer(i.Field))
		sb.WriteString(": ")
	}

	sb.WriteString(i.Message)

	// Add context if present
	if len(i.Context) > 0 {
		var ctxParts []string
		for k, v := range i.Context {
			ctxParts = append(ctxParts, fmt.Sprintf("%s=%s", k, v))
		}
		// Sort for deterministic output
		sort.Strings(ctxParts)

		sb.WriteString(" ")
		sb.WriteString(color.New(color.FgHiBlack).Sprintf("(%s)", strings.Join(ctxParts, ", ")))
	}

	if i.Value != nil {
		valStr := fmt.Sprintf("%v", i.Value)
		// Truncate long values
		if len(valStr) > 50 {
			valStr = valStr[:47] + "..."
		}
		sb.WriteString(color.New(color.FgHiBlack).Sprintf(" [%s]", valStr))
	}

	fmt.Fprintln(r.out, sb.String())
}
