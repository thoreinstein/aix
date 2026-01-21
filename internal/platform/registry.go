package platform

import (
	"errors"
	"sort"
	"sync"

	"github.com/thoreinstein/aix/internal/paths"
)

// Sentinel errors for registry operations.
var (
	// ErrPlatformNotRegistered is returned when attempting to access
	// a platform that has not been registered.
	ErrPlatformNotRegistered = errors.New("platform not registered")

	// ErrPlatformAlreadyRegistered is returned when attempting to register
	// a platform with a name that is already in use.
	ErrPlatformAlreadyRegistered = errors.New("platform already registered")

	// ErrInvalidPlatformName is returned when attempting to register
	// a platform with an invalid name.
	ErrInvalidPlatformName = errors.New("invalid platform name")

	// ErrNilPlatform is returned when attempting to register a nil platform.
	ErrNilPlatform = errors.New("platform is nil")
)

// Registry manages platform adapter registration and lookup.
// It is safe for concurrent use.
type Registry struct {
	mu        sync.RWMutex
	platforms map[string]Platform
}

// NewRegistry creates a new empty platform registry.
func NewRegistry() *Registry {
	return &Registry{
		platforms: make(map[string]Platform),
	}
}

// Register adds a platform adapter to the registry.
// Returns an error if:
//   - The platform is nil
//   - The platform name is empty or invalid
//   - A platform with the same name is already registered
func (r *Registry) Register(p Platform) error {
	if p == nil {
		return ErrNilPlatform
	}

	name := p.Name()
	if !paths.ValidPlatform(name) {
		return ErrInvalidPlatformName
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.platforms[name]; exists {
		return ErrPlatformAlreadyRegistered
	}

	r.platforms[name] = p
	return nil
}

// Get returns a platform by name.
// Returns nil if not found.
func (r *Registry) Get(name string) Platform {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.platforms[name]
}

// All returns all registered platforms in deterministic order.
// Platforms are sorted alphabetically by name.
func (r *Registry) All() []Platform {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.platforms) == 0 {
		return nil
	}

	names := make([]string, 0, len(r.platforms))
	for name := range r.platforms {
		names = append(names, name)
	}
	sort.Strings(names)

	result := make([]Platform, 0, len(names))
	for _, name := range names {
		result = append(result, r.platforms[name])
	}

	return result
}

// Available returns only platforms that are both registered and installed.
// Uses DetectPlatform() to check installation status.
// Platforms are returned in deterministic order, sorted alphabetically by name.
func (r *Registry) Available() []Platform {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.platforms) == 0 {
		return nil
	}

	names := make([]string, 0, len(r.platforms))
	for name := range r.platforms {
		names = append(names, name)
	}
	sort.Strings(names)

	result := make([]Platform, 0, len(names))
	for _, name := range names {
		detection := DetectPlatform(name)
		if detection != nil && detection.Status == StatusInstalled {
			result = append(result, r.platforms[name])
		}
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

// Names returns the names of all registered platforms in deterministic order.
// Platforms are sorted alphabetically.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.platforms) == 0 {
		return nil
	}

	names := make([]string, 0, len(r.platforms))
	for name := range r.platforms {
		names = append(names, name)
	}
	sort.Strings(names)

	return names
}
