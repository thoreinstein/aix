package command

import "testing"

func TestInferName(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "simple filename with md extension",
			path: "review.md",
			want: "review",
		},
		{
			name: "hyphenated filename",
			path: "my-command.md",
			want: "my-command",
		},
		{
			name: "absolute path strips directory",
			path: "/path/to/review.md",
			want: "review",
		},
		{
			name: "filename without extension unchanged",
			path: "review",
			want: "review",
		},
		{
			name: "multiple dots only strips md",
			path: "file.test.md",
			want: "file.test",
		},
		{
			name: "empty string returns dot",
			path: "",
			want: ".",
		},
		{
			name: "only md extension returns empty",
			path: ".md",
			want: "",
		},
		{
			name: "relative path with parent dirs",
			path: "../commands/check.md",
			want: "check",
		},
		{
			name: "nested directory in filename strips path",
			path: "nested.dir/file.md",
			want: "file",
		},

		{
			name: "multiple parent directories",
			path: "../../deeply/nested/path/cmd.md",
			want: "cmd",
		},
		{
			name: "uppercase preserved",
			path: "MyCommand.md",
			want: "MyCommand",
		},
		{
			name: "numbers in name",
			path: "review2.md",
			want: "review2",
		},
		{
			name: "underscores preserved",
			path: "my_command.md",
			want: "my_command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InferName(tt.path)
			if got != tt.want {
				t.Errorf("InferName(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
