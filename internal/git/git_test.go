package git

import "testing"

func TestIsURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "HTTPS URL",
			input: "https://github.com/user/repo.git",
			want:  true,
		},
		{
			name:  "HTTPS URL without .git suffix",
			input: "https://github.com/user/repo",
			want:  true,
		},
		{
			name:  "HTTP URL",
			input: "http://github.com/user/repo",
			want:  true,
		},
		{
			name:  "git protocol URL",
			input: "git://github.com/user/repo.git",
			want:  true,
		},
		{
			name:  "SSH URL with git@",
			input: "git@github.com:user/repo.git",
			want:  true,
		},
		{
			name:  "SSH URL with git@ no .git suffix",
			input: "git@github.com:user/repo",
			want:  true,
		},
		{
			name:  "URL ending in .git only",
			input: "repo.git",
			want:  true,
		},
		{
			name:  "plain path - relative",
			input: "./my-skill",
			want:  false,
		},
		{
			name:  "plain path - absolute",
			input: "/home/user/my-skill",
			want:  false,
		},
		{
			name:  "plain directory name",
			input: "my-skill",
			want:  false,
		},
		{
			name:  "empty string",
			input: "",
			want:  false,
		},
		{
			name:  "path with git in name but not URL",
			input: "./git-tools",
			want:  false,
		},
		{
			name:  "path containing @ but not git@",
			input: "user@host:path",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsURL(tt.input)
			if got != tt.want {
				t.Errorf("IsURL(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
