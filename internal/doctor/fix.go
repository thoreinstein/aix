package doctor

import (
	"fmt"
	"os"

	"github.com/thoreinstein/aix/internal/errors"
)

// Fixer is an optional interface that checks can implement to support auto-remediation.
// Checks that implement Fixer can fix issues they detect when the --fix flag is used.
type Fixer interface {
	// CanFix returns true if this check has fixable issues.
	// Must be called after Run() to check if there are issues that can be fixed.
	CanFix() bool

	// Fix attempts to remediate the issues found by Run().
	// Returns a slice of FixResult indicating what was fixed or why it couldn't be fixed.
	// Must be called after Run().
	Fix() []FixResult
}

// FixResult describes the outcome of an attempted fix operation.
type FixResult struct {
	// Path is the file or directory that was targeted for fixing.
	Path string

	// Fixed indicates whether the fix was successfully applied.
	Fixed bool

	// Description explains what was fixed or why it couldn't be fixed.
	Description string

	// Error contains the error if the fix failed.
	Error error
}

// secureFilePerm is the target permission for config files (rw-r--r--).
const secureFilePerm os.FileMode = 0644

// secureDirPerm is the target permission for config directories (rwxr-xr-x).
const secureDirPerm os.FileMode = 0755

// PermissionFixer fixes file and directory permission issues.
// It is embedded in PathPermissionCheck to provide fix capability.
type PermissionFixer struct {
	issues []pathIssue
}

// CanFix returns true if there are any fixable permission issues.
func (f *PermissionFixer) CanFix() bool {
	for _, issue := range f.issues {
		if issue.Fixable {
			return true
		}
	}
	return false
}

// Fix attempts to fix all fixable permission issues.
// Returns a FixResult for each fixable issue.
func (f *PermissionFixer) Fix() []FixResult {
	// Count fixable issues for pre-allocation
	fixableCount := 0
	for _, issue := range f.issues {
		if issue.Fixable {
			fixableCount++
		}
	}

	results := make([]FixResult, 0, fixableCount)
	for _, issue := range f.issues {
		if !issue.Fixable {
			continue
		}

		result := f.fixIssue(issue)
		results = append(results, result)
	}

	return results
}

// fixIssue attempts to fix a single permission issue.
func (f *PermissionFixer) fixIssue(issue pathIssue) FixResult {
	result := FixResult{
		Path: issue.Path,
	}

	// Determine target permission based on type
	var targetPerm os.FileMode
	switch issue.Type {
	case "file":
		targetPerm = secureFilePerm
	case "directory":
		targetPerm = secureDirPerm
	default:
		result.Description = "unknown type: " + issue.Type
		result.Error = errors.Newf("cannot fix unknown type: %s", issue.Type)
		return result
	}

	// Apply the fix
	if err := os.Chmod(issue.Path, targetPerm); err != nil {
		result.Description = fmt.Sprintf("failed to chmod %04o: %v", targetPerm, err)
		result.Error = errors.Wrapf(err, "chmod %04o %s", targetPerm, issue.Path)
		return result
	}

	result.Fixed = true
	result.Description = fmt.Sprintf("chmod %04o", targetPerm)
	return result
}

// setIssues stores the issues found by the check for later fixing.
// This is called internally by PathPermissionCheck after running.
func (f *PermissionFixer) setIssues(issues []pathIssue) {
	f.issues = issues
}

// CountFixable returns the number of fixable issues.
func (f *PermissionFixer) CountFixable() int {
	count := 0
	for _, issue := range f.issues {
		if issue.Fixable {
			count++
		}
	}
	return count
}
