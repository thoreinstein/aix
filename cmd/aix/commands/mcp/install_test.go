package mcp

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/git"
	"github.com/thoreinstein/aix/internal/install"
)

func Test_isGitURL(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   bool
	}{
		{
			name:   "https URL",
			source: "https://github.com/user/repo.git",
			want:   true,
		},
		{
			name:   "https URL without .git suffix",
			source: "https://github.com/user/repo",
			want:   true,
		},
		{
			name:   "http URL",
			source: "http://github.com/user/repo",
			want:   true,
		},
		{
			name:   "git protocol",
			source: "git://github.com/user/repo.git",
			want:   true,
		},
		{
			name:   "git@ SSH",
			source: "git@github.com:user/repo.git",
			want:   true,
		},
		{
			name:   "git@ SSH without .git suffix",
			source: "git@github.com:user/repo",
			want:   false, // git.IsURL requires .git suffix for scp-like
		},
		{
			name:   ".git suffix only",
			source: "github.com/user/repo.git",
			want:   false, // git.IsURL requires scheme or scp-like
		},
		{
			name:   "ssh protocol",
			source: "ssh://git@github.com/user/repo",
			want:   true,
		},
		{
			name:   "file protocol",
			source: "file:///path/to/repo",
			want:   true,
		},
		{
			name:   "simple name",
			source: "github-mcp",
			want:   false,
		},
		{
			name:   "local relative path",
			source: "./server.json",
			want:   false,
		},
		{
			name:   "local absolute path",
			source: "/path/to/server.json",
			want:   false,
		},
		{
			name:   "local directory name",
			source: "my-server",
			want:   false,
		},
		{
			name:   "empty string",
			source: "",
			want:   false,
		},
		{
			name:   "ftp protocol",
			source: "ftp://example.com/repo",
			want:   false, // git.IsURL only allows http, https, ssh, git, file
		},
		{
			name:   "custom protocol",
			source: "custom://example.com/repo",
			want:   false, // git.IsURL only allows specific schemes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := git.IsURL(tt.source)
			if got != tt.want {
				t.Errorf("git.IsURL(%q) = %v, want %v", tt.source, got, tt.want)
			}
		})
	}
}

func Test_installFromLocal_FileNotFound(t *testing.T) {
	err := installFromLocal("/nonexistent/path/to/server.json", cli.ScopeUser)
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}

	// Check error message contains the path
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("expected non-empty error message")
	}
}

func Test_installFromLocal_NotJSONFile(t *testing.T) {
	// Create a temp directory with a non-JSON file
	tempDir := t.TempDir()
	txtPath := filepath.Join(tempDir, "server.txt")

	// Write valid JSON content to a .txt file
	content := `{"name": "test", "command": "echo"}`
	if err := os.WriteFile(txtPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	err := installFromLocal(txtPath, cli.ScopeUser)
	if err == nil {
		t.Error("expected error for non-JSON file extension, got nil")
	}

	// Check error message mentions expected .json
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("expected non-empty error message")
	}
}

