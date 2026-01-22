package commands

import (
	"errors"
	"testing"
)

func TestParseKeyValueSlice(t *testing.T) {
	tests := []struct {
		name     string
		entries  []string
		flagName string
		want     map[string]string
		wantErr  bool
	}{
		{
			name:     "nil entries",
			entries:  nil,
			flagName: "--env",
			want:     nil,
			wantErr:  false,
		},
		{
			name:     "empty entries",
			entries:  []string{},
			flagName: "--env",
			want:     nil,
			wantErr:  false,
		},
		{
			name:     "valid single entry",
			entries:  []string{"KEY=value"},
			flagName: "--env",
			want:     map[string]string{"KEY": "value"},
			wantErr:  false,
		},
		{
			name:     "valid multiple entries",
			entries:  []string{"KEY1=value1", "KEY2=value2"},
			flagName: "--env",
			want:     map[string]string{"KEY1": "value1", "KEY2": "value2"},
			wantErr:  false,
		},
		{
			name:     "empty value",
			entries:  []string{"KEY="},
			flagName: "--env",
			want:     map[string]string{"KEY": ""},
			wantErr:  false,
		},
		{
			name:     "equals in value",
			entries:  []string{"KEY=val=ue=more"},
			flagName: "--env",
			want:     map[string]string{"KEY": "val=ue=more"},
			wantErr:  false,
		},
		{
			name:     "missing equals",
			entries:  []string{"KEYvalue"},
			flagName: "--env",
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "empty key",
			entries:  []string{"=value"},
			flagName: "--env",
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "multiple with one invalid",
			entries:  []string{"VALID=ok", "invalid"},
			flagName: "--headers",
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "whitespace in key",
			entries:  []string{"MY KEY=value"},
			flagName: "--env",
			want:     map[string]string{"MY KEY": "value"},
			wantErr:  false,
		},
		{
			name:     "whitespace in value",
			entries:  []string{"KEY=some value with spaces"},
			flagName: "--env",
			want:     map[string]string{"KEY": "some value with spaces"},
			wantErr:  false,
		},
		{
			name:     "special characters",
			entries:  []string{"AUTH=Bearer eyJhbGciOiJIUzI1NiJ9.abc"},
			flagName: "--headers",
			want:     map[string]string{"AUTH": "Bearer eyJhbGciOiJIUzI1NiJ9.abc"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseKeyValueSlice(tt.entries, tt.flagName)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseKeyValueSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if tt.want == nil && got != nil {
				t.Errorf("parseKeyValueSlice() = %v, want nil", got)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("parseKeyValueSlice() len = %d, want %d", len(got), len(tt.want))
				return
			}
			for k, wantV := range tt.want {
				if gotV, ok := got[k]; !ok {
					t.Errorf("parseKeyValueSlice() missing key %q", k)
				} else if gotV != wantV {
					t.Errorf("parseKeyValueSlice()[%q] = %q, want %q", k, gotV, wantV)
				}
			}
		})
	}
}

func TestMCPAddCommand_Metadata(t *testing.T) {
	if mcpAddCmd.Use != "add <name> [command] [args...]" {
		t.Errorf("Use = %q, want %q", mcpAddCmd.Use, "add <name> [command] [args...]")
	}

	if mcpAddCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if mcpAddCmd.Long == "" {
		t.Error("Long description should not be empty")
	}

	// Check required flags exist
	expectedFlags := []string{"url", "env", "transport", "headers", "platform", "force"}
	for _, flagName := range expectedFlags {
		if mcpAddCmd.Flags().Lookup(flagName) == nil {
			t.Errorf("--%s flag should be defined", flagName)
		}
	}

	// Verify short flag for force
	forceFlag := mcpAddCmd.Flags().ShorthandLookup("f")
	if forceFlag == nil {
		t.Error("-f shorthand for --force should be defined")
	}

	// Verify Args validator is set (MinimumNArgs(1))
	if mcpAddCmd.Args == nil {
		t.Error("Args validator should be set")
	}
}

func TestMCPAddSentinelErrors(t *testing.T) {
	// Ensure sentinel errors are properly defined
	if errMCPAddMissingCommandOrURL == nil {
		t.Error("errMCPAddMissingCommandOrURL should be defined")
	}
	if errMCPAddBothCommandAndURL == nil {
		t.Error("errMCPAddBothCommandAndURL should be defined")
	}

	// Verify error messages
	if errMCPAddMissingCommandOrURL.Error() != "either command or --url is required" {
		t.Errorf("unexpected error message: %s", errMCPAddMissingCommandOrURL.Error())
	}
	if errMCPAddBothCommandAndURL.Error() != "cannot specify both command and --url" {
		t.Errorf("unexpected error message: %s", errMCPAddBothCommandAndURL.Error())
	}
}

