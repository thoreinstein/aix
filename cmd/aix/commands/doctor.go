package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/doctor"
)

var (
	doctorJSON    bool
	doctorQuiet   bool
	doctorVerbose bool
	doctorFix     bool
)

func init() {
	doctorCmd.Flags().BoolVar(&doctorJSON, "json", false,
		"output results as JSON")
	doctorCmd.Flags().BoolVar(&doctorQuiet, "quiet", false,
		"suppress output, exit code only")
	doctorCmd.Flags().BoolVar(&doctorVerbose, "verbose", false,
		"show detailed check-by-check output")
	doctorCmd.Flags().BoolVar(&doctorFix, "fix", false,
		"automatically fix issues where possible")
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

Auto-fix mode:
  --fix       Automatically fix issues where possible (e.g., file permissions)

Exit codes:
  0 - All checks passed (no errors or warnings)
  1 - Warnings present, no errors
  2 - Errors present`,
	Example: `  # Run standard diagnostics
  aix doctor

  # Show all checks including passed ones
  aix doctor --verbose

  # Automatically fix issues
  aix doctor --fix

  # Output as JSON for scripts
  aix doctor --json

See Also: aix status, aix config`,
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

// checkRegistry holds check instances for fixing.
// We need to keep references to run fixes after initial diagnosis.
type checkRegistry struct {
	pathPermissions *doctor.PathPermissionCheck
}

func runDoctor(_ *cobra.Command, _ []string) error {
	runner := doctor.NewRunner()
	registry := &checkRegistry{}

	// Create and register checks, keeping references for fixing
	registry.pathPermissions = doctor.NewPathPermissionCheck()
	runner.AddCheck(registry.pathPermissions)
	runner.AddCheck(doctor.NewPlatformCheck())
	runner.AddCheck(doctor.NewConfigSyntaxCheck())
	runner.AddCheck(doctor.NewConfigSemanticCheck())

	report := runner.Run()

	// If --fix is set and there are fixable issues, attempt fixes
	if doctorFix {
		fixResults := runFixes(registry, report)
		if len(fixResults) > 0 {
			if err := outputFixResults(fixResults); err != nil {
				return err
			}

			// Re-run checks to show final state
			if !doctorQuiet {
				fmt.Println("\nRe-running checks...")
			}

			// Create a fresh runner and registry for re-check
			runner = doctor.NewRunner()
			registry = &checkRegistry{}
			registry.pathPermissions = doctor.NewPathPermissionCheck()
			runner.AddCheck(registry.pathPermissions)
			runner.AddCheck(doctor.NewPlatformCheck())
			runner.AddCheck(doctor.NewConfigSyntaxCheck())
			runner.AddCheck(doctor.NewConfigSemanticCheck())

			report = runner.Run()
		}
	}

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

// runFixes attempts to fix all fixable issues and returns the results.
func runFixes(registry *checkRegistry, report *doctor.DoctorReport) []doctor.FixResult {
	var allResults []doctor.FixResult

	// Check PathPermissionCheck for fixable issues
	if registry.pathPermissions != nil && registry.pathPermissions.CanFix() {
		results := registry.pathPermissions.Fix()
		allResults = append(allResults, results...)
	}

	// Add other fixable checks here as they are implemented
	// For now, only PathPermissionCheck supports fixing

	return allResults
}

// outputFixResults displays the results of fix operations.
func outputFixResults(results []doctor.FixResult) error {
	if doctorQuiet {
		return nil
	}

	if doctorJSON {
		return outputFixResultsJSON(results)
	}

	return outputFixResultsText(results)
}

// outputFixResultsJSON outputs fix results as JSON.
func outputFixResultsJSON(results []doctor.FixResult) error {
	output := struct {
		Fixes []doctor.FixResult `json:"fixes"`
	}{
		Fixes: results,
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(output); err != nil {
		return errors.Wrap(err, "encoding fix results JSON")
	}
	return nil
}

// outputFixResultsText outputs fix results as human-readable text.
func outputFixResultsText(results []doctor.FixResult) error {
	fmt.Println("Fix results:")
	for _, r := range results {
		if r.Fixed {
			fmt.Printf("  ✓ Fixed: %s (%s)\n", r.Path, r.Description)
		} else {
			fmt.Printf("  ✗ Failed: %s - %s\n", r.Path, r.Description)
		}
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
		return errors.Wrap(err, "encoding JSON")
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
			if doctorFix && result.Fixable {
				// Don't show hint for fixable issues when --fix was used (we already fixed them)
				continue
			}
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
