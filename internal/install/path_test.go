package install

import (
	"testing"
)

func TestLooksLikePath(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   bool
	}{
		{
			name:   "relative path with dot slash",
			source: "./file",
			want:   true,
		},
		{
			name:   "parent relative path",
			source: "../file",
			want:   true,
		},
		{
			name:   "deep parent path",
			source: "../../configs/file",
			want:   true,
		},
		{
			name:   "absolute path unix",
			source: "/path/to/file",
			want:   true,
		},
		{
			name:   "path with separator",
			source: "path/to/file",
			want:   true,
		},
		{
			name:   "simple name",
			source: "my-resource",
			want:   false,
		},
		{
			name:   "name with dash",
			source: "my-name",
			want:   false,
		},
		{
			name:   "name with underscore",
			source: "my_name",
			want:   false,
		},
		{
			name:   "name with dots but no slash",
			source: "resource.json",
			want:   false,
		},
		{
			name:   "empty string",
			source: "",
			want:   false,
		},
		{
			name:   "just a dot",
			source: ".",
			want:   false,
		},
		{
			name:   "single slash",
			source: "/",
			want:   true,
		},
		{
			name:   "current directory explicit",
			source: "./",
			want:   true,
		},
		{
			name:   "nested path no extension",
			source: "subdir/resource",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LooksLikePath(tt.source)
			if got != tt.want {
				t.Errorf("LooksLikePath(%q) = %v, want %v", tt.source, got, tt.want)
			}
		})
	}
}

func TestMightBePath(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		resourceType string
		want         bool
	}{
		{
			name:         "markdown file for skill",
			source:       "my-skill.md",
			resourceType: "skill",
			want:         true,
		},
		{
			name:         "markdown file for agent",
			source:       "AGENT.md",
			resourceType: "agent",
			want:         true,
		},
		{
			name:         "json file for mcp",
			source:       "server.json",
			resourceType: "mcp",
			want:         true,
		},
		{
			name:         "json file for skill (not a path)",
			source:       "server.json",
			resourceType: "skill",
			want:         false,
		},
		{
			name:         "windows path backslash",
			source:       `C:\path\to\file`,
			resourceType: "skill",
			want:         true,
		},
		{
			name:         "just a name",
			source:       "my-resource",
			resourceType: "skill",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MightBePath(tt.source, tt.resourceType)
			if got != tt.want {
				t.Errorf("MightBePath(%q, %q) = %v, want %v", tt.source, tt.resourceType, got, tt.want)
			}
		})
	}
}
