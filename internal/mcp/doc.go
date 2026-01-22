// Package mcp provides canonical MCP (Model Context Protocol) server configuration
// types that serve as the bridge between different AI assistant platforms.
//
// This package defines platform-agnostic types for representing MCP server
// configurations. These types can be translated to and from platform-specific
// formats (Claude Code, OpenCode, Codex, Gemini CLI) using the respective
// platform adapters.
//
// # Server Configuration
//
// The [Server] type represents a single MCP server with support for both
// local (stdio) and remote (SSE) transports:
//
//	// Local stdio server
//	server := &mcp.Server{
//	    Name:    "github",
//	    Command: "npx",
//	    Args:    []string{"-y", "@modelcontextprotocol/server-github"},
//	    Env:     map[string]string{"GITHUB_TOKEN": "${GITHUB_TOKEN}"},
//	}
//
//	// Remote SSE server
//	server := &mcp.Server{
//	    Name:      "remote-api",
//	    URL:       "https://api.example.com/mcp",
//	    Transport: mcp.TransportSSE,
//	    Headers:   map[string]string{"Authorization": "Bearer ${API_KEY}"},
//	}
//
// # Transport Types
//
// MCP supports two transport mechanisms:
//
//   - [TransportStdio]: Local process communication via stdin/stdout (default)
//   - [TransportSSE]: Remote server communication via Server-Sent Events
//
// Use the [Server.IsLocal] and [Server.IsRemote] helper methods to determine
// the transport type:
//
//	if server.IsLocal() {
//	    // Launch process with Command and Args
//	}
//	if server.IsRemote() {
//	    // Connect to URL with Headers
//	}
//
// # Configuration Container
//
// The [Config] type holds a collection of server configurations:
//
//	config := mcp.NewConfig()
//	config.Servers["github"] = &mcp.Server{
//	    Name:    "github",
//	    Command: "npx",
//	    Args:    []string{"-y", "@modelcontextprotocol/server-github"},
//	}
//
// # Forward Compatibility
//
// Both [Server] and [Config] preserve unknown JSON fields during serialization.
// This ensures forward compatibility when future MCP versions add new fields:
//
//	// Unknown fields are preserved through marshal/unmarshal cycles
//	data := []byte(`{"servers": {...}, "futureField": "value"}`)
//	var config mcp.Config
//	json.Unmarshal(data, &config)
//	output, _ := json.Marshal(&config)
//	// output contains futureField
//
// # Platform Translation
//
// These canonical types are designed to be translated to/from platform-specific
// formats. See the platform packages for translation logic:
//
//   - internal/platform/claude: Claude Code format (Command string, Args []string)
//   - internal/platform/opencode: OpenCode format (Command []string, Type string)
package mcp
