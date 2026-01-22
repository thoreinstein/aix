package doctor

import "time"

// Check is the interface that diagnostic checks must implement.
type Check interface {
	// Name returns the unique identifier for this check.
	Name() string

	// Category returns the grouping for this check (e.g., "platform", "config").
	Category() string

	// Run executes the diagnostic check and returns its result.
	Run() *CheckResult
}

// Runner executes diagnostic checks and aggregates their results.
type Runner struct {
	checks []Check
}

// NewRunner creates a new diagnostic runner.
func NewRunner() *Runner {
	return &Runner{
		checks: make([]Check, 0),
	}
}

// AddCheck registers a diagnostic check with the runner.
func (r *Runner) AddCheck(c Check) {
	r.checks = append(r.checks, c)
}

// Run executes all registered checks and returns a report.
func (r *Runner) Run() *DoctorReport {
	report := &DoctorReport{
		Timestamp: time.Now().UTC(),
		Results:   make([]*CheckResult, 0, len(r.checks)),
	}

	for _, check := range r.checks {
		result := check.Run()
		report.Results = append(report.Results, result)

		// Update summary counts
		switch result.Status {
		case SeverityPass:
			report.Summary.Passed++
		case SeverityInfo:
			report.Summary.Info++
		case SeverityWarning:
			report.Summary.Warnings++
		case SeverityError:
			report.Summary.Errors++
		}
	}

	return report
}

// DoctorReport aggregates all check results with timing and summary.
type DoctorReport struct {
	// Timestamp is when the diagnostic run started.
	Timestamp time.Time `json:"timestamp"`

	// Results contains the outcome of each check.
	Results []*CheckResult `json:"results"`

	// Summary contains counts by severity level.
	Summary Summary `json:"summary"`
}

// HasErrors returns true if any check has SeverityError.
func (r *DoctorReport) HasErrors() bool {
	return r.Summary.Errors > 0
}

// HasWarnings returns true if any check has SeverityWarning.
func (r *DoctorReport) HasWarnings() bool {
	return r.Summary.Warnings > 0
}
