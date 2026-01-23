package search

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/resource"
)

func TestSearchCmd_Metadata(t *testing.T) {
	if Cmd.Use != "search [query]" {
		t.Errorf("Use = %q, want %q", Cmd.Use, "search [query]")
	}

	if Cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if Cmd.Long == "" {
		t.Error("Long description should not be empty")
	}

	if Cmd.Example == "" {
		t.Error("Example should not be empty")
	}
}

func TestSearchCmd_FlagParsing(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
	}{
		{
			name:     "type flag",
			flagName: "type",
		},
		{
			name:     "repo flag",
			flagName: "repo",
		},
		{
			name:     "json flag",
			flagName: "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := Cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("--%s flag should be defined", tt.flagName)
			}
		})
	}
}

func TestSearchCmd_HelpOutput(t *testing.T) {
	// Check that command has help text
	help := Cmd.UsageString()

	if !strings.Contains(help, "search") {
		t.Error("help should contain 'search'")
	}

	if !strings.Contains(help, "--type") {
		t.Error("help should contain '--type' flag")
	}

	if !strings.Contains(help, "--repo") {
		t.Error("help should contain '--repo' flag")
	}

	if !strings.Contains(help, "--json") {
		t.Error("help should contain '--json' flag")
	}
}

func TestOutputJSON_ValidFormat(t *testing.T) {
	resources := []resource.Resource{
		{
			Name:        "deploy-skill",
			Description: "Deploy applications to production",
			Type:        resource.TypeSkill,
			RepoName:    "official",
			Path:        "skills/deploy",
		},
		{
			Name:        "test-command",
			Description: "Run test suites",
			Type:        resource.TypeCommand,
			RepoName:    "community",
			Path:        "commands/test",
		},
	}

	var buf bytes.Buffer
	err := outputJSON(&buf, resources)
	if err != nil {
		t.Fatalf("outputJSON() error = %v", err)
	}

	// Verify JSON is valid
	var result []resource.Resource
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(result))
	}

	// Verify first resource
	if result[0].Name != "deploy-skill" {
		t.Errorf("result[0].Name = %q, want %q", result[0].Name, "deploy-skill")
	}
	if result[0].Type != resource.TypeSkill {
		t.Errorf("result[0].Type = %q, want %q", result[0].Type, resource.TypeSkill)
	}
	if result[0].RepoName != "official" {
		t.Errorf("result[0].RepoName = %q, want %q", result[0].RepoName, "official")
	}

	// Verify second resource
	if result[1].Name != "test-command" {
		t.Errorf("result[1].Name = %q, want %q", result[1].Name, "test-command")
	}
	if result[1].Type != resource.TypeCommand {
		t.Errorf("result[1].Type = %q, want %q", result[1].Type, resource.TypeCommand)
	}
}

func TestOutputJSON_EmptyResources(t *testing.T) {
	resources := []resource.Resource{}

	var buf bytes.Buffer
	err := outputJSON(&buf, resources)
	if err != nil {
		t.Fatalf("outputJSON() error = %v", err)
	}

	// Verify JSON is valid empty array
	var result []resource.Resource
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 resources, got %d", len(result))
	}
}

func TestOutputJSON_FormattedOutput(t *testing.T) {
	resources := []resource.Resource{
		{
			Name:        "test",
			Description: "Test resource",
			Type:        resource.TypeSkill,
			RepoName:    "repo",
			Path:        "path/to/test",
		},
	}

	var buf bytes.Buffer
	err := outputJSON(&buf, resources)
	if err != nil {
		t.Fatalf("outputJSON() error = %v", err)
	}

	output := buf.String()
	// Check that output is indented (contains newlines and spaces for formatting)
	if !strings.Contains(output, "\n") {
		t.Error("JSON output should be formatted with newlines")
	}
	if !strings.Contains(output, "  ") {
		t.Error("JSON output should be formatted with indentation")
	}
}

