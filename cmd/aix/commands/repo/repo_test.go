package repo

import (
	"bytes"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/repo"
)

func TestPrintValidationWarnings(t *testing.T) {
	tests := []struct {
		name         string
		warnings     []repo.ValidationWarning
		wantContains []string
		wantEmpty    bool
	}{
		{
			name:      "no warnings",
			warnings:  nil,
			wantEmpty: true,
		},
		{
			name: "only directory not found warnings are filtered",
			warnings: []repo.ValidationWarning{
				{Path: "skills", Message: "directory not found"},
				{Path: "commands", Message: "directory not found"},
			},
			wantEmpty: true,
		},
		{
			name: "actionable warning is shown",
			warnings: []repo.ValidationWarning{
				{Path: "skills/test/SKILL.md", Message: "invalid frontmatter: unexpected EOF"},
			},
			wantContains: []string{
				"Validation warnings:",
				"skills/test/SKILL.md",
				"invalid frontmatter",
			},
		},
		{
			name: "mixed warnings filter correctly",
			warnings: []repo.ValidationWarning{
				{Path: "skills", Message: "directory not found"},
				{Path: "mcp/server.json", Message: "invalid JSON: unexpected end of JSON input"},
				{Path: "commands", Message: "directory not found"},
			},
			wantContains: []string{
				"Validation warnings:",
				"mcp/server.json",
				"invalid JSON",
			},
		},
		{
			name: "multiple actionable warnings",
			warnings: []repo.ValidationWarning{
				{Path: "skills/broken", Message: "skill directory missing SKILL.md"},
				{Path: "agents/test/AGENT.md", Message: "invalid frontmatter: yaml error"},
			},
			wantContains: []string{
				"Validation warnings:",
				"skills/broken",
				"skill directory missing SKILL.md",
				"agents/test/AGENT.md",
				"invalid frontmatter",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			printValidationWarnings(&buf, tt.warnings)
			output := buf.String()

			if tt.wantEmpty {
				if output != "" {
					t.Errorf("expected empty output, got: %q", output)
				}
				return
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("output missing expected content %q\noutput: %s", want, output)
				}
			}
		})
	}
}
