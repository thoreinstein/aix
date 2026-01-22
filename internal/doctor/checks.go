package doctor

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/thoreinstein/aix/internal/paths"
	"github.com/thoreinstein/aix/internal/platform"
)

// maxSecureFilePerm is the maximum secure permission for config files (-rw-r--r--).
const maxSecureFilePerm os.FileMode = 0644

// PathPermissionCheck validates file paths and permissions for platform configurations.
type PathPermissionCheck struct{}

var _ Check = (*PathPermissionCheck)(nil)

// NewPathPermissionCheck creates a new path permission check.
func NewPathPermissionCheck() *PathPermissionCheck {
	return &PathPermissionCheck{}
}

// Name returns the unique identifier for this check.
func (c *PathPermissionCheck) Name() string {
	return "path-permissions"
}

// Category returns the grouping for this check.
func (c *PathPermissionCheck) Category() string {
	return "filesystem"
}

// Run executes the path and permission diagnostic check.
func (c *PathPermissionCheck) Run() *CheckResult {
	results := platform.DetectAll()

	var issues []pathIssue
	var checked int

	for _, p := range results {
		// Check config directory
		if p.GlobalConfig != "" {
			dirIssues := c.checkDirectory(p.GlobalConfig, p.Name)
			issues = append(issues, dirIssues...)
			checked++
		}

		// Check MCP config file
		if p.MCPConfig != "" {
			fileIssues := c.checkFile(p.MCPConfig, p.Name)
			issues = append(issues, fileIssues...)
			checked++
		}
	}

	return c.buildResult(issues, checked)
}

// pathIssue represents a single path or permission problem.
type pathIssue struct {
	Path        string
	Platform    string
	Type        string // "file" or "directory"
	Problem     string
	Severity    Severity
	Permissions string // octal representation if available
	Fixable     bool
	FixHint     string
}

// checkFile validates a config file path and permissions.
func (c *PathPermissionCheck) checkFile(path, platformName string) []pathIssue {
	var issues []pathIssue

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		// File doesn't exist is not an error - platform may not be configured
		return nil
	}
	if err != nil {
		issues = append(issues, pathIssue{
			Path:     path,
			Platform: platformName,
			Type:     "file",
			Problem:  fmt.Sprintf("cannot stat file: %v", err),
			Severity: SeverityError,
		})
		return issues
	}

	// Check if file is readable
	f, err := os.Open(path)
	if err != nil {
		issues = append(issues, pathIssue{
			Path:        path,
			Platform:    platformName,
			Type:        "file",
			Problem:     "file is not readable",
			Severity:    SeverityError,
			Permissions: formatPermissions(info.Mode()),
			FixHint:     "chmod 644 " + path,
		})
		return issues
	}
	f.Close()

	// Check permissions (skip on Windows where Unix permissions don't apply)
	if runtime.GOOS != "windows" {
		permIssues := c.checkFilePermissions(path, platformName, info.Mode())
		issues = append(issues, permIssues...)
	}

	return issues
}

// checkDirectory validates a config directory path and permissions.
func (c *PathPermissionCheck) checkDirectory(path, platformName string) []pathIssue {
	var issues []pathIssue

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		// Directory doesn't exist is not an error - platform may not be installed
		return nil
	}
	if err != nil {
		issues = append(issues, pathIssue{
			Path:     path,
			Platform: platformName,
			Type:     "directory",
			Problem:  fmt.Sprintf("cannot stat directory: %v", err),
			Severity: SeverityError,
		})
		return issues
	}

	if !info.IsDir() {
		issues = append(issues, pathIssue{
			Path:     path,
			Platform: platformName,
			Type:     "directory",
			Problem:  "expected directory but found file",
			Severity: SeverityError,
		})
		return issues
	}

	// Check if directory is writable by creating a temp file
	writable, err := c.isDirectoryWritable(path)
	if err != nil || !writable {
		issues = append(issues, pathIssue{
			Path:        path,
			Platform:    platformName,
			Type:        "directory",
			Problem:     "directory is not writable",
			Severity:    SeverityWarning,
			Permissions: formatPermissions(info.Mode()),
			FixHint:     "chmod u+w " + path,
		})
	}

	// Check permissions (skip on Windows where Unix permissions don't apply)
	if runtime.GOOS != "windows" {
		permIssues := c.checkDirectoryPermissions(path, platformName, info.Mode())
		issues = append(issues, permIssues...)
	}

	return issues
}