func TestOutputTabular_WithResources(t *testing.T) {
	resources := []resource.Resource{
		{
			Name:        "deploy-skill",
			Description: "Deploy applications to production",
			Type:        resource.TypeSkill,
			RepoName:    "official",
			Path:        "skills/deploy",
		},
		{
			Name:        "test-agent",
			Description: "Run automated tests",
			Type:        resource.TypeAgent,
			RepoName:    "community",
			Path:        "agents/test",
		},
	}

	var buf bytes.Buffer
	err := outputTabular(&buf, resources)
	if err != nil {
		t.Fatalf("outputTabular() error = %v", err)
	}

	output := buf.String()

	// Check headers
	if !strings.Contains(output, "TYPE") {
		t.Error("output should contain TYPE header")
	}
	if !strings.Contains(output, "REPO") {
		t.Error("output should contain REPO header")
	}
	if !strings.Contains(output, "NAME") {
		t.Error("output should contain NAME header")
	}
	if !strings.Contains(output, "DESCRIPTION") {
		t.Error("output should contain DESCRIPTION header")
	}

	// Check resource data
	if !strings.Contains(output, "deploy-skill") {
		t.Error("output should contain deploy-skill")
	}
	if !strings.Contains(output, "test-agent") {
		t.Error("output should contain test-agent")
	}
	if !strings.Contains(output, "skill") {
		t.Error("output should contain skill type")
	}
	if !strings.Contains(output, "agent") {
		t.Error("output should contain agent type")
	}
	if !strings.Contains(output, "official") {
		t.Error("output should contain official repo name")
	}
	if !strings.Contains(output, "community") {
		t.Error("output should contain community repo name")
	}
}

func TestOutputTabular_NoResults(t *testing.T) {
	resources := []resource.Resource{}

	var buf bytes.Buffer
	err := outputTabular(&buf, resources)
	if err != nil {
		t.Fatalf("outputTabular() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No resources found") {
		t.Error("output should contain 'No resources found' message")
	}
}

func TestOutputTabular_TruncatesLongDescriptions(t *testing.T) {
	longDesc := strings.Repeat("a", 100)
	resources := []resource.Resource{
		{
			Name:        "test",
			Description: longDesc,
			Type:        resource.TypeSkill,
			RepoName:    "repo",
			Path:        "path",
		},
	}

	var buf bytes.Buffer
	err := outputTabular(&buf, resources)
	if err != nil {
		t.Fatalf("outputTabular() error = %v", err)
	}

	output := buf.String()
	// Should contain truncated description with "..."
	if !strings.Contains(output, "...") {
		t.Error("long description should be truncated with ...")
	}
	// Should not contain the full 100 character description
	if strings.Contains(output, longDesc) {
		t.Error("description should be truncated, not full length")
	}
}

func TestOutputTabular_AllResourceTypes(t *testing.T) {
	resources := []resource.Resource{
		{
			Name:        "my-skill",
			Description: "A skill",
			Type:        resource.TypeSkill,
			RepoName:    "repo",
			Path:        "skills/my-skill",
		},
		{
			Name:        "my-command",
			Description: "A command",
			Type:        resource.TypeCommand,
			RepoName:    "repo",
			Path:        "commands/my-command",
		},
		{
			Name:        "my-agent",
			Description: "An agent",
			Type:        resource.TypeAgent,
			RepoName:    "repo",
			Path:        "agents/my-agent",
		},
		{
			Name:        "my-mcp",
			Description: "An MCP server",
			Type:        resource.TypeMCP,
			RepoName:    "repo",
			Path:        "mcp/my-mcp",
		},
	}

	var buf bytes.Buffer
	err := outputTabular(&buf, resources)
	if err != nil {
		t.Fatalf("outputTabular() error = %v", err)
	}

	output := buf.String()

	// All resource types should be present
	if !strings.Contains(output, "skill") {
		t.Error("output should contain skill type")
	}
	if !strings.Contains(output, "command") {
		t.Error("output should contain command type")
	}
	if !strings.Contains(output, "agent") {
		t.Error("output should contain agent type")
	}
	if !strings.Contains(output, "mcp") {
		t.Error("output should contain mcp type")
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "short string unchanged",
			input:  "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "exact length unchanged",
			input:  "hello",
			maxLen: 5,
			want:   "hello",
		},
		{
			name:   "long string truncated with ellipsis",
			input:  "hello world",
			maxLen: 8,
			want:   "hello...",
		},
		{
			name:   "empty string",
			input:  "",
			maxLen: 10,
			want:   "",
		},
		{
			name:   "maxLen less than or equal to 3",
			input:  "hello",
			maxLen: 3,
			want:   "hel",
		},
		{
			name:   "maxLen of 1",
			input:  "hello",
			maxLen: 1,
			want:   "h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}
