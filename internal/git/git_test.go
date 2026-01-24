package git

import "testing"

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		// Valid URLs
		{"https", "https://github.com/user/repo.git", false},
		{"http", "http://github.com/user/repo.git", false},
		{"ssh", "ssh://git@github.com/user/repo.git", false},
		{"git", "git://github.com/user/repo.git", false},
		{"file", "file:///path/to/repo.git", false},
		{"scp-like", "git@github.com:user/repo.git", false},
		{"scp-like subdomain", "git@sub.domain.com:user/repo.git", false},
		{"scp-like user", "user@host.com:path/to/repo.git", false},

		// Invalid URLs
		{"empty", "", true},
		{"argument injection", "-oProxyCommand=touch /tmp/pwned", true},
		{"ext protocol", "ext::sh -c touch% /tmp/pwned", true},
		{"unknown scheme", "ftp://github.com/user/repo.git", true},
		{"missing scheme", "github.com/user/repo.git", true},              // We require scheme or scp-like
		{"scp-like missing git suffix", "git@github.com:user/repo", true}, // Regex requires .git suffix
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}