// checkFilePermissions validates file permissions for security concerns.
func (c *PathPermissionCheck) checkFilePermissions(path, platformName string, mode os.FileMode) []pathIssue {
	var issues []pathIssue
	perm := mode.Perm()

	// World-writable is always a security concern
	if perm&0002 != 0 {
		issues = append(issues, pathIssue{
			Path:        path,
			Platform:    platformName,
			Type:        "file",
			Problem:     "file is world-writable (security risk)",
			Severity:    SeverityWarning,
			Permissions: formatPermissions(mode),
			Fixable:     true,
			FixHint:     "chmod 644 " + path,
		})
	}

	// Config files that may contain secrets should not be world-readable
	// Check if permissions are more permissive than 0644
	if perm > maxSecureFilePerm && c.mayContainSecrets(path) {
		issues = append(issues, pathIssue{
			Path:        path,
			Platform:    platformName,
			Type:        "file",
			Problem:     fmt.Sprintf("file has overly permissive permissions (mode %s, expected %s or less)", formatPermissions(mode), formatOctal(maxSecureFilePerm)),
			Severity:    SeverityWarning,
			Permissions: formatPermissions(mode),
			Fixable:     true,
			FixHint:     "chmod 644 " + path,
		})
	}

	return issues
}

// checkDirectoryPermissions validates directory permissions for security concerns.
func (c *PathPermissionCheck) checkDirectoryPermissions(path, platformName string, mode os.FileMode) []pathIssue {
	var issues []pathIssue
	perm := mode.Perm()

	// World-writable directories are always a security concern
	if perm&0002 != 0 {
		issues = append(issues, pathIssue{
			Path:        path,
			Platform:    platformName,
			Type:        "directory",
			Problem:     "directory is world-writable (security risk)",
			Severity:    SeverityWarning,
			Permissions: formatPermissions(mode),
			Fixable:     true,
			FixHint:     "chmod 755 " + path,
		})
	}

	return issues
}

// isDirectoryWritable tests if a directory is writable by creating a temp file.
func (c *PathPermissionCheck) isDirectoryWritable(path string) (bool, error) {
	tmpFile, err := os.CreateTemp(path, ".aix-doctor-test-*")
	if err != nil {
		return false, err
	}

	// Clean up the test file
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	os.Remove(tmpPath)

	return true, nil
}

// mayContainSecrets returns true if the file path suggests it may contain secrets.
// This is a heuristic based on common config file patterns.
func (c *PathPermissionCheck) mayContainSecrets(path string) bool {
	base := filepath.Base(path)
	lower := strings.ToLower(base)

	// Common config files that may contain API keys or tokens
	secretPatterns := []string{
		".json",    // JSON configs often contain env vars with secrets
		".toml",    // TOML configs (Gemini)
		"config",   // Generic config files
		"mcp",      // MCP configs may have auth headers
		"claude",   // Claude config
		"opencode", // OpenCode config
	}

	for _, pattern := range secretPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	return false
}

// buildResult constructs the final CheckResult from accumulated issues.
func (c *PathPermissionCheck) buildResult(issues []pathIssue, checked int) *CheckResult {
	if len(issues) == 0 {
		return &CheckResult{
			Name:     c.Name(),
			Category: c.Category(),
			Status:   SeverityPass,
			Message:  fmt.Sprintf("all %d paths have valid permissions", checked),
		}
	}

	// Find the highest severity among all issues
	highestSeverity := SeverityPass
	var hasError, hasWarning bool
	for _, issue := range issues {
		if issue.Severity == SeverityError {
			hasError = true
		}
		if issue.Severity == SeverityWarning {
			hasWarning = true
		}
	}

	if hasError {
		highestSeverity = SeverityError
	} else if hasWarning {
		highestSeverity = SeverityWarning
	}

	// Build details map
	details := make(map[string]any)
	details["checked_paths"] = checked
	details["issue_count"] = len(issues)

	// Convert issues to a slice of maps for JSON serialization
	issueDetails := make([]map[string]any, 0, len(issues))
	for _, issue := range issues {
		issueMap := map[string]any{
			"path":     issue.Path,
			"platform": issue.Platform,
			"type":     issue.Type,
			"problem":  issue.Problem,
			"severity": issue.Severity.String(),
		}
		if issue.Permissions != "" {
			issueMap["permissions"] = issue.Permissions
		}
		if issue.FixHint != "" {
			issueMap["fix_hint"] = issue.FixHint
		}
		issueDetails = append(issueDetails, issueMap)
	}
	details["issues"] = issueDetails

	// Check if any issues are fixable
	fixable := false
	var fixHints []string
	for _, issue := range issues {
		if issue.Fixable {
			fixable = true
			if issue.FixHint != "" {
				fixHints = append(fixHints, issue.FixHint)
			}
		}
	}

	message := fmt.Sprintf("found %d permission issue(s) across %d paths", len(issues), checked)

	result := &CheckResult{
		Name:     c.Name(),
		Category: c.Category(),
		Status:   highestSeverity,
		Message:  message,
		Details:  details,
		Fixable:  fixable,
	}

	if len(fixHints) > 0 {
		result.FixHint = strings.Join(fixHints, "; ")
	}

	return result
}

