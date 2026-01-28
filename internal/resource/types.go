// Package resource defines types for shareable aix resources (skills, commands,
// agents, and MCP servers) that can be discovered and installed from repositories.
package resource

import (
	"path/filepath"

	"github.com/thoreinstein/aix/internal/paths"
)

// ResourceType identifies the kind of resource.
type ResourceType string

// Resource type constants.
const (
	TypeSkill   ResourceType = "skill"
	TypeCommand ResourceType = "command"
	TypeAgent   ResourceType = "agent"
	TypeMCP     ResourceType = "mcp"
)

// Resource represents a shareable aix resource that can be discovered and
// installed from a repository.
type Resource struct {
	// Name is the unique identifier for this resource within its repository.
	Name string `json:"name"`

	// Description provides a brief explanation of what this resource does.
	Description string `json:"description,omitempty"`

	// Type identifies the kind of resource (skill, command, agent, or mcp).
	Type ResourceType `json:"type"`

	// RepoName is the short name of the repository containing this resource.
	RepoName string `json:"repo_name"`

	// RepoURL is the full URL to the repository.
	RepoURL string `json:"repo_url,omitempty"`

	// Path is the relative path to this resource within the repository.
	Path string `json:"path"`

	// Metadata contains additional key-value pairs for extensibility.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// SourcePath returns the absolute path to the resource in the persistent repository cache.
func (r *Resource) SourcePath() string {
	return filepath.Join(paths.ReposCacheDir(), r.RepoName, r.Path)
}
