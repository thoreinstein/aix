package resource

import (
	"testing"
)

func TestFilterByNameAndType(t *testing.T) {
	resources := []Resource{
		{Name: "code-review", Type: TypeSkill, RepoName: "official"},
		{Name: "code-review", Type: TypeCommand, RepoName: "community"},
		{Name: "deploy", Type: TypeCommand, RepoName: "official"},
		{Name: "helper", Type: TypeAgent, RepoName: "official"},
		{Name: "github", Type: TypeMCP, RepoName: "official"},
	}

	tests := []struct {
		name         string
		searchName   string
		resourceType ResourceType
		wantCount    int
	}{
		{
			name:         "find skill by name",
			searchName:   "code-review",
			resourceType: TypeSkill,
			wantCount:    1,
		},
		{
			name:         "find command by name",
			searchName:   "code-review",
			resourceType: TypeCommand,
			wantCount:    1,
		},
		{
			name:         "no match - wrong type",
			searchName:   "code-review",
			resourceType: TypeAgent,
			wantCount:    0,
		},
		{
			name:         "no match - wrong name",
			searchName:   "nonexistent",
			resourceType: TypeSkill,
			wantCount:    0,
		},
		{
			name:         "find agent",
			searchName:   "helper",
			resourceType: TypeAgent,
			wantCount:    1,
		},
		{
			name:         "find mcp server",
			searchName:   "github",
			resourceType: TypeMCP,
			wantCount:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterByNameAndType(resources, tt.searchName, tt.resourceType)
			if len(result) != tt.wantCount {
				t.Errorf("filterByNameAndType() returned %d results, want %d", len(result), tt.wantCount)
			}
			// Verify all results match the criteria
			for _, r := range result {
				if r.Name != tt.searchName {
					t.Errorf("filterByNameAndType() returned resource with name %q, want %q", r.Name, tt.searchName)
				}
				if r.Type != tt.resourceType {
					t.Errorf("filterByNameAndType() returned resource with type %q, want %q", r.Type, tt.resourceType)
				}
			}
		})
	}
}

func TestFilterByNameAndType_DuplicatesAcrossRepos(t *testing.T) {
	// Same resource name exists in multiple repos
	resources := []Resource{
		{Name: "deploy", Type: TypeCommand, RepoName: "official"},
		{Name: "deploy", Type: TypeCommand, RepoName: "community"},
		{Name: "deploy", Type: TypeCommand, RepoName: "private"},
	}

	result := filterByNameAndType(resources, "deploy", TypeCommand)
	if len(result) != 3 {
		t.Errorf("filterByNameAndType() returned %d results, want 3", len(result))
	}
}

func TestFilterByNameAndType_EmptyInput(t *testing.T) {
	t.Run("nil resources", func(t *testing.T) {
		result := filterByNameAndType(nil, "test", TypeSkill)
		if len(result) != 0 {
			t.Errorf("filterByNameAndType(nil, ...) returned %d results, want 0", len(result))
		}
	})

	t.Run("empty resources", func(t *testing.T) {
		result := filterByNameAndType([]Resource{}, "test", TypeSkill)
		if len(result) != 0 {
			t.Errorf("filterByNameAndType([], ...) returned %d results, want 0", len(result))
		}
	})

	t.Run("empty name", func(t *testing.T) {
		resources := []Resource{
			{Name: "test", Type: TypeSkill, RepoName: "r"},
		}
		result := filterByNameAndType(resources, "", TypeSkill)
		if len(result) != 0 {
			t.Errorf("filterByNameAndType(..., \"\", ...) returned %d results, want 0", len(result))
		}
	})
}

func TestFilterByNameAndType_ExactMatchOnly(t *testing.T) {
	resources := []Resource{
		{Name: "code", Type: TypeSkill, RepoName: "r"},
		{Name: "code-review", Type: TypeSkill, RepoName: "r"},
		{Name: "my-code", Type: TypeSkill, RepoName: "r"},
	}

	// Should only match exact name, not prefix or suffix
	result := filterByNameAndType(resources, "code", TypeSkill)
	if len(result) != 1 {
		t.Errorf("filterByNameAndType() returned %d results, want 1 (exact match only)", len(result))
	}
	if len(result) > 0 && result[0].Name != "code" {
		t.Errorf("filterByNameAndType() returned %q, want %q", result[0].Name, "code")
	}
}

func TestFilterByNameAndType_CaseSensitive(t *testing.T) {
	resources := []Resource{
		{Name: "Deploy", Type: TypeCommand, RepoName: "r"},
		{Name: "deploy", Type: TypeCommand, RepoName: "r"},
		{Name: "DEPLOY", Type: TypeCommand, RepoName: "r"},
	}

	// Filter is case-sensitive - exact match only
	result := filterByNameAndType(resources, "deploy", TypeCommand)
	if len(result) != 1 {
		t.Errorf("filterByNameAndType() returned %d results, want 1 (case-sensitive)", len(result))
	}
	if len(result) > 0 && result[0].Name != "deploy" {
		t.Errorf("filterByNameAndType() returned %q, want %q", result[0].Name, "deploy")
	}
}

func TestFilterByNameAndType_AllTypes(t *testing.T) {
	// Verify filtering works for all resource types
	types := []ResourceType{TypeSkill, TypeCommand, TypeAgent, TypeMCP}

	for _, rt := range types {
		t.Run(string(rt), func(t *testing.T) {
			resources := []Resource{
				{Name: "test", Type: TypeSkill, RepoName: "r"},
				{Name: "test", Type: TypeCommand, RepoName: "r"},
				{Name: "test", Type: TypeAgent, RepoName: "r"},
				{Name: "test", Type: TypeMCP, RepoName: "r"},
			}

			result := filterByNameAndType(resources, "test", rt)
			if len(result) != 1 {
				t.Errorf("filterByNameAndType(..., %q) returned %d results, want 1", rt, len(result))
			}
			if len(result) > 0 && result[0].Type != rt {
				t.Errorf("filterByNameAndType() returned type %q, want %q", result[0].Type, rt)
			}
		})
	}
}
