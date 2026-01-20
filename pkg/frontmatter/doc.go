// Package frontmatter provides generic parsing of YAML frontmatter from
// Markdown files used by the aix CLI for skills and commands.
//
// Frontmatter is delimited by lines containing only "---" at the start and end.
// The content between delimiters is parsed as YAML and unmarshaled into the
// type parameter T. The remaining content after the closing delimiter is
// returned as the body.
//
// # Basic Usage
//
//	type SkillMeta struct {
//		Name        string   `yaml:"name"`
//		Description string   `yaml:"description"`
//		Tools       []string `yaml:"tools"`
//	}
//
//	meta, body, err := frontmatter.ParseFile[SkillMeta]("skill.md")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Skill: %s\nInstructions:\n%s", meta.Name, body)
//
// # Error Handling
//
// The package defines sentinel errors for common failure conditions:
//
//   - [ErrNoFrontmatter]: file doesn't start with "---" delimiter
//   - [ErrInvalidYAML]: frontmatter exists but contains invalid YAML
//
// These can be checked using [errors.Is]:
//
//	meta, body, err := frontmatter.Parse[SkillMeta](r)
//	if errors.Is(err, frontmatter.ErrNoFrontmatter) {
//		// handle missing frontmatter
//	}
//
// # Supported Formats
//
// The parser supports YAML frontmatter with the standard "---" delimiters.
// Both Unix (LF) and Windows (CRLF) line endings are handled correctly.
package frontmatter