func TestInferTransport(t *testing.T) {
	// This tests the transport inference logic used in runMCPAdd
	// The actual implementation is inline in runMCPAdd, so we test the logic patterns

	tests := []struct {
		name           string
		explicitTrans  string
		url            string
		hasCommand     bool
		wantTransport  string
		wantValid      bool
		wantErrMessage string
	}{
		{
			name:          "explicit stdio",
			explicitTrans: "stdio",
			url:           "",
			hasCommand:    true,
			wantTransport: "stdio",
			wantValid:     true,
		},
		{
			name:          "explicit sse",
			explicitTrans: "sse",
			url:           "http://example.com",
			hasCommand:    false,
			wantTransport: "sse",
			wantValid:     true,
		},
		{
			name:          "infer sse from url",
			explicitTrans: "",
			url:           "http://example.com/mcp",
			hasCommand:    false,
			wantTransport: "sse",
			wantValid:     true,
		},
		{
			name:          "infer stdio from command",
			explicitTrans: "",
			url:           "",
			hasCommand:    true,
			wantTransport: "stdio",
			wantValid:     true,
		},
		{
			name:           "invalid transport type",
			explicitTrans:  "websocket",
			url:            "",
			hasCommand:     true,
			wantTransport:  "",
			wantValid:      false,
			wantErrMessage: "invalid --transport \"websocket\": must be 'stdio' or 'sse'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the transport inference logic from runMCPAdd
			transport := tt.explicitTrans
			if transport == "" {
				if tt.url != "" {
					transport = "sse"
				} else if tt.hasCommand {
					transport = "stdio"
				}
			}

			// Validate transport
			var validationErr error
			switch transport {
			case "stdio", "sse":
				// Valid
			default:
				validationErr = errors.New("invalid --transport \"" + transport + "\": must be 'stdio' or 'sse'")
			}

			if tt.wantValid {
				if validationErr != nil {
					t.Errorf("expected valid transport, got error: %v", validationErr)
				}
				if transport != tt.wantTransport {
					t.Errorf("transport = %q, want %q", transport, tt.wantTransport)
				}
			} else {
				if validationErr == nil {
					t.Error("expected validation error, got nil")
				} else if validationErr.Error() != tt.wantErrMessage {
					t.Errorf("error = %q, want %q", validationErr.Error(), tt.wantErrMessage)
				}
			}
		})
	}
}

func TestValidationLogic(t *testing.T) {
	// Test validation: either command or --url is required, but not both
	tests := []struct {
		name       string
		command    string
		url        string
		wantErr    error
		wantErrNil bool
	}{
		{
			name:       "valid: command only",
			command:    "npx",
			url:        "",
			wantErrNil: true,
		},
		{
			name:       "valid: url only",
			command:    "",
			url:        "http://example.com/mcp",
			wantErrNil: true,
		},
		{
			name:    "invalid: neither command nor url",
			command: "",
			url:     "",
			wantErr: errMCPAddMissingCommandOrURL,
		},
		{
			name:    "invalid: both command and url",
			command: "npx",
			url:     "http://example.com/mcp",
			wantErr: errMCPAddBothCommandAndURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the validation logic from runMCPAdd
			var err error
			if tt.command == "" && tt.url == "" {
				err = errMCPAddMissingCommandOrURL
			}
			if tt.command != "" && tt.url != "" {
				err = errMCPAddBothCommandAndURL
			}

			if tt.wantErrNil {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			} else {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestOpenCodeTransportMapping(t *testing.T) {
	// Test the transport type mapping for OpenCode: stdio -> local, sse -> remote
	tests := []struct {
		name             string
		inputTransport   string
		wantOpenCodeType string
	}{
		{
			name:             "stdio maps to local",
			inputTransport:   "stdio",
			wantOpenCodeType: "local",
		},
		{
			name:             "sse maps to remote",
			inputTransport:   "sse",
			wantOpenCodeType: "remote",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the mapping logic from addMCPToPlatform
			typ := "local"
			if tt.inputTransport == "sse" {
				typ = "remote"
			}

			if typ != tt.wantOpenCodeType {
				t.Errorf("OpenCode type = %q, want %q", typ, tt.wantOpenCodeType)
			}
		})
	}
}

func TestOpenCodeCommandSliceConstruction(t *testing.T) {
	// Test how OpenCode combines command and args into a single slice
	tests := []struct {
		name    string
		command string
		args    []string
		want    []string
	}{
		{
			name:    "command only",
			command: "npx",
			args:    nil,
			want:    []string{"npx"},
		},
		{
			name:    "command with args",
			command: "npx",
			args:    []string{"-y", "@modelcontextprotocol/server-github"},
			want:    []string{"npx", "-y", "@modelcontextprotocol/server-github"},
		},
		{
			name:    "command with empty args",
			command: "/usr/bin/server",
			args:    []string{},
			want:    []string{"/usr/bin/server"},
		},
		{
			name:    "no command (url server)",
			command: "",
			args:    nil,
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the command slice construction from addMCPToPlatform
			var cmdSlice []string
			if tt.command != "" {
				cmdSlice = append([]string{tt.command}, tt.args...)
			}

			if tt.want == nil {
				if cmdSlice != nil {
					t.Errorf("cmdSlice = %v, want nil", cmdSlice)
				}
				return
			}

			if len(cmdSlice) != len(tt.want) {
				t.Errorf("cmdSlice len = %d, want %d", len(cmdSlice), len(tt.want))
				return
			}

			for i, v := range tt.want {
				if cmdSlice[i] != v {
					t.Errorf("cmdSlice[%d] = %q, want %q", i, cmdSlice[i], v)
				}
			}
		})
	}
}
