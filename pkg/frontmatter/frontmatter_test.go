package frontmatter

import (
	"strings"
	"testing"
)

func TestParseHeader(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		wantDescription string
	}{
		{
			name: "valid frontmatter",
			content: `---
description: Header only
---
Ignored body content`,
			wantDescription: "Header only",
		},
		{
			name: "no frontmatter",
			content: `Just body
content`,
			wantDescription: "",
		},
		{
			name:            "empty content",
			content:         "",
			wantDescription: "",
		},
		{
			name: "unclosed frontmatter",
			content: `---
description: Unclosed
body starts here`,
			wantDescription: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var matter struct {
				Description string `yaml:"description"`
			}

			err := ParseHeader(strings.NewReader(tt.content), &matter)
			if err != nil {
				t.Fatalf("ParseHeader() error = %v", err)
			}

			if matter.Description != tt.wantDescription {
				t.Errorf("Description = %q, want %q", matter.Description, tt.wantDescription)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		wantDescription string
		wantBody        string
	}{
		{
			name: "with frontmatter",
			content: `---
description: Test description
---

Body content here.`,
			wantDescription: "Test description",
			wantBody:        "\nBody content here.",
		},
		{
			name:            "without frontmatter",
			content:         "Just body content\nNo frontmatter here",
			wantDescription: "",
			wantBody:        "Just body content\nNo frontmatter here",
		},
		{
			name:            "empty content",
			content:         "",
			wantDescription: "",
			wantBody:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var matter struct {
				Description string `yaml:"description"`
			}

			body, err := Parse(strings.NewReader(tt.content), &matter)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if matter.Description != tt.wantDescription {
				t.Errorf("Parse() description = %q, want %q", matter.Description, tt.wantDescription)
			}
			if string(body) != tt.wantBody {
				t.Errorf("Parse() body = %q, want %q", string(body), tt.wantBody)
			}
		})
	}
}

func TestMustParse(t *testing.T) {
	t.Run("with frontmatter succeeds", func(t *testing.T) {
		content := `---
description: Required frontmatter
---

Body content.`

		var matter struct {
			Description string `yaml:"description"`
		}

		body, err := MustParse(strings.NewReader(content), &matter)
		if err != nil {
			t.Fatalf("MustParse() error = %v", err)
		}

		if matter.Description != "Required frontmatter" {
			t.Errorf("MustParse() description = %q, want %q", matter.Description, "Required frontmatter")
		}
		if string(body) != "\nBody content." {
			t.Errorf("MustParse() body = %q, want %q", string(body), "\nBody content.")
		}
	})

	t.Run("without frontmatter errors", func(t *testing.T) {
		content := "No frontmatter here"

		var matter struct {
			Description string `yaml:"description"`
		}

		_, err := MustParse(strings.NewReader(content), &matter)
		if err == nil {
			t.Error("MustParse() expected error for missing frontmatter")
		}
	})
}

func TestFormat(t *testing.T) {
	tests := []struct {
		name       string
		matter     any
		body       string
		wantPrefix string
		wantSuffix string
	}{
		{
			name: "with description",
			matter: struct {
				Description string `yaml:"description"`
			}{Description: "Test"},
			body:       "Body content",
			wantPrefix: "---\ndescription: Test\n---\n\nBody content\n",
			wantSuffix: "",
		},
		{
			name: "empty body",
			matter: struct {
				Description string `yaml:"description"`
			}{Description: "Meta only"},
			body:       "",
			wantPrefix: "---\ndescription: Meta only\n---\n",
			wantSuffix: "",
		},
		{
			name: "body with trailing newline",
			matter: struct {
				Description string `yaml:"description"`
			}{Description: "Test"},
			body:       "Body\n",
			wantPrefix: "---\ndescription: Test\n---\n\nBody\n",
			wantSuffix: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Format(tt.matter, tt.body)
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}

			got := string(result)
			if got != tt.wantPrefix {
				t.Errorf("Format() = %q, want %q", got, tt.wantPrefix)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	type Matter struct {
		Name        string   `yaml:"name"`
		Description string   `yaml:"description"`
		Version     string   `yaml:"version,omitempty"`
		Tools       []string `yaml:"tools,omitempty"`
	}

	original := Matter{
		Name:        "test-skill",
		Description: "A test skill",
		Version:     "1.0.0",
		Tools:       []string{"Read", "Write"},
	}
	originalBody := "These are the instructions.\n\nWith multiple paragraphs."

	// Format
	formatted, err := Format(original, originalBody)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Parse back
	var parsed Matter
	body, err := MustParse(strings.NewReader(string(formatted)), &parsed)
	if err != nil {
		t.Fatalf("MustParse() error = %v", err)
	}

	// Verify
	if parsed.Name != original.Name {
		t.Errorf("Name = %q, want %q", parsed.Name, original.Name)
	}
	if parsed.Description != original.Description {
		t.Errorf("Description = %q, want %q", parsed.Description, original.Description)
	}
	if parsed.Version != original.Version {
		t.Errorf("Version = %q, want %q", parsed.Version, original.Version)
	}
	if len(parsed.Tools) != len(original.Tools) {
		t.Errorf("Tools len = %d, want %d", len(parsed.Tools), len(original.Tools))
	}

	// Body gets a leading newline from the library
	expectedBody := "\n" + originalBody + "\n"
	if string(body) != expectedBody {
		t.Errorf("body = %q, want %q", string(body), expectedBody)
	}
}
