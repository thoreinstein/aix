package platform

import (
	"sync"
	"testing"

	"github.com/thoreinstein/aix/internal/paths"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry() returned nil")
	}

	// Should be empty
	if got := r.All(); got != nil {
		t.Errorf("NewRegistry().All() = %v, want nil", got)
	}
}

func TestRegistry_Register_Success(t *testing.T) {
	tests := []struct {
		name     string
		platform string
	}{
		{name: "claude", platform: paths.PlatformClaude},
		{name: "opencode", platform: paths.PlatformOpenCode},
		{name: "codex", platform: paths.PlatformCodex},
		{name: "gemini", platform: paths.PlatformGemini},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRegistry()

			err := r.Register(tt.platform)
			if err != nil {
				t.Errorf("Register(%q) error = %v, want nil", tt.platform, err)
			}

			// Verify it was registered
			if !r.Get(tt.platform) {
				t.Errorf("Get(%q) = false, want true", tt.platform)
			}
		})
	}
}

func TestRegistry_Register_InvalidName(t *testing.T) {
	tests := []struct {
		name     string
		platform string
	}{
		{name: "unknown platform", platform: "unknown"},
		{name: "empty string", platform: ""},
		{name: "case sensitive", platform: "Claude"},
		{name: "typo", platform: "claudde"},
		{name: "with spaces", platform: "claude code"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRegistry()

			err := r.Register(tt.platform)
			if err != ErrInvalidPlatformName {
				t.Errorf("Register(%q) error = %v, want %v", tt.platform, err, ErrInvalidPlatformName)
			}
		})
	}
}

func TestRegistry_Register_AlreadyRegistered(t *testing.T) {
	r := NewRegistry()

	// First registration should succeed
	if err := r.Register(paths.PlatformClaude); err != nil {
		t.Fatalf("First Register() error = %v, want nil", err)
	}

	// Second registration should fail
	err := r.Register(paths.PlatformClaude)
	if err != ErrPlatformAlreadyRegistered {
		t.Errorf("Second Register() error = %v, want %v", err, ErrPlatformAlreadyRegistered)
	}

	// Original should still be registered
	if !r.Get(paths.PlatformClaude) {
		t.Error("Platform no longer registered after duplicate registration attempt")
	}
}

func TestRegistry_Get_Registered(t *testing.T) {
	r := NewRegistry()

	if err := r.Register(paths.PlatformClaude); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if !r.Get(paths.PlatformClaude) {
		t.Errorf("Get(%q) = false, want true", paths.PlatformClaude)
	}
}

func TestRegistry_Get_Unregistered(t *testing.T) {
	r := NewRegistry()

	if r.Get(paths.PlatformClaude) {
		t.Errorf("Get(%q) = true, want false", paths.PlatformClaude)
	}
}

func TestRegistry_Get_InvalidName(t *testing.T) {
	r := NewRegistry()

	// Even with an invalid name, Get should return false (not panic)
	if r.Get("invalid-platform") {
		t.Errorf("Get(%q) = true, want false", "invalid-platform")
	}
}

func TestRegistry_All_DeterministicOrder(t *testing.T) {
	r := NewRegistry()

	// Register in random order
	platforms := []string{
		paths.PlatformGemini,
		paths.PlatformClaude,
		paths.PlatformCodex,
		paths.PlatformOpenCode,
	}

	for _, name := range platforms {
		if err := r.Register(name); err != nil {
			t.Fatalf("Register(%q) error = %v", name, err)
		}
	}

	// Call multiple times and verify same order
	for range 10 {
		all := r.All()
		if len(all) != 4 {
			t.Fatalf("All() returned %d platforms, want 4", len(all))
		}

		// Should be alphabetically sorted
		expected := []string{
			paths.PlatformClaude,
			paths.PlatformCodex,
			paths.PlatformGemini,
			paths.PlatformOpenCode,
		}

		for j, name := range all {
			if name != expected[j] {
				t.Errorf("All()[%d] = %q, want %q", j, name, expected[j])
			}
		}
	}
}

