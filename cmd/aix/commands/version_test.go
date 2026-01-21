package commands

import (
	"bytes"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/thoreinstein/aix/internal/paths"
)

// captureStdout captures stdout during function execution.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	// Save original stdout
	oldStdout := os.Stdout

	// Create a pipe
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	// Redirect stdout to write end of pipe
	os.Stdout = w

	// Capture output in goroutine
	var buf bytes.Buffer
	var wg sync.WaitGroup
	wg.Go(func() {
		_, _ = io.Copy(&buf, r)
	})

	// Run the function
	fn()

	// Close write end and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Wait for output capture to complete
	wg.Wait()

	return buf.String()
}

// executeVersionCommand runs the version command and captures its output.
func executeVersionCommand(t *testing.T) string {
	t.Helper()

	var output string
	capturedOutput := captureStdout(t, func() {
		rootCmd.SetArgs([]string{"version"})
		err := rootCmd.Execute()
		if err != nil {
			// Can't use t.Fatalf inside goroutine-adjacent code
			panic("version command failed: " + err.Error())
		}
	})
	output = capturedOutput

	return output
}

func TestPlatformInstalled(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		want     bool
	}{
		{
			name:     "unknown platform returns false",
			platform: "unknown-platform",
			want:     false,
		},
		{
			name:     "empty platform returns false",
			platform: "",
			want:     false,
		},
		{
			name:     "case sensitive platform check",
			platform: "CLAUDE",
			want:     false,
		},
		{
			name:     "numeric platform returns false",
			platform: "12345",
			want:     false,
		},
		{
			name:     "special characters in platform returns false",
			platform: "../../../etc/passwd",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := platformInstalled(tt.platform)
			if got != tt.want {
				t.Errorf("platformInstalled(%q) = %v, want %v", tt.platform, got, tt.want)
			}
		})
	}
}

func TestPlatformInstalled_WithTempDir(t *testing.T) {
	// Set up temp directory as HOME to isolate from real user config
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	// Test 1: No directories exist -> all platforms return false
	for _, platform := range paths.Platforms() {
		if platformInstalled(platform) {
			t.Errorf("platformInstalled(%q) = true, want false (no dirs exist)", platform)
		}
	}

	// Test 2: Create claude config dir -> claude returns true, others false
	claudeDir := paths.GlobalConfigDir("claude")
	if claudeDir == "" {
		t.Fatal("GlobalConfigDir(claude) returned empty string")
	}
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("failed to create claude dir: %v", err)
	}

	if !platformInstalled("claude") {
		t.Error("platformInstalled(claude) = false, want true after creating dir")
	}

	// Other platforms should still be false
	for _, platform := range paths.Platforms() {
		if platform == "claude" {
			continue
		}
		if platformInstalled(platform) {
			t.Errorf("platformInstalled(%q) = true, want false", platform)
		}
	}
}

func TestVersionCommand_OutputFormat(t *testing.T) {
	output := executeVersionCommand(t)

	tests := []struct {
		name     string
		contains string
	}{
		{
			name:     "contains version header",
			contains: "aix version",
		},
		{
			name:     "contains commit field",
			contains: "commit:",
		},
		{
			name:     "contains built field",
			contains: "built:",
		},
		{
			name:     "contains go field",
			contains: "go:",
		},
		{
			name:     "contains platforms section",
			contains: "platforms:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(output, tt.contains) {
				t.Errorf("version output missing %q\nGot:\n%s", tt.contains, output)
			}
		})
	}
}

func TestVersionCommand_GoVersion(t *testing.T) {
	output := executeVersionCommand(t)

	// The output should contain the actual Go runtime version
	goVersion := runtime.Version()
	if !strings.Contains(output, goVersion) {
		t.Errorf("version output should contain Go version %q\nGot:\n%s", goVersion, output)
	}
}