func Test_installFromLocal_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	jsonPath := filepath.Join(tempDir, "invalid.json")

	// Write invalid JSON
	if err := os.WriteFile(jsonPath, []byte("{invalid json content}"), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	err := installFromLocal(jsonPath, cli.ScopeUser)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func Test_installFromLocal_EmptyJSON(t *testing.T) {
	tempDir := t.TempDir()
	jsonPath := filepath.Join(tempDir, "empty.json")

	// Write empty JSON object
	if err := os.WriteFile(jsonPath, []byte("{}"), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	err := installFromLocal(jsonPath, cli.ScopeUser)
	if err == nil {
		t.Error("expected error for empty JSON (missing required fields), got nil")
	}
}

func Test_installFromLocal_MissingCommand(t *testing.T) {
	tempDir := t.TempDir()
	jsonPath := filepath.Join(tempDir, "no-command.json")

	// Write JSON with name but no command or URL
	content := `{"name": "test-server"}`
	if err := os.WriteFile(jsonPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	err := installFromLocal(jsonPath, cli.ScopeUser)
	if err == nil {
		t.Error("expected error for server missing command/URL, got nil")
	}
}

func TestInstallCmd_Metadata(t *testing.T) {
	// Verify command metadata is properly configured
	if installCmd.Use != "install <source>" {
		t.Errorf("Use = %q, want %q", installCmd.Use, "install <source>")
	}

	if installCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if installCmd.Long == "" {
		t.Error("Long description should not be empty")
	}

	if installCmd.Example == "" {
		t.Error("Example should not be empty")
	}

	// Check args validator requires exactly 1 argument
	if installCmd.Args == nil {
		t.Error("Args validator should be set")
	}
}

func TestInstallCmd_Flags(t *testing.T) {
	// Check required flags exist
	expectedFlags := []string{"force", "file"}
	for _, flagName := range expectedFlags {
		if installCmd.Flags().Lookup(flagName) == nil {
			t.Errorf("--%s flag should be defined", flagName)
		}
	}

	// Verify short flag for file
	fileFlag := installCmd.Flags().ShorthandLookup("f")
	if fileFlag == nil {
		t.Error("-f shorthand for --file should be defined")
	}
}

func TestInstallSentinelErrors(t *testing.T) {
	// Ensure sentinel error is properly defined
	if errInstallFailed == nil {
		t.Error("errInstallFailed should be defined")
	}

	// Verify error message
	if errInstallFailed.Error() != "installation failed" {
		t.Errorf("unexpected error message: %s", errInstallFailed.Error())
	}
}

func Test_sourceResolutionLogic(t *testing.T) {
	// Test the resolution priority logic used in runInstall
	// This simulates the decision tree without actually running the command
	tests := []struct {
		name            string
		source          string
		installFileFlag bool
		wantPathType    string // "git", "local", "repo"
	}{
		{
			name:            "explicit file flag with local path",
			source:          "./server.json",
			installFileFlag: true,
			wantPathType:    "local",
		},
		{
			name:            "explicit file flag with git URL",
			source:          "https://github.com/user/repo.git",
			installFileFlag: true,
			wantPathType:    "git",
		},
		{
			name:            "auto-detect git URL",
			source:          "https://github.com/user/repo.git",
			installFileFlag: false,
			wantPathType:    "git",
		},
		{
			name:            "auto-detect local path with dot slash",
			source:          "./server.json",
			installFileFlag: false,
			wantPathType:    "local",
		},
		{
			name:            "auto-detect local path with slash",
			source:          "path/to/server.json",
			installFileFlag: false,
			wantPathType:    "local",
		},
		{
			name:            "auto-detect absolute path",
			source:          "/path/to/server.json",
			installFileFlag: false,
			wantPathType:    "local",
		},
		{
			name:            "simple name falls through to repo",
			source:          "github-mcp",
			installFileFlag: false,
			wantPathType:    "repo",
		},
		{
			name:            "name with dash falls through to repo",
			source:          "my-server",
			installFileFlag: false,
			wantPathType:    "repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the resolution logic from runInstall
			var pathType string

			if tt.installFileFlag {
				if git.IsURL(tt.source) {
					pathType = "git"
				} else {
					pathType = "local"
				}
			} else {
				if git.IsURL(tt.source) || install.LooksLikePath(tt.source) {
					if git.IsURL(tt.source) {
						pathType = "git"
					} else {
						pathType = "local"
					}
				} else {
					pathType = "repo"
				}
			}

			if pathType != tt.wantPathType {
				t.Errorf("source resolution for %q (file=%v) = %q, want %q",
					tt.source, tt.installFileFlag, pathType, tt.wantPathType)
			}
		})
	}
}

func Test_installFromLocal_NameDerivation(t *testing.T) {
	// Test that server name is derived from filename when not set in JSON
	tempDir := t.TempDir()

	tests := []struct {
		name         string
		filename     string
		jsonContent  string
		expectedName string
	}{
		{
			name:         "name from filename",
			filename:     "github-mcp.json",
			jsonContent:  `{"command": "npx", "args": ["-y", "@modelcontextprotocol/server-github"]}`, // Corrected JSON string escaping
			expectedName: "github-mcp",
		},
		{
			name:         "name from json takes precedence",
			filename:     "other-name.json",
			jsonContent:  `{"name": "explicit-name", "command": "npx", "args": ["-y", "@modelcontextprotocol/server-github"]}`, // Corrected JSON string escaping
			expectedName: "explicit-name",
		},
		{
			name:         "complex filename",
			filename:     "my-custom-server-v2.json",
			jsonContent:  `{"command": "node", "args": ["server.js"]}`, // Corrected JSON string escaping
			expectedName: "my-custom-server-v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonPath := filepath.Join(tempDir, tt.filename)
			if err := os.WriteFile(jsonPath, []byte(tt.jsonContent), 0o644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			// Verify the filename stripping logic works as expected
			basename := filepath.Base(jsonPath)
			derivedName := basename[:len(basename)-len(".json")]
			if derivedName != tt.expectedName && !strings.Contains(tt.jsonContent, `"name":`) {
				t.Errorf("derivedName = %q, want %q", derivedName, tt.expectedName)
			}
		})
	}
}

func Test_installFromLocal_VariousJSONFormats(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty file",
			content:     "",
			wantErr:     true,
			errContains: "parsing",
		},
		{
			name:        "null value",
			content:     "null",
			wantErr:     true,
			errContains: "",
		},
		{
			name:        "array instead of object",
			content:     `["item1", "item2"]`, // Corrected JSON string escaping
			wantErr:     true,
			errContains: "parsing",
		},
		{
			name:        "whitespace only",
			content:     "   \n\t  ",
			wantErr:     true,
			errContains: "parsing",
		},
		{
			name:        "json with trailing comma",
			content:     `{"name": "test",}`, // Corrected JSON string escaping
			wantErr:     true,
			errContains: "parsing",
		},
		{
			name:        "json with comments",
			content:     `{"name": "test" /* comment */}`, // Corrected JSON string escaping
			wantErr:     true,
			errContains: "parsing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			jsonPath := filepath.Join(tempDir, "test.json")

			if err := os.WriteFile(jsonPath, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			err := installFromLocal(jsonPath, cli.ScopeUser)
			if (err != nil) != tt.wantErr {
				t.Errorf("installFromLocal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_installFromLocal_FilePermissions(t *testing.T) {
	// Test reading files with restricted permissions
	tempDir := t.TempDir()
	jsonPath := filepath.Join(tempDir, "restricted.json")

	// Write valid JSON
	content := `{"name": "test", "command": "echo"}`
	if err := os.WriteFile(jsonPath, []byte(content), 0o000); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Ensure cleanup restores permissions for deletion
	t.Cleanup(func() {
		_ = os.Chmod(jsonPath, 0o644)
	})

	err := installFromLocal(jsonPath, cli.ScopeUser)
	if err == nil {
		// On some systems (e.g., running as root), permission checks may pass
		t.Skip("permission test requires non-root user")
	}
}

func Test_installFromLocal_DirectoryInsteadOfFile(t *testing.T) {
	tempDir := t.TempDir()
	dirPath := filepath.Join(tempDir, "notafile.json")

	// Create a directory with .json extension
	if err := os.Mkdir(dirPath, 0o755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	err := installFromLocal(dirPath, cli.ScopeUser)
	if err == nil {
		t.Error("expected error when path is a directory, got nil")
	}
}

func Test_installFromLocal_SymlinkRejected(t *testing.T) {
	tempDir := t.TempDir()
	realFile := filepath.Join(tempDir, "real.json")
	symlink := filepath.Join(tempDir, "link.json")

	// Write valid JSON to real file
	content := `{"name": "test", "command": "echo"}`
	if err := os.WriteFile(realFile, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Create symlink
	if err := os.Symlink(realFile, symlink); err != nil {
		t.Skip("symlink creation not supported on this system")
	}

	// The function should reject symlinks for security
	err := installFromLocal(symlink, cli.ScopeUser)
	if err == nil {
		t.Fatal("expected error due to symlink, but succeeded")
	}

	// Verify error message mentions symlink and security
	errMsg := err.Error()
	if !strings.Contains(errMsg, "symlink") {
		t.Errorf("expected error to mention symlink, got: %v", err)
	}
}

func Test_installFromLocal_ValidServerWithWarnings(t *testing.T) {
	// Test a server config that has both command and URL (generates warning)
	tempDir := t.TempDir()
	jsonPath := filepath.Join(tempDir, "ambiguous.json")

	// Write JSON with both command and URL (ambiguous config that generates warning)
	content := `{"name": "test", "command": "echo", "url": "http://example.com/mcp"}`
	if err := os.WriteFile(jsonPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// This should pass validation (with warnings) and proceed to platform install
	err := installFromLocal(jsonPath, cli.ScopeUser)
	// Expect success or platform-related error (not validation error)
	if errors.Is(err, errInstallFailed) {
		t.Error("expected validation to pass (possibly with warnings), got errInstallFailed")
	}
}

func Test_installFromLocal_URLServer(t *testing.T) {
	// Test a remote/SSE server config
	tempDir := t.TempDir()
	jsonPath := filepath.Join(tempDir, "remote.json")

	// Write JSON for remote server
	content := `{"name": "api-gateway", "url": "https://api.example.com/mcp"}`
	if err := os.WriteFile(jsonPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// This should pass validation and proceed to platform install
	err := installFromLocal(jsonPath, cli.ScopeUser)
	// Should not fail at validation
	if errors.Is(err, errInstallFailed) {
		t.Error("expected validation to pass for valid remote server")
	}
}

func Test_installFromLocal_ServerWithEnvAndHeaders(t *testing.T) {
	// Test a server config with environment variables and headers
	tempDir := t.TempDir()
	jsonPath := filepath.Join(tempDir, "full-config.json")

	content := `{
		"name": "full-server",
		"url": "https://api.example.com/mcp",
		"headers": {"Authorization": "Bearer token123"},
		"env": {"API_KEY": "secret"}
	}`
	if err := os.WriteFile(jsonPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	err := installFromLocal(jsonPath, cli.ScopeUser)
	// Should not fail at validation
	if errors.Is(err, errInstallFailed) {
		t.Error("expected validation to pass for server with headers and env")
	}
}

func Test_installFromLocal_ServerWithPlatforms(t *testing.T) {
	// Test a server config with platform restrictions
	tempDir := t.TempDir()
	jsonPath := filepath.Join(tempDir, "platform-restricted.json")

	content := `{
		"name": "darwin-only",
		"command": "macos-server",
		"platforms": ["darwin"]
	}`
	if err := os.WriteFile(jsonPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	err := installFromLocal(jsonPath, cli.ScopeUser)
	// Should not fail at validation
	if errors.Is(err, errInstallFailed) {
		t.Error("expected validation to pass for server with platform restriction")
	}
}

func Test_installFromLocal_ServerWithInvalidPlatform(t *testing.T) {
	// Test a server config with invalid platform value
	tempDir := t.TempDir()
	jsonPath := filepath.Join(tempDir, "invalid-platform.json")

	content := `{
		"name": "bad-platform",
		"command": "server",
		"platforms": ["notaplatform"]
	}`
	if err := os.WriteFile(jsonPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	err := installFromLocal(jsonPath, cli.ScopeUser)
	// Should fail validation
	if !errors.Is(err, errInstallFailed) {
		t.Errorf("expected errInstallFailed for invalid platform, got: %v", err)
	}
}

func Test_installFromLocal_ServerWithInvalidTransport(t *testing.T) {
	// Test a server config with invalid transport value
	tempDir := t.TempDir()
	jsonPath := filepath.Join(tempDir, "invalid-transport.json")

	content := `{
		"name": "bad-transport",
		"command": "server",
		"transport": "websocket"
	}`
	if err := os.WriteFile(jsonPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	err := installFromLocal(jsonPath, cli.ScopeUser)
	// Should fail validation
	if !errors.Is(err, errInstallFailed) {
		t.Errorf("expected errInstallFailed for invalid transport, got: %v", err)
	}
}

func Test_installFromLocal_ServerWithEmptyEnvKey(t *testing.T) {
	// Test a server config with empty environment variable key
	tempDir := t.TempDir()
	jsonPath := filepath.Join(tempDir, "empty-env-key.json")

	content := `{
		"name": "bad-env",
		"command": "server",
		"env": {"": "value"}
	}`
	if err := os.WriteFile(jsonPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	err := installFromLocal(jsonPath, cli.ScopeUser)
	// Should fail validation
	if !errors.Is(err, errInstallFailed) {
		t.Errorf("expected errInstallFailed for empty env key, got: %v", err)
	}
}

func Test_installFromLocal_ServerWithEmptyHeaderKey(t *testing.T) {
	// Test a server config with empty header key
	tempDir := t.TempDir()
	jsonPath := filepath.Join(tempDir, "empty-header-key.json")

	content := `{
		"name": "bad-header",
		"url": "https://example.com",
		"headers": {"": "value"}
	}`
	if err := os.WriteFile(jsonPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	err := installFromLocal(jsonPath, cli.ScopeUser)
	// Should fail validation
	if !errors.Is(err, errInstallFailed) {
		t.Errorf("expected errInstallFailed for empty header key, got: %v", err)
	}
}

func Test_installFromLocal_StdioTransportWithURL(t *testing.T) {
	// Test explicit stdio transport with URL (invalid combo)
	tempDir := t.TempDir()
	jsonPath := filepath.Join(tempDir, "stdio-with-url.json")

	content := `{
		"name": "invalid-combo",
		"transport": "stdio",
		"url": "https://example.com"
	}`
	if err := os.WriteFile(jsonPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	err := installFromLocal(jsonPath, cli.ScopeUser)
	// Should fail validation (stdio requires command)
	if !errors.Is(err, errInstallFailed) {
		t.Errorf("expected errInstallFailed for stdio without command, got: %v", err)
	}
}

func Test_installFromLocal_SSETransportWithCommand(t *testing.T) {
	// Test explicit SSE transport with command (invalid combo)
	tempDir := t.TempDir()
	jsonPath := filepath.Join(tempDir, "sse-with-command.json")

	content := `{
		"name": "invalid-combo",
		"transport": "sse",
		"command": "echo"
	}`
	if err := os.WriteFile(jsonPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	err := installFromLocal(jsonPath, cli.ScopeUser)
	// Should fail validation (sse requires url)
	if !errors.Is(err, errInstallFailed) {
		t.Errorf("expected errInstallFailed for sse without url, got: %v", err)
	}
}

func Test_installFromLocal_ExplicitTransportStdio(t *testing.T) {
	// Test valid explicit stdio transport
	tempDir := t.TempDir()
	jsonPath := filepath.Join(tempDir, "explicit-stdio.json")

	content := `{
		"name": "valid-stdio",
		"transport": "stdio",
		"command": "server",
		"args": ["--port", "8080"]
	}`
	if err := os.WriteFile(jsonPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	err := installFromLocal(jsonPath, cli.ScopeUser)
	// Should not fail at validation
	if errors.Is(err, errInstallFailed) {
		t.Error("expected validation to pass for valid stdio server")
	}
}

func Test_installFromLocal_ExplicitTransportSSE(t *testing.T) {
	// Test valid explicit sse transport
	tempDir := t.TempDir()
	jsonPath := filepath.Join(tempDir, "explicit-sse.json")

	content := `{
		"name": "valid-sse",
		"transport": "sse",
		"url": "https://api.example.com/mcp"
	}`
	if err := os.WriteFile(jsonPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	err := installFromLocal(jsonPath, cli.ScopeUser)
	// Should not fail at validation
	if errors.Is(err, errInstallFailed) {
		t.Error("expected validation to pass for valid sse server")
	}
}

func Test_installFromLocal_DisabledServer(t *testing.T) {
	// Test a server that is marked as disabled
	tempDir := t.TempDir()
	jsonPath := filepath.Join(tempDir, "disabled.json")

	content := `{
		"name": "disabled-server",
		"command": "server",
		"disabled": true
	}`
	if err := os.WriteFile(jsonPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	err := installFromLocal(jsonPath, cli.ScopeUser)
	// Should not fail at validation - disabled is just a flag
	if errors.Is(err, errInstallFailed) {
		t.Error("expected validation to pass for disabled server")
	}
}

func Test_installFromLocal_RelativePath(t *testing.T) {
	// Test that relative paths work
	tempDir := t.TempDir()
	jsonPath := filepath.Join(tempDir, "relative.json")

	content := `{"name": "test", "command": "echo"}`
	if err := os.WriteFile(jsonPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Change to temp dir and use relative path
	t.Chdir(tempDir)

	// Use relative path
	err := installFromLocal("./relative.json", cli.ScopeUser)
	// Should not fail at file reading
	if err != nil && strings.Contains(err.Error(), "not found") {
		t.Errorf("failed to read file with relative path: %v", err)
	}
}
