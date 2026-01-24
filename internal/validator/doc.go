// Package validator provides a unified validation framework for aix components.
//
// It defines shared types for representing validation issues (errors, warnings,
// info) and results across different domains like Skills, MCP servers,
// and Agents.
//
// # Core Concepts
//
//   - [Severity]: Distinguishes between blocking errors and non-blocking warnings.
//   - [Issue]: Represents a single validation problem with field context.
//   - [Result]: Aggregates multiple issues and provides helper methods.
//
// # Basic Usage
//
//	result := &validator.Result{}
//	if name == "" {
//		result.AddError("name", "is required", name)
//	}
//
//	if result.HasErrors() {
//		// handle validation failure
//	}
package validator
