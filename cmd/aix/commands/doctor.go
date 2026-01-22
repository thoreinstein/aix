package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/doctor"
)

var (
	doctorJSON    bool
	doctorQuiet   bool
	doctorVerbose bool
)

func init() {
	doctorCmd.Flags().BoolVar(&doctorJSON, "json", false,
		"output results as JSON")
	doctorCmd.Flags().BoolVar(&doctorQuiet, "quiet", false,
		"suppress output, exit code only")
	doctorCmd.Flags().BoolVar(&doctorVerbose, "verbose", false,
		"show detailed check-by-check output")
	rootCmd.AddCommand(doctorCmd)
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose configuration issues",
	Long: `Run diagnostic checks on aix and platform configurations.

Validates configuration files, checks platform detection, and identifies
potential issues before they cause problems.

Output modes (mutually exclusive):
  (default)   Show errors and warnings
  --verbose   Show all checks including passed ones
  --quiet     No output, exit code only
  --json      Machine-readable JSON output

Exit codes:
  0 - All checks passed (no errors or warnings)
  1 - Warnings present, no errors
  2 - Errors present`,
	PreRunE: validateDoctorFlags,
	RunE:    runDoctor,
}

// validateDoctorFlags ensures output flags are mutually exclusive.
func validateDoctorFlags(_ *cobra.Command, _ []string) error {
	count := 0
	if doctorJSON {
		count++
	}
	if doctorQuiet {
		count++
	}
	if doctorVerbose {
		count++
	}

	if count > 1 {
		return errors.New("flags --json, --quiet, and --verbose are mutually exclusive")
	}

	return nil
}

func runDoctor(_ *cobra.Command, _ []string) error {
	runner := doctor.NewRunner()

	// TODO: Add checks here in future tickets (aix-6ei.1.1, 1.2, 1.4)

	report := runner.Run()

	if err := outputDoctorReport(report); err != nil {
		return err
	}

	// Determine exit code based on results
	if report.HasErrors() {
		return errDoctorErrors
	}
	if report.HasWarnings() {
		return errDoctorWarnings
	}
	return nil
}

func outputDoctorReport(report *doctor.DoctorReport) error {
	if doctorQuiet {
		return nil
	}

	if doctorJSON {
		return outputDoctorJSON(report)
	}

	return outputDoctorText(report)
}

func outputDoctorJSON(report *doctor.DoctorReport) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}
	return nil
}

func outputDoctorText(report *doctor.DoctorReport) error {
	// In normal mode, show only errors and warnings
	// In verbose mode, show all checks
	showAll := doctorVerbose

	hasOutput := false
	for _, result := range report.Results {
		if !showAll && result.Status != doctor.SeverityError && result.Status != doctor.SeverityWarning {
			continue
		}

		hasOutput = true
		icon := statusIcon(result.Status)
		fmt.Printf("%s [%s] %s: %s\n", icon, result.Category, result.Name, result.Message)

		if result.FixHint != "" && (result.Status == doctor.SeverityError || result.Status == doctor.SeverityWarning) {
			fmt.Printf("  hint: %s\n", result.FixHint)
		}
	}

	// Print summary
	if hasOutput || showAll {
		fmt.Println()
	}

	fmt.Printf("Summary: %d passed, %d info, %d warnings, %d errors\n",
		report.Summary.Passed, report.Summary.Info, report.Summary.Warnings, report.Summary.Errors)

	return nil
}

func statusIcon(s doctor.Severity) string {
	switch s {
	case doctor.SeverityPass:
		return "✓"
	case doctor.SeverityInfo:
		return "ℹ"
	case doctor.SeverityWarning:
		return "⚠"
	case doctor.SeverityError:
		return "✗"
	default:
		return "?"
	}
}

// errDoctorWarnings is a sentinel error for exit code 1.
var errDoctorWarnings = errors.New("warnings found")

// errDoctorErrors is a sentinel error for exit code 2.
var errDoctorErrors = errors.New("errors found")
