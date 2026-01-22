package doctor

import (
	"testing"
)

func TestPlatformCheck_Name(t *testing.T) {
	c := NewPlatformCheck()
	if got := c.Name(); got != "platform-detection" {
		t.Errorf("Name() = %q, want %q", got, "platform-detection")
	}
}

func TestPlatformCheck_Category(t *testing.T) {
	c := NewPlatformCheck()
	if got := c.Category(); got != "platform" {
		t.Errorf("Category() = %q, want %q", got, "platform")
	}
}

func TestPlatformCheck_ImplementsCheck(t *testing.T) {
	// Compile-time check that PlatformCheck implements Check interface
	var _ Check = (*PlatformCheck)(nil)
}

func TestPlatformCheck_Run(t *testing.T) {
	c := NewPlatformCheck()
	result := c.Run()

	// Basic validation of result structure
	if result == nil {
		t.Fatal("Run() returned nil")
	}

	if result.Name != "platform-detection" {
		t.Errorf("result.Name = %q, want %q", result.Name, "platform-detection")
	}

	if result.Category != "platform" {
		t.Errorf("result.Category = %q, want %q", result.Category, "platform")
	}

	// Status should be Pass, Warning, or Info (never Error for platform detection)
	switch result.Status {
	case SeverityPass, SeverityWarning, SeverityInfo:
		// Valid statuses for platform detection
	case SeverityError:
		t.Error("platform detection should not return SeverityError")
	default:
		t.Errorf("unexpected status: %v", result.Status)
	}

	// Details should contain platform information
	if result.Details == nil {
		t.Error("result.Details is nil, expected platform information")
	}

	// Check that details contains expected fields
	if _, ok := result.Details["platforms"]; !ok {
		t.Error("result.Details missing 'platforms' key")
	}
	if _, ok := result.Details["installed"]; !ok {
		t.Error("result.Details missing 'installed' key")
	}
	if _, ok := result.Details["not_installed"]; !ok {
		t.Error("result.Details missing 'not_installed' key")
	}
	if _, ok := result.Details["total"]; !ok {
		t.Error("result.Details missing 'total' key")
	}
}

func TestPlatformCheck_Run_Details(t *testing.T) {
	c := NewPlatformCheck()
	result := c.Run()

	platforms, ok := result.Details["platforms"].(map[string]any)
	if !ok {
		t.Fatal("result.Details['platforms'] is not a map[string]any")
	}

	// Should have entries for known platforms
	expectedPlatforms := []string{"claude", "opencode", "codex", "gemini"}
	for _, name := range expectedPlatforms {
		info, exists := platforms[name]
		if !exists {
			t.Errorf("missing platform info for %q", name)
			continue
		}

		infoMap, ok := info.(map[string]any)
		if !ok {
			t.Errorf("platform info for %q is not a map[string]any", name)
			continue
		}

		// Each platform info should have status, global_config, mcp_config
		if _, ok := infoMap["status"]; !ok {
			t.Errorf("platform %q info missing 'status' key", name)
		}
		if _, ok := infoMap["global_config"]; !ok {
			t.Errorf("platform %q info missing 'global_config' key", name)
		}
		if _, ok := infoMap["mcp_config"]; !ok {
			t.Errorf("platform %q info missing 'mcp_config' key", name)
		}
	}
}

func TestPlatformCheck_Run_CountConsistency(t *testing.T) {
	c := NewPlatformCheck()
	result := c.Run()

	installed, _ := result.Details["installed"].(int)
	notInstalled, _ := result.Details["not_installed"].(int)
	partial, _ := result.Details["partial"].(int)
	total, _ := result.Details["total"].(int)

	sum := installed + notInstalled + partial
	if sum != total {
		t.Errorf("count mismatch: installed(%d) + not_installed(%d) + partial(%d) = %d, but total = %d",
			installed, notInstalled, partial, sum, total)
	}
}