// formatPermissions returns a human-readable permission string (e.g., "0644").
func formatPermissions(mode os.FileMode) string {
	return fmt.Sprintf("%04o", mode.Perm())
}

// formatOctal returns the octal representation of a file mode.
func formatOctal(mode os.FileMode) string {
	return fmt.Sprintf("%04o", mode)
}

// ConfigSyntaxCheck validates configuration file syntax (JSON/TOML parsing).
type ConfigSyntaxCheck struct{}

var _ Check = (*ConfigSyntaxCheck)(nil)

// NewConfigSyntaxCheck creates a new ConfigSyntaxCheck instance.
func NewConfigSyntaxCheck() *ConfigSyntaxCheck {
	return &ConfigSyntaxCheck{}
}

// Name returns the unique identifier for this check.
func (c *ConfigSyntaxCheck) Name() string {
	return "config-syntax"
}

// Category returns the grouping for this check.
func (c *ConfigSyntaxCheck) Category() string {
	return "config"
}

// syntaxFileResult represents the validation result for a single file.
type syntaxFileResult struct {
	Path    string `json:"path"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// Run executes the syntax validation check across all installed platforms.
func (c *ConfigSyntaxCheck) Run() *CheckResult {
	result := &CheckResult{
		Name:     c.Name(),
		Category: c.Category(),
		Status:   SeverityPass,
		Details:  make(map[string]any),
	}

	installed := platform.DetectInstalled()
	if len(installed) == 0 {
		result.Status = SeverityInfo
		result.Message = "no platforms installed"
		return result
	}

	var fileResults []syntaxFileResult
	var errorCount, passCount, infoCount int

	for _, p := range installed {
		// Check MCP config file
		mcpPath := p.MCPConfig
		if mcpPath != "" {
			fr := c.validateFile(mcpPath)
			fileResults = append(fileResults, fr)
			switch fr.Status {
			case "pass":
				passCount++
			case "error":
				errorCount++
			case "info":
				infoCount++
			}
		}

		// Check global config file if different from MCP config
		globalConfigPath := c.getGlobalConfigPath(p.Name)
		if globalConfigPath != "" && globalConfigPath != mcpPath {
			fr := c.validateFile(globalConfigPath)
			fileResults = append(fileResults, fr)
			switch fr.Status {
			case "pass":
				passCount++
			case "error":
				errorCount++
			case "info":
				infoCount++
			}
		}
	}

	result.Details["files"] = fileResults
	result.Details["checked"] = len(fileResults)
	result.Details["passed"] = passCount
	result.Details["errors"] = errorCount
	result.Details["missing"] = infoCount

	// Determine overall status
	switch {
	case errorCount > 0:
		result.Status = SeverityError
		result.Message = fmt.Sprintf("%d config file(s) have syntax errors", errorCount)
		result.Fixable = false
		result.FixHint = "review the error details and fix the syntax in each file"
	case passCount > 0:
		result.Status = SeverityPass
		result.Message = fmt.Sprintf("%d config file(s) validated successfully", passCount)
	default:
		result.Status = SeverityInfo
		result.Message = "no config files found to validate"
	}

	return result
}

// getGlobalConfigPath returns the main global config path for a platform.
// This may differ from the MCP config path for some platforms.
func (c *ConfigSyntaxCheck) getGlobalConfigPath(platformName string) string {
	globalDir := paths.GlobalConfigDir(platformName)
	if globalDir == "" {
		return ""
	}

	switch platformName {
	case paths.PlatformClaude:
		// Claude's main settings are in ~/.claude/settings.json
		return filepath.Join(globalDir, "settings.json")
	case paths.PlatformOpenCode:
		// OpenCode's main config is config.toml
		return filepath.Join(globalDir, "config.toml")
	case paths.PlatformCodex:
		// Codex settings file
		return filepath.Join(globalDir, "settings.json")
	case paths.PlatformGemini:
		// Gemini uses settings.toml (which is also MCP config)
		return filepath.Join(globalDir, "settings.toml")
	default:
		return ""
	}
}

// validateFile checks if a file is syntactically valid.
func (c *ConfigSyntaxCheck) validateFile(filePath string) syntaxFileResult {
	fr := syntaxFileResult{Path: filePath}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fr.Status = "info"
			fr.Message = "file does not exist (not configured)"
			return fr
		}
		if errors.Is(err, os.ErrPermission) {
			fr.Status = "error"
			fr.Message = fmt.Sprintf("permission denied: %v", err)
			return fr
		}
		fr.Status = "error"
		fr.Message = fmt.Sprintf("read error: %v", err)
		return fr
	}

	// Empty files are valid (no content to parse)
	if len(data) == 0 {
		fr.Status = "pass"
		fr.Message = "empty file"
		return fr
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".json":
		fr = c.validateJSON(data, fr)
	case ".toml":
		fr = c.validateTOML(data, fr)
	default:
		// For unknown extensions, try JSON first (more common), then TOML
		fr = c.validateJSON(data, fr)
		if fr.Status == "error" {
			// Try TOML as fallback
			tomlResult := c.validateTOML(data, syntaxFileResult{Path: filePath})
			if tomlResult.Status == "pass" {
				fr = tomlResult
			}
		}
	}

	return fr
}

// validateJSON validates JSON syntax and returns position info on errors.
func (c *ConfigSyntaxCheck) validateJSON(data []byte, fr syntaxFileResult) syntaxFileResult {
	var v any
	err := json.Unmarshal(data, &v)
	if err != nil {
		fr.Status = "error"
		fr.Message = formatJSONError(err, data)
		return fr
	}
	fr.Status = "pass"
	return fr
}

// validateTOML validates TOML syntax and returns position info on errors.
func (c *ConfigSyntaxCheck) validateTOML(data []byte, fr syntaxFileResult) syntaxFileResult {
	var v any
	err := toml.Unmarshal(data, &v)
	if err != nil {
		fr.Status = "error"
		fr.Message = formatTOMLError(err)
		return fr
	}
	fr.Status = "pass"
	return fr
}

// formatJSONError extracts position information from JSON syntax errors.
func formatJSONError(err error, data []byte) string {
	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		line, col := offsetToLineCol(data, int(syntaxErr.Offset))
		return fmt.Sprintf("JSON syntax error at line %d, column %d: %s", line, col, syntaxErr.Error())
	}

	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		line, col := offsetToLineCol(data, int(typeErr.Offset))
		return fmt.Sprintf("JSON type error at line %d, column %d: %s", line, col, typeErr.Error())
	}

	return fmt.Sprintf("JSON error: %v", err)
}

// formatTOMLError extracts position information from TOML decode errors.
func formatTOMLError(err error) string {
	// go-toml/v2 DecodeError includes line/column via Position() method
	var decodeErr *toml.DecodeError
	if errors.As(err, &decodeErr) {
		row, col := decodeErr.Position()
		return fmt.Sprintf("TOML syntax error at line %d, column %d: %s",
			row, col, decodeErr.Error())
	}

	return fmt.Sprintf("TOML error: %v", err)
}

// offsetToLineCol converts a byte offset to line and column numbers.
// Lines and columns are 1-indexed.
func offsetToLineCol(data []byte, offset int) (line, col int) {
	if offset > len(data) {
		offset = len(data)
	}
	if offset < 0 {
		offset = 0
	}

	line = 1
	lineStart := 0

	for i := range offset {
		if data[i] == '\n' {
			line++
			lineStart = i + 1
		}
	}

	col = offset - lineStart + 1
	return line, col
}
