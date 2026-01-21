package platform

import (
	"sync"
	"testing"

	"github.com/thoreinstein/aix/internal/paths"
)

// mockPlatform is a test implementation of the Platform interface.
type mockPlatform struct {
	name            string
	globalConfigDir string
	mcpConfigPath   string
	instructionFile string
}

func (m *mockPlatform) Name() string                { return m.name }
func (m *mockPlatform) GlobalConfigDir() string     { return m.globalConfigDir }
func (m *mockPlatform) MCPConfigPath() string       { return m.mcpConfigPath }
func (m *mockPlatform) InstructionFilename() string { return m.instructionFile }

// newMockPlatform creates a mock platform with the given name.
// Uses paths package to get realistic config values.
func newMockPlatform(name string) *mockPlatform {
	return &mockPlatform{
		name:            name,
		globalConfigDir: paths.GlobalConfigDir(name),
		mcpConfigPath:   paths.MCPConfigPath(name),
		instructionFile: paths.InstructionFilename(name),
	}
}

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry() returned nil")
	}

	// Should be empty
	if got := r.All(); got != nil {
		t.Errorf("NewRegistry().All() = %v, want nil", got)
	}

	if got := r.Names(); got != nil {
		t.Errorf("NewRegistry().Names() = %v, want nil", got)
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
			p := newMockPlatform(tt.platform)

			err := r.Register(p)
			if err != nil {
				t.Errorf("Register(%q) error = %v, want nil", tt.platform, err)
			}

			// Verify it was registered
			got := r.Get(tt.platform)
			if got != p {
				t.Errorf("Get(%q) = %v, want %v", tt.platform, got, p)
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
			p := &mockPlatform{name: tt.platform}

			err := r.Register(p)
			if err != ErrInvalidPlatformName {
				t.Errorf("Register(%q) error = %v, want %v", tt.platform, err, ErrInvalidPlatformName)
			}
		})
	}
}

func TestRegistry_Register_AlreadyRegistered(t *testing.T) {
	r := NewRegistry()
	p1 := newMockPlatform(paths.PlatformClaude)
	p2 := newMockPlatform(paths.PlatformClaude)

	// First registration should succeed
	if err := r.Register(p1); err != nil {
		t.Fatalf("First Register() error = %v, want nil", err)
	}

	// Second registration should fail
	err := r.Register(p2)
	if err != ErrPlatformAlreadyRegistered {
		t.Errorf("Second Register() error = %v, want %v", err, ErrPlatformAlreadyRegistered)
	}

	// Original should still be registered
	if got := r.Get(paths.PlatformClaude); got != p1 {
		t.Error("Original platform was overwritten")
	}
}

func TestRegistry_Register_NilPlatform(t *testing.T) {
	r := NewRegistry()

	err := r.Register(nil)
	if err != ErrNilPlatform {
		t.Errorf("Register(nil) error = %v, want %v", err, ErrNilPlatform)
	}
}

func TestRegistry_Get_Registered(t *testing.T) {
	r := NewRegistry()
	p := newMockPlatform(paths.PlatformClaude)

	if err := r.Register(p); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	got := r.Get(paths.PlatformClaude)
	if got != p {
		t.Errorf("Get(%q) = %v, want %v", paths.PlatformClaude, got, p)
	}
}

func TestRegistry_Get_Unregistered(t *testing.T) {
	r := NewRegistry()

	got := r.Get(paths.PlatformClaude)
	if got != nil {
		t.Errorf("Get(%q) = %v, want nil", paths.PlatformClaude, got)
	}
}

func TestRegistry_Get_InvalidName(t *testing.T) {
	r := NewRegistry()

	// Even with an invalid name, Get should return nil (not panic)
	got := r.Get("invalid-platform")
	if got != nil {
		t.Errorf("Get(%q) = %v, want nil", "invalid-platform", got)
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
		if err := r.Register(newMockPlatform(name)); err != nil {
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

		for j, p := range all {
			if p.Name() != expected[j] {
				t.Errorf("All()[%d].Name() = %q, want %q", j, p.Name(), expected[j])
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
		if err := r.Register(newMockPlatform(name)); err != nil {
			t.Fatalf("Register(%q) error = %v", name, err)
		}
	}

	// Available should only return installed platforms
	available := r.Available()

	// Verify each returned platform is actually installed
	for _, p := range available {
		detection := DetectPlatform(p.Name())
		if detection == nil || detection.Status != StatusInstalled {
			t.Errorf("Available() returned non-installed platform %q", p.Name())
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
		if err := r.Register(newMockPlatform(name)); err != nil {
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
			if current[j].Name() != first[j].Name() {
				t.Errorf("Available() order not deterministic at index %d: %q vs %q",
					j, current[j].Name(), first[j].Name())
			}
		}
	}
}

func TestRegistry_Names_DeterministicOrder(t *testing.T) {
	r := NewRegistry()

	// Register in random order
	platforms := []string{
		paths.PlatformCodex,
		paths.PlatformOpenCode,
		paths.PlatformClaude,
		paths.PlatformGemini,
	}

	for _, name := range platforms {
		if err := r.Register(newMockPlatform(name)); err != nil {
			t.Fatalf("Register(%q) error = %v", name, err)
		}
	}

	// Should be alphabetically sorted
	expected := []string{
		paths.PlatformClaude,
		paths.PlatformCodex,
		paths.PlatformGemini,
		paths.PlatformOpenCode,
	}

	names := r.Names()
	if len(names) != len(expected) {
		t.Fatalf("Names() returned %d names, want %d", len(names), len(expected))
	}

	for i, name := range names {
		if name != expected[i] {
			t.Errorf("Names()[%d] = %q, want %q", i, name, expected[i])
		}
	}
}

func TestRegistry_Names_Empty(t *testing.T) {
	r := NewRegistry()

	got := r.Names()
	if got != nil {
		t.Errorf("Names() on empty registry = %v, want nil", got)
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
			switch idx % 4 {
			case 0:
				// Try to register (may fail with already registered, that's OK)
				name := platforms[idx%len(platforms)]
				_ = r.Register(newMockPlatform(name))
			case 1:
				// Get by name
				name := platforms[idx%len(platforms)]
				_ = r.Get(name)
			case 2:
				// List all
				_ = r.All()
			case 3:
				// List names
				_ = r.Names()
			}
		}(i)
	}

	wg.Wait()
}

func TestRegistry_ConcurrentRegisterAndGet(t *testing.T) {
	r := NewRegistry()
	name := paths.PlatformClaude
	p := newMockPlatform(name)

	var wg sync.WaitGroup
	const readers = 50
	const writers = 10

	// Writers try to register
	registerErrors := make(chan error, writers)
	for range writers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			registerErrors <- r.Register(p)
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
	if got := r.Get(name); got == nil {
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
			name: "ErrPlatformNotRegistered",
			err:  ErrPlatformNotRegistered,
			want: "platform not registered",
		},
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
		{
			name: "ErrNilPlatform",
			err:  ErrNilPlatform,
			want: "platform is nil",
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
