package platform

import (
	"sync"

	"github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/internal/paths"
)

// Sentinel errors for registry operations.
var (
	// ErrPlatformAlreadyRegistered is returned when attempting to register
	// a platform with a name that is already in use.
	ErrPlatformAlreadyRegistered = errors.New("platform already registered")

	// ErrInvalidPlatformName is returned when attempting to register
	// a platform with an invalid name.
	ErrInvalidPlatformName = errors.New("invalid platform name")
)

// Registry manages platform name registration and lookup.
// It is safe for concurrent use.
type Registry struct {
	mu        sync.RWMutex
	platforms map[string]struct{}
}

// NewRegistry creates a new empty platform registry.
func NewRegistry() *Registry {
	return &Registry{
		platforms: make(map[string]struct{}),
	}
}

// Register adds a platform name to the registry.
// Returns an error if:
//   - The platform name is empty or invalid (per paths.ValidPlatform)
//   - A platform with the same name is already registered
func (r *Registry) Register(name string) error {
	if !paths.ValidPlatform(name) {
		return ErrInvalidPlatformName
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.platforms[name]; exists {
		return ErrPlatformAlreadyRegistered
	}

	r.platforms[name] = struct{}{}
	return nil
}

// Get returns true if the platform name is registered.
func (r *Registry) Get(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.platforms[name]
	return exists
}

// All returns all registered platform names in deterministic order
// defined in paths.Platforms().
func (r *Registry) All() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	all := paths.Platforms()
	results := make([]string, 0, len(r.platforms))

	for _, name := range all {
		if _, registered := r.platforms[name]; registered {
			results = append(results, name)
		}
	}

	return results
}

// Available returns only registered platforms that are installed.
// Uses DetectPlatform() to check installation status.
// Platforms are returned in deterministic order defined in paths.Platforms().
func (r *Registry) Available() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	all := paths.Platforms()
	results := make([]string, 0, len(r.platforms))

	for _, name := range all {
		if _, registered := r.platforms[name]; registered {
			detection := DetectPlatform(name)
			if detection != nil && detection.Status == StatusInstalled {
				results = append(results, name)
			}
		}
	}

	return results
}