func TestVersionCommand_PlatformsList(t *testing.T) {
	output := executeVersionCommand(t)

	// All supported platforms should be listed
	for _, platform := range paths.Platforms() {
		// Platform name should appear followed by a colon
		expectedPattern := platform + ":"
		if !strings.Contains(output, expectedPattern) {
			t.Errorf("version output should list platform %q\nGot:\n%s", platform, output)
		}
	}
}

func TestVersionCommand_PlatformStatus(t *testing.T) {
	output := executeVersionCommand(t)

	// Each platform should show either "installed" or "not installed"
	for _, platform := range paths.Platforms() {
		t.Run(platform, func(t *testing.T) {
			// Find the line containing this platform
			lines := strings.Split(output, "\n")
			found := false
			for _, line := range lines {
				if strings.Contains(line, platform+":") {
					found = true
					if !strings.Contains(line, "installed") && !strings.Contains(line, "not installed") {
						t.Errorf("platform %q line should contain 'installed' or 'not installed'\nLine: %s", platform, line)
					}
					break
				}
			}
			if !found {
				t.Errorf("platform %q not found in output\n%s", platform, output)
			}
		})
	}
}

func TestVersionCommand_DefaultValues(t *testing.T) {
	// When not set at build time, defaults should be present
	output := executeVersionCommand(t)

	// Check for default values (these are what version.go sets if not overridden)
	tests := []struct {
		name     string
		value    string
		contains string
	}{
		{
			name:     "version shows current value",
			value:    Version,
			contains: "aix version " + Version,
		},
		{
			name:     "commit shows current value",
			value:    Commit,
			contains: "commit:    " + Commit,
		},
		{
			name:     "date shows current value",
			value:    Date,
			contains: "built:     " + Date,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(output, tt.contains) {
				t.Errorf("version output should contain %q\nGot:\n%s", tt.contains, output)
			}
		})
	}
}

func TestVersionCommand_OutputLineCount(t *testing.T) {
	output := executeVersionCommand(t)
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Expected structure:
	// 1: aix version X
	// 2:   commit:    X
	// 3:   built:     X
	// 4:   go:        X
	// 5:   platforms:
	// 6+:   <platform>: <status> (one per platform)
	platformCount := len(paths.Platforms())
	expectedMinLines := 5 + platformCount

	if len(lines) < expectedMinLines {
		t.Errorf("version output has %d lines, expected at least %d\nOutput:\n%s",
			len(lines), expectedMinLines, output)
	}
}

func TestVersionCommand_PlatformAlignment(t *testing.T) {
	output := executeVersionCommand(t)
	lines := strings.Split(output, "\n")

	// Find platform lines and check they're properly formatted
	// Format should be: "    <platform>:   <status>"
	inPlatformsSection := false
	for _, line := range lines {
		if strings.Contains(line, "platforms:") {
			inPlatformsSection = true
			continue
		}
		if inPlatformsSection && strings.TrimSpace(line) != "" {
			// Check line starts with proper indentation (4 spaces)
			if !strings.HasPrefix(line, "    ") {
				t.Errorf("platform line should have 4-space indent: %q", line)
			}
		}
	}
}

// TestVersionCommand_NoError verifies the command completes without error.
func TestVersionCommand_NoError(t *testing.T) {
	// Capture stdout to prevent test output pollution
	_ = captureStdout(t, func() {
		rootCmd.SetArgs([]string{"version"})
		err := rootCmd.Execute()
		if err != nil {
			t.Errorf("version command should not return an error, got: %v", err)
		}
	})
}

// TestVersionCommand_CommandMetadata verifies the command's metadata is set correctly.
func TestVersionCommand_CommandMetadata(t *testing.T) {
	if versionCmd.Use != "version" {
		t.Errorf("versionCmd.Use = %q, want %q", versionCmd.Use, "version")
	}

	if versionCmd.Short == "" {
		t.Error("versionCmd.Short should not be empty")
	}

	if versionCmd.Long == "" {
		t.Error("versionCmd.Long should not be empty")
	}
}
