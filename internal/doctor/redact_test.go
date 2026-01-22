package doctor

import (
	"testing"
)

func TestShouldMask(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		// Positive cases - should mask
		{"GITHUB_TOKEN", true},
		{"github_token", true},
		{"API_KEY", true},
		{"api_key", true},
		{"SECRET_VALUE", true},
		{"my_secret", true},
		{"PASSWORD", true},
		{"db_password", true},
		{"AUTH_HEADER", true},
		{"oauth_token", true},
		{"CREDENTIAL", true},
		{"aws_credential", true},
		{"PRIVATE_KEY", true},
		{"ssh_private", true},

		// Negative cases - should not mask
		{"PATH", false},
		{"HOME", false},
		{"USER", false},
		{"SHELL", false},
		{"DEBUG", false},
		{"LOG_LEVEL", false},
		{"DATABASE_URL", false}, // URL might contain creds, but key doesn't indicate secret
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := ShouldMask(tt.key)
			if got != tt.want {
				t.Errorf("ShouldMask(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestContainsTokenPrefix(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		// Positive cases - known prefixes
		{"ghp_abc123def456", true},
		{"gho_abc123def456", true},
		{"ghu_abc123def456", true},
		{"ghs_abc123def456", true},
		{"ghr_abc123def456", true},
		{"sk-abc123def456", true},
		{"pk-abc123def456", true},
		{"AKIAIOSFODNN7EXAMPLE", true},
		{"xoxb-123-456-abc", true},
		{"xoxp-123-456-abc", true},
		{"xoxa-123-456-abc", true},
		{"xoxr-123-456-abc", true},

		// Negative cases
		{"some_random_value", false},
		{"ghp", false},   // Too short, not a prefix
		{"_ghp_", false}, // Prefix in middle
		{"", false},
		{"sk", false},
		{"normal_string", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got := ContainsTokenPrefix(tt.value)
			if got != tt.want {
				t.Errorf("ContainsTokenPrefix(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestMaskValue(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"empty", "", "********"},
		{"single char", "a", "********"},
		{"four chars", "abcd", "********"},
		{"five chars", "abcde", "****bcde"},
		{"long value", "ghp_abc123def456xyz", "****6xyz"},
		{"medium", "secret", "****cret"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskValue(tt.value)
			if got != tt.want {
				t.Errorf("MaskValue(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestMaskURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "no credentials",
			url:  "https://example.com/path",
			want: "https://example.com/path",
		},
		{
			name: "user only no password",
			url:  "https://user@example.com/path",
			want: "https://user@example.com/path",
		},
		{
			name: "user and short password",
			url:  "https://user:pwd@example.com/path",
			// Note: url.UserPassword URL-encodes the asterisks
			want: "https://user:%2A%2A%2A%2A%2A%2A%2A%2A@example.com/path",
		},
		{
			name: "user and long password",
			url:  "https://user:secretpassword@example.com/path",
			// Note: url.UserPassword URL-encodes the asterisks
			want: "https://user:%2A%2A%2A%2Aword@example.com/path",
		},
		{
			name: "empty string",
			url:  "",
			want: "",
		},
		{
			name: "invalid url passthrough",
			url:  "not a url at all ::::",
			want: "not a url at all ::::",
		},
		{
			name: "with port",
			url:  "https://admin:supersecret123@db.example.com:5432/mydb",
			// Note: url.UserPassword URL-encodes the asterisks
			want: "https://admin:%2A%2A%2A%2At123@db.example.com:5432/mydb",
		},
		{
			name: "empty password",
			url:  "https://user:@example.com/path",
			want: "https://user:@example.com/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskURL(tt.url)
			if got != tt.want {
				t.Errorf("MaskURL(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestMaskSecrets(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want map[string]string
	}{
		{
			name: "nil map",
			env:  nil,
			want: nil,
		},
		{
			name: "empty map",
			env:  map[string]string{},
			want: map[string]string{},
		},
		{
			name: "no secrets",
			env: map[string]string{
				"PATH":  "/usr/bin",
				"HOME":  "/home/user",
				"DEBUG": "true",
			},
			want: map[string]string{
				"PATH":  "/usr/bin",
				"HOME":  "/home/user",
				"DEBUG": "true",
			},
		},
		{
			name: "key-based masking",
			env: map[string]string{
				"GITHUB_TOKEN": "ghp_abc123xyz",
				"API_KEY":      "sk-1234567890",
				"PATH":         "/usr/bin",
			},
			want: map[string]string{
				"GITHUB_TOKEN": "****3xyz",
				"API_KEY":      "****7890",
				"PATH":         "/usr/bin",
			},
		},
		{
			name: "value-based masking (token prefix)",
			env: map[string]string{
				"MY_CUSTOM_VAR": "ghp_abc123xyz", // Value has token prefix
				"PATH":          "/usr/bin",
			},
			want: map[string]string{
				"MY_CUSTOM_VAR": "****3xyz",
				"PATH":          "/usr/bin",
			},
		},
		{
			name: "short secret",
			env: map[string]string{
				"API_KEY": "abc",
			},
			want: map[string]string{
				"API_KEY": "********",
			},
		},
		{
			name: "mixed case keys",
			env: map[string]string{
				"github_TOKEN": "value12345",
				"Api_Key":      "value67890",
			},
			want: map[string]string{
				"github_TOKEN": "****2345",
				"Api_Key":      "****7890",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskSecrets(tt.env)

			if tt.want == nil {
				if got != nil {
					t.Errorf("MaskSecrets() = %v, want nil", got)
				}
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("MaskSecrets() length = %d, want %d", len(got), len(tt.want))
				return
			}

			for k, wantV := range tt.want {
				gotV, ok := got[k]
				if !ok {
					t.Errorf("MaskSecrets() missing key %q", k)
					continue
				}
				if gotV != wantV {
					t.Errorf("MaskSecrets()[%q] = %q, want %q", k, gotV, wantV)
				}
			}
		})
	}
}

func TestMaskSecrets_DoesNotMutateInput(t *testing.T) {
	original := map[string]string{
		"GITHUB_TOKEN": "ghp_original_secret",
		"PATH":         "/usr/bin",
	}

	// Copy original values
	originalToken := original["GITHUB_TOKEN"]
	originalPath := original["PATH"]

	_ = MaskSecrets(original)

	// Verify original was not mutated
	if original["GITHUB_TOKEN"] != originalToken {
		t.Errorf("MaskSecrets mutated input: GITHUB_TOKEN changed from %q to %q",
			originalToken, original["GITHUB_TOKEN"])
	}
	if original["PATH"] != originalPath {
		t.Errorf("MaskSecrets mutated input: PATH changed from %q to %q",
			originalPath, original["PATH"])
	}
}