func TestRegistry_All_Empty(t *testing.T) {
	r := NewRegistry()

	got := r.All()
	if got != nil {
		t.Errorf("All() on empty registry = %v, want nil", got)
	}
}

func TestRegistry_Available_FiltersInstalled(t *testing.T) {
	r := NewRegistry()

	// Register all platforms
	for _, name := range paths.Platforms() {
		if err := r.Register(name); err != nil {
			t.Fatalf("Register(%q) error = %v", name, err)
		}
	}

	// Available should only return installed platforms
	available := r.Available()

	// Verify each returned platform is actually installed
	for _, name := range available {
		detection := DetectPlatform(name)
		if detection == nil || detection.Status != StatusInstalled {
			t.Errorf("Available() returned non-installed platform %q", name)
		}
	}
}

func TestRegistry_Available_Empty(t *testing.T) {
	r := NewRegistry()

	got := r.Available()
	if got != nil {
		t.Errorf("Available() on empty registry = %v, want nil", got)
	}
}

func TestRegistry_Available_DeterministicOrder(t *testing.T) {
	r := NewRegistry()

	// Register in random order
	platforms := []string{
		paths.PlatformGemini,
		paths.PlatformClaude,
		paths.PlatformCodex,
		paths.PlatformOpenCode,
	}

	for _, name := range platforms {
		if err := r.Register(name); err != nil {
			t.Fatalf("Register(%q) error = %v", name, err)
		}
	}

	// Call multiple times and verify same order
	first := r.Available()
	for range 5 {
		current := r.Available()
		if len(current) != len(first) {
			t.Fatalf("Available() returned different lengths: %d vs %d", len(current), len(first))
		}

		for j := range current {
			if current[j] != first[j] {
				t.Errorf("Available() order not deterministic at index %d: %q vs %q",
					j, current[j], first[j])
			}
		}
	}
}

func TestRegistry_ConcurrentSafety(t *testing.T) {
	r := NewRegistry()
	platforms := paths.Platforms()

	var wg sync.WaitGroup
	const goroutines = 100

	// Spawn goroutines that register, get, and list concurrently
	for i := range goroutines {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Each goroutine does different operations
			switch idx % 3 {
			case 0:
				// Try to register (may fail with already registered, that's OK)
				name := platforms[idx%len(platforms)]
				_ = r.Register(name)
			case 1:
				// Get by name
				name := platforms[idx%len(platforms)]
				_ = r.Get(name)
			case 2:
				// List all
				_ = r.All()
			}
		}(i)
	}

	wg.Wait()
}

func TestRegistry_ConcurrentRegisterAndGet(t *testing.T) {
	r := NewRegistry()
	name := paths.PlatformClaude

	var wg sync.WaitGroup
	const readers = 50
	const writers = 10

	// Writers try to register
	registerErrors := make(chan error, writers)
	for range writers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			registerErrors <- r.Register(name)
		}()
	}

	// Readers try to get
	for range readers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = r.Get(name)
		}()
	}

	wg.Wait()
	close(registerErrors)

	// Exactly one registration should succeed
	successCount := 0
	for err := range registerErrors {
		if err == nil {
			successCount++
		} else if err != ErrPlatformAlreadyRegistered {
			t.Errorf("Unexpected error: %v", err)
		}
	}

	if successCount != 1 {
		t.Errorf("Expected exactly 1 successful registration, got %d", successCount)
	}

	// Platform should be registered
	if !r.Get(name) {
		t.Error("Platform not registered after concurrent operations")
	}
}

func TestSentinelErrors(t *testing.T) {
	// Verify sentinel errors have meaningful messages
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrPlatformAlreadyRegistered",
			err:  ErrPlatformAlreadyRegistered,
			want: "platform already registered",
		},
		{
			name: "ErrInvalidPlatformName",
			err:  ErrInvalidPlatformName,
			want: "invalid platform name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("%s.Error() = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}
