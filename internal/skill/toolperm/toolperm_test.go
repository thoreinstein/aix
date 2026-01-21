package toolperm

import (
	"errors"
	"testing"
)

func TestPermission_String(t *testing.T) {
	tests := []struct {
		name string
		perm Permission
		want string
	}{
		{
			name: "simple tool",
			perm: Permission{Name: "Read"},
			want: "Read",
		},
		{
			name: "tool with scope",
			perm: Permission{Name: "Bash", Scope: "git:*"},
			want: "Bash(git:*)",
		},
		{
			name: "tool with complex scope",
			perm: Permission{Name: "Bash", Scope: "npm:install"},
			want: "Bash(npm:install)",
		},
		{
			name: "empty scope is simple",
			perm: Permission{Name: "Write", Scope: ""},
			want: "Write",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.perm.String(); got != tt.want {
				t.Errorf("Permission.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []Permission
		wantErr bool
	}{
		{
			name:    "empty string",
			input:   "",
			want:    []Permission{},
			wantErr: false,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			want:    []Permission{},
			wantErr: false,
		},
		{
			name:  "single simple tool",
			input: "Read",
			want: []Permission{
				{Name: "Read"},
			},
			wantErr: false,
		},
		{
			name:  "multiple simple tools",
			input: "Read Write Edit",
			want: []Permission{
				{Name: "Read"},
				{Name: "Write"},
				{Name: "Edit"},
			},
			wantErr: false,
		},
		{
			name:  "tool with scope",
			input: "Bash(git:*)",
			want: []Permission{
				{Name: "Bash", Scope: "git:*"},
			},
			wantErr: false,
		},
		{
			name:  "mixed simple and scoped",
			input: "Read Bash(git:*) Write Bash(npm:install)",
			want: []Permission{
				{Name: "Read"},
				{Name: "Bash", Scope: "git:*"},
				{Name: "Write"},
				{Name: "Bash", Scope: "npm:install"},
			},
			wantErr: false,
		},
		{
			name:  "all valid tools from spec",
			input: "Read Write Edit Glob Grep WebFetch Task TodoWrite",
			want: []Permission{
				{Name: "Read"},
				{Name: "Write"},
				{Name: "Edit"},
				{Name: "Glob"},
				{Name: "Grep"},
				{Name: "WebFetch"},
				{Name: "Task"},
				{Name: "TodoWrite"},
			},
			wantErr: false,
		},
		{
			name:  "scoped with make",
			input: "Bash(make:*)",
			want: []Permission{
				{Name: "Bash", Scope: "make:*"},
			},
			wantErr: false,
		},
		{
			name:  "leading and trailing whitespace",
			input: "  Read Write  ",
			want: []Permission{
				{Name: "Read"},
				{Name: "Write"},
			},
			wantErr: false,
		},
		{
			name:  "multiple spaces between tools",
			input: "Read    Write     Edit",
			want: []Permission{
				{Name: "Read"},
				{Name: "Write"},
				{Name: "Edit"},
			},
			wantErr: false,
		},
		{
			name:  "tabs and spaces",
			input: "Read\t\tWrite  Edit",
			want: []Permission{
				{Name: "Read"},
				{Name: "Write"},
				{Name: "Edit"},
			},
			wantErr: false,
		},
		{
			name:    "invalid syntax unclosed paren",
			input:   "Bash(",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid syntax no tool name",
			input:   "(scope)",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid characters in tool name",
			input:   "Tool@Name",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid syntax with hyphen",
			input:   "Tool-Name",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid nested parens",
			input:   "Bash((git:*))",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid empty parens",
			input:   "Bash()",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "underscore in tool name is invalid",
			input:   "Tool_Name",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "tool name starting with number is invalid",
			input:   "2Tool",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "tool name starting with underscore is invalid",
			input:   "_Tool",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "tool name starting with lowercase is invalid",
			input:   "read",
			want:    nil,
			wantErr: true,
		},
		{
			name:  "number in tool name is valid",
			input: "Tool2",
			want: []Permission{
				{Name: "Tool2"},
			},
			wantErr: false,
		},
		{
			name:  "scope with colon and asterisk",
			input: "Bash(docker:*)",
			want: []Permission{
				{Name: "Bash", Scope: "docker:*"},
			},
			wantErr: false,
		},
		{
			name:  "scope with specific command",
			input: "Bash(go:build)",
			want: []Permission{
				{Name: "Bash", Scope: "go:build"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			got, err := p.Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("Parser.Parse() got %d permissions, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i].Name != tt.want[i].Name {
					t.Errorf("Parser.Parse()[%d].Name = %q, want %q", i, got[i].Name, tt.want[i].Name)
				}
				if got[i].Scope != tt.want[i].Scope {
					t.Errorf("Parser.Parse()[%d].Scope = %q, want %q", i, got[i].Scope, tt.want[i].Scope)
				}
			}
		})
	}
}

func TestParser_ParseSingle(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Permission
		wantErr bool
	}{
		{
			name:    "simple tool",
			input:   "Read",
			want:    Permission{Name: "Read"},
			wantErr: false,
		},
		{
			name:    "tool with scope",
			input:   "Bash(git:*)",
			want:    Permission{Name: "Bash", Scope: "git:*"},
			wantErr: false,
		},
		{
			name:    "tool with whitespace around",
			input:   "  Write  ",
			want:    Permission{Name: "Write"},
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			want:    Permission{},
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			want:    Permission{},
			wantErr: true,
		},
		{
			name:    "unclosed paren",
			input:   "Bash(git:*",
			want:    Permission{},
			wantErr: true,
		},
		{
			name:    "no closing paren",
			input:   "Bash(",
			want:    Permission{},
			wantErr: true,
		},
		{
			name:    "empty parens",
			input:   "Bash()",
			want:    Permission{},
			wantErr: true,
		},
		{
			name:    "only parens",
			input:   "(scope)",
			want:    Permission{},
			wantErr: true,
		},
		{
			name:    "special characters",
			input:   "Tool@Name",
			want:    Permission{},
			wantErr: true,
		},
		{
			name:    "lowercase tool is invalid",
			input:   "read",
			want:    Permission{},
			wantErr: true,
		},
		{
			name:    "mixed case tool",
			input:   "WebFetch",
			want:    Permission{Name: "WebFetch"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			got, err := p.ParseSingle(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.ParseSingle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.Name != tt.want.Name {
				t.Errorf("Parser.ParseSingle().Name = %q, want %q", got.Name, tt.want.Name)
			}
			if got.Scope != tt.want.Scope {
				t.Errorf("Parser.ParseSingle().Scope = %q, want %q", got.Scope, tt.want.Scope)
			}
		})
	}
}

func TestParser_Format(t *testing.T) {
	tests := []struct {
		name  string
		perms []Permission
		want  string
	}{
		{
			name:  "empty slice",
			perms: []Permission{},
			want:  "",
		},
		{
			name:  "nil slice",
			perms: nil,
			want:  "",
		},
		{
			name: "single simple tool",
			perms: []Permission{
				{Name: "Read"},
			},
			want: "Read",
		},
		{
			name: "multiple simple tools",
			perms: []Permission{
				{Name: "Read"},
				{Name: "Write"},
				{Name: "Edit"},
			},
			want: "Read Write Edit",
		},
		{
			name: "tool with scope",
			perms: []Permission{
				{Name: "Bash", Scope: "git:*"},
			},
			want: "Bash(git:*)",
		},
		{
			name: "mixed tools",
			perms: []Permission{
				{Name: "Read"},
				{Name: "Bash", Scope: "git:*"},
				{Name: "Write"},
			},
			want: "Read Bash(git:*) Write",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			if got := p.Format(tt.perms); got != tt.want {
				t.Errorf("Parser.Format() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParser_RoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string // normalized form
	}{
		{
			name:  "simple tools",
			input: "Read Write Edit",
			want:  "Read Write Edit",
		},
		{
			name:  "mixed tools",
			input: "Read Bash(git:*) Write",
			want:  "Read Bash(git:*) Write",
		},
		{
			name:  "extra whitespace normalized",
			input: "  Read   Write  ",
			want:  "Read Write",
		},
		{
			name:  "single scoped tool",
			input: "Bash(npm:install)",
			want:  "Bash(npm:install)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			perms, err := p.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parser.Parse() error = %v", err)
			}
			got := p.Format(perms)
			if got != tt.want {
				t.Errorf("Round trip: Parse(%q) -> Format() = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToolPermError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *ToolPermError
		want string
	}{
		{
			name: "with token",
			err:  &ToolPermError{Token: "Bash(", Message: "invalid tool permission syntax: tool name must be PascalCase (start with uppercase letter, e.g., Read, Write, Bash)"},
			want: `invalid tool permission "Bash(": invalid tool permission syntax: tool name must be PascalCase (start with uppercase letter, e.g., Read, Write, Bash)`,
		},
		{
			name: "empty token",
			err:  &ToolPermError{Token: "", Message: "empty tool permission"},
			want: "tool permission error: empty tool permission",
		},
		{
			name: "special characters in token",
			err:  &ToolPermError{Token: "Tool@Name", Message: "invalid tool permission syntax: tool name must be PascalCase (start with uppercase letter, e.g., Read, Write, Bash)"},
			want: `invalid tool permission "Tool@Name": invalid tool permission syntax: tool name must be PascalCase (start with uppercase letter, e.g., Read, Write, Bash)`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("ToolPermError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParser_Parse_ErrorType(t *testing.T) {
	p := New()
	_, err := p.Parse("Bash(")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var permErr *ToolPermError
	if !errors.As(err, &permErr) {
		t.Errorf("error should be *ToolPermError, got %T", err)
	}
	if permErr.Token != "Bash(" {
		t.Errorf("ToolPermError.Token = %q, want %q", permErr.Token, "Bash(")
	}
}

func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Error("New() returned nil")
	}
}
