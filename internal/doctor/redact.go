// Package doctor provides diagnostic and validation utilities for aix configurations.
package doctor

import (
	"net/url"
	"strings"
)

// SecretKeyPatterns contains substrings that indicate a key likely contains sensitive data.
// Keys are matched case-insensitively.
var SecretKeyPatterns = []string{
	"TOKEN",
	"KEY",
	"SECRET",
	"PASSWORD",
	"AUTH",
	"CREDENTIAL",
	"API_KEY",
	"PRIVATE",
}

// TokenPrefixes contains known API token prefixes that indicate sensitive values
// regardless of key name.
var TokenPrefixes = []string{
	"ghp_",  // GitHub personal access token
	"gho_",  // GitHub OAuth token
	"ghu_",  // GitHub user-to-server token
	"ghs_",  // GitHub server-to-server token
	"ghr_",  // GitHub refresh token
	"sk-",   // OpenAI/Anthropic keys
	"pk-",   // Public keys that shouldn't be exposed
	"AKIA",  // AWS access key prefix
	"xoxb-", // Slack bot token
	"xoxp-", // Slack user token
	"xoxa-", // Slack app token
	"xoxr-", // Slack refresh token
}

// MaskSecrets masks sensitive values in the given environment variable map.
// Keys matching SecretKeyPatterns or values matching TokenPrefixes are masked.
// Returns a new map with sensitive values redacted.
func MaskSecrets(env map[string]string) map[string]string {
	if env == nil {
		return nil
	}

	masked := make(map[string]string, len(env))
	for k, v := range env {
		if ShouldMask(k) || ContainsTokenPrefix(v) {
			masked[k] = MaskValue(v)
		} else {
			masked[k] = v
		}
	}
	return masked
}

// MaskValue masks a potentially sensitive string value.
// Values with 4 or fewer characters are fully masked as "********".
// Longer values show the last 4 characters: "****xxxx".
func MaskValue(value string) string {
	if len(value) <= 4 {
		return "********"
	}
	return "****" + value[len(value)-4:]
}

// MaskURL redacts credentials from URLs.
// URLs with embedded credentials (user:pass@host) become (user:****@host).
// If the URL cannot be parsed, it is returned unchanged.
func MaskURL(rawURL string) string {
	if rawURL == "" {
		return rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// No user info, nothing to mask
	if parsed.User == nil {
		return rawURL
	}

	password, hasPassword := parsed.User.Password()
	if !hasPassword || password == "" {
		return rawURL
	}

	// Create new URL with masked password
	maskedPassword := MaskValue(password)
	parsed.User = url.UserPassword(parsed.User.Username(), maskedPassword)

	return parsed.String()
}

// ShouldMask returns true if the key name suggests it contains sensitive data.
// Matching is case-insensitive.
func ShouldMask(key string) bool {
	upper := strings.ToUpper(key)
	for _, pattern := range SecretKeyPatterns {
		if strings.Contains(upper, pattern) {
			return true
		}
	}
	return false
}

// ContainsTokenPrefix returns true if the value starts with a known token prefix.
// This catches cases where the key name doesn't indicate sensitivity but the value
// is clearly a token (e.g., "MY_VAR=ghp_abc123").
func ContainsTokenPrefix(value string) bool {
	for _, prefix := range TokenPrefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}
