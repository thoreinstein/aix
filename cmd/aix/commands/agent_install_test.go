package commands

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/platform/claude"
	"github.com/thoreinstein/aix/internal/platform/opencode"
)

func TestResolveAgentPath(t *testing.T) {
	t.Run("file path returns path directly", func(t *testing.T) {
		// Create a temp file
		tempDir := t.TempDir()
		agentFile := filepath.Join(tempDir, "test-agent.md")
		if err := os.WriteFile(agentFile, []byte("test"), 0o644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		got, err := resolveAgentPath(agentFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != agentFile {
			t.Errorf("resolveAgentPath() = %q, want %q", got, agentFile)
		}
	})

	t.Run("directory with AGENT.md returns AGENT.md path", func(t *testing.T) {
		tempDir := t.TempDir()
		agentFile := filepath.Join(tempDir, "AGENT.md")
		if err := os.WriteFile(agentFile, []byte("test"), 0o644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		got, err := resolveAgentPath(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != agentFile {
			t.Errorf("resolveAgentPath() = %q, want %q", got, agentFile)
		}
	})

	t.Run("directory without AGENT.md returns error", func(t *testing.T) {
		tempDir := t.TempDir()

		_, err := resolveAgentPath(tempDir)
		if err == nil {
			t.Error("expected error for directory without AGENT.md, got nil")
		}
	})

	t.Run("non-existent path returns error", func(t *testing.T) {
		_, err := resolveAgentPath("/non/existent/path")
		if err == nil {
			t.Error("expected error for non-existent path, got nil")
		}
	})
}

func TestParseAgentForPlatform_Claude(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantName  string
		wantDesc  string
		wantInstr string
		wantErr   error
	}{
		{
			name: "valid agent with all fields",
			content: `---
name: test-agent
description: A test agent
---

You are a helpful assistant.`,
			wantName:  "test-agent",
			wantDesc:  "A test agent",
			wantInstr: "\nYou are a helpful assistant.",
			wantErr:   nil,
		},
		{
			name: "valid agent without description",
			content: `---
name: minimal-agent
---

Instructions only.`,
			wantName:  "minimal-agent",
			wantDesc:  "",
			wantInstr: "\nInstructions only.",
			wantErr:   nil,
		},
		{
			name: "missing name returns error",
			content: `---
description: No name here
---

Some instructions.`,
			wantErr: errAgentNameRequired,
		},
		{
			name:    "no frontmatter returns error for missing name",
			content: "Just plain markdown without frontmatter.",
			wantErr: errAgentNameRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAgentForPlatform("claude", []byte(tt.content))

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			agent, ok := got.(*claude.Agent)
			if !ok {
				t.Fatalf("expected *claude.Agent, got %T", got)
			}

			if agent.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", agent.Name, tt.wantName)
			}
			if agent.Description != tt.wantDesc {
				t.Errorf("Description = %q, want %q", agent.Description, tt.wantDesc)
			}
			if agent.Instructions != tt.wantInstr {
				t.Errorf("Instructions = %q, want %q", agent.Instructions, tt.wantInstr)
			}
		})
	}
}

func TestParseAgentForPlatform_OpenCode(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantName  string
		wantDesc  string
		wantMode  string
		wantTemp  float64
		wantInstr string
		wantErr   error
	}{
		{
			name: "valid agent with all fields",
			content: `---
name: opencode-agent
description: An OpenCode agent
mode: chat
temperature: 0.7
---

You are an OpenCode assistant.`,
			wantName:  "opencode-agent",
			wantDesc:  "An OpenCode agent",
			wantMode:  "chat",
			wantTemp:  0.7,
			wantInstr: "\nYou are an OpenCode assistant.",
			wantErr:   nil,
		},
		{
			name: "valid agent with minimal fields",
			content: `---
name: minimal
---

Basic instructions.`,
			wantName:  "minimal",
			wantDesc:  "",
			wantMode:  "",
			wantTemp:  0,
			wantInstr: "\nBasic instructions.",
			wantErr:   nil,
		},
		{
			name: "missing name returns error",
			content: `---
mode: edit
temperature: 0.5
---

Some instructions.`,
			wantErr: errAgentNameRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAgentForPlatform("opencode", []byte(tt.content))

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			agent, ok := got.(*opencode.Agent)
			if !ok {
				t.Fatalf("expected *opencode.Agent, got %T", got)
			}

			if agent.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", agent.Name, tt.wantName)
			}
			if agent.Description != tt.wantDesc {
				t.Errorf("Description = %q, want %q", agent.Description, tt.wantDesc)
			}
			if agent.Mode != tt.wantMode {
				t.Errorf("Mode = %q, want %q", agent.Mode, tt.wantMode)
			}
			if agent.Temperature != tt.wantTemp {
				t.Errorf("Temperature = %v, want %v", agent.Temperature, tt.wantTemp)
			}
			if agent.Instructions != tt.wantInstr {
				t.Errorf("Instructions = %q, want %q", agent.Instructions, tt.wantInstr)
			}
		})
	}
}

func TestParseAgentForPlatform_UnsupportedPlatform(t *testing.T) {
	content := `---
name: test
---

Instructions.`

	_, err := parseAgentForPlatform("unknown", []byte(content))
	if err == nil {
		t.Error("expected error for unsupported platform, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported platform") {
		t.Errorf("error %q should contain 'unsupported platform'", err.Error())
	}
}
