// Package repo provides repository management for skill repositories.
package repo

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/thoreinstein/aix/internal/config"
	"github.com/thoreinstein/aix/internal/git"
	"github.com/thoreinstein/aix/internal/paths"
	"github.com/thoreinstein/aix/pkg/fileutil"
)

// Sentinel errors for repository operations.
var (
	ErrNotFound           = errors.New("repository not found")
	ErrInvalidURL         = errors.New("invalid git URL")
	ErrNameCollision      = errors.New("repository with this name already exists")
	ErrInvalidName        = errors.New("invalid repository name")
	ErrCacheCleanupFailed = errors.New("cache cleanup failed")
)

// namePattern validates repository names.
// Names must be lowercase alphanumeric with hyphens, starting with a letter.
var namePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

// Option configures Add behavior.
type Option func(*addOptions)

// addOptions holds optional parameters for Add.
type addOptions struct {
	name string
}

// WithName overrides the repository name derived from the URL.
func WithName(name string) Option {
	return func(o *addOptions) {
		o.name = name
	}
}

// Manager manages skill repositories.
type Manager struct {
	configPath string // Path to config file for persistence
}

// NewManager creates a new repository manager.
// The configPath specifies where the config file is stored.
func NewManager(configPath string) *Manager {
	return &Manager{configPath: configPath}
}

// Add clones a repository and registers it in the config.
// Returns the created RepoConfig or an error.
func (m *Manager) Add(url string, opts ...Option) (*config.RepoConfig, error) {
	// Apply options
	var options addOptions
	for _, opt := range opts {
		opt(&options)
	}

	// Validate URL
	if !git.IsURL(url) {
		return nil, errors.WithDetail(ErrInvalidURL, url)
	}

	// Derive name from URL if not provided
	name := options.name
	if name == "" {
		name = deriveNameFromURL(url)
	}

	// Validate name
	if !namePattern.MatchString(name) {
		return nil, errors.WithDetailf(ErrInvalidName, "name %q must be lowercase alphanumeric with hyphens, starting with a letter", name)
	}

	// Load config to check for collisions
	cfg, err := m.loadConfig()
	if err != nil {
		return nil, errors.Wrap(err, "loading config")
	}

	// Check for name collision
	if cfg.Repos != nil {
		if existing, exists := cfg.Repos[name]; exists {
			return nil, errors.WithDetailf(ErrNameCollision,
				"name %q is already used by %s; use --name to specify an alternate name",
				name, existing.URL)
		}
	}

	// Create cache directory
	cacheDir := paths.ReposCacheDir()
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, errors.Wrap(err, "creating cache directory")
	}

	// Build destination path
	destPath := filepath.Join(cacheDir, name)

	// Clone repository - clean up partial clone on failure
	if err := git.Clone(url, destPath, 1); err != nil {
		// Remove any partially-created directory
		if cleanupErr := os.RemoveAll(destPath); cleanupErr != nil {
			return nil, errors.Wrapf(err, "cloning repository (cleanup also failed: %v)", cleanupErr)
		}
		return nil, errors.Wrap(err, "cloning repository")
	}

	// Create repo config entry
	repo := config.RepoConfig{
		URL:     url,
		Name:    name,
		Path:    destPath,
		AddedAt: time.Now(),
	}

	// Initialize repos map if nil
	if cfg.Repos == nil {
		cfg.Repos = make(map[string]config.RepoConfig)
	}

	// Add repo to config
	cfg.Repos[name] = repo

	// Save config
	if err := m.saveConfig(cfg); err != nil {
		// Clean up cloned repo on save failure
		os.RemoveAll(destPath)
		return nil, errors.Wrap(err, "saving config")
	}

	return &repo, nil
}

// List returns all registered repositories.
// Returns an empty slice if no repositories are registered.
func (m *Manager) List() ([]config.RepoConfig, error) {
	cfg, err := m.loadConfig()
	if err != nil {
		return nil, errors.Wrap(err, "loading config")
	}

	if cfg.Repos == nil {
		return []config.RepoConfig{}, nil
	}

	repos := make([]config.RepoConfig, 0, len(cfg.Repos))
	for _, repo := range cfg.Repos {
		repos = append(repos, repo)
	}

	return repos, nil
}

// Remove unregisters a repository and deletes its cached clone.
// The config is persisted before deleting cached data to ensure
// consistent state if the operation fails partway through.
func (m *Manager) Remove(name string) error {
	cfg, err := m.loadConfig()
	if err != nil {
		return errors.Wrap(err, "loading config")
	}

	if cfg.Repos == nil {
		return errors.WithDetailf(ErrNotFound, "repository %q not found", name)
	}

	repo, exists := cfg.Repos[name]
	if !exists {
		return errors.WithDetailf(ErrNotFound, "repository %q not found", name)
	}

	// Remove from config first
	delete(cfg.Repos, name)

	// Persist config before deleting data - if this fails, cached data remains intact
	if err := m.saveConfig(cfg); err != nil {
		return errors.Wrap(err, "saving config")
	}

	// Remove cached directory - if this fails, log warning but don't fail
	// The config is already updated, so the repo is "removed" from aix's perspective
	if err := os.RemoveAll(repo.Path); err != nil {
		// Return a wrapped error that indicates partial success
		return errors.Wrapf(ErrCacheCleanupFailed, "config updated but failed to remove cached directory %q: %v", repo.Path, err)
	}

	return nil
}

// Update pulls the latest changes for repositories.
// If name is provided, only that repository is updated.
// If name is empty, all repositories are updated.
func (m *Manager) Update(name string) error {
	cfg, err := m.loadConfig()
	if err != nil {
		return errors.Wrap(err, "loading config")
	}

	if len(cfg.Repos) == 0 {
		if name != "" {
			return errors.WithDetailf(ErrNotFound, "repository %q not found", name)
		}
		return nil // No repos to update
	}

	// Update specific repo
	if name != "" {
		repo, exists := cfg.Repos[name]
		if !exists {
			return errors.WithDetailf(ErrNotFound, "repository %q not found", name)
		}
		return git.Pull(repo.Path)
	}

	// Update all repos - return first error encountered
	for _, repo := range cfg.Repos {
		if err := git.Pull(repo.Path); err != nil {
			return errors.Wrapf(err, "updating repository %q", repo.Name)
		}
	}

	return nil
}

// UpdateByPath pulls the latest changes for a repository at the given path.
// This is more efficient when you already have the repo config and don't need
// to reload configuration.
func (m *Manager) UpdateByPath(path string) error {
	return git.Pull(path)
}

// Get retrieves a repository by name.
func (m *Manager) Get(name string) (*config.RepoConfig, error) {
	cfg, err := m.loadConfig()
	if err != nil {
		return nil, errors.Wrap(err, "loading config")
	}

	if cfg.Repos == nil {
		return nil, errors.WithDetailf(ErrNotFound, "repository %q not found", name)
	}

	repo, exists := cfg.Repos[name]
	if !exists {
		return nil, errors.WithDetailf(ErrNotFound, "repository %q not found", name)
	}

	return &repo, nil
}

// deriveNameFromURL extracts a repository name from a git URL.
// It takes the last path segment and strips the .git suffix if present.
func deriveNameFromURL(url string) string {
	// Handle SSH URLs (git@github.com:user/repo.git)
	if strings.HasPrefix(url, "git@") {
		if colonIdx := strings.LastIndex(url, ":"); colonIdx != -1 {
			url = url[colonIdx+1:]
		}
	}

	// Get the last path segment
	name := filepath.Base(url)

	// Strip .git suffix
	name = strings.TrimSuffix(name, ".git")

	// Convert to lowercase
	name = strings.ToLower(name)

	return name
}

// loadConfig loads the configuration from the manager's config path.
// If the config file doesn't exist, it returns a default config.
func (m *Manager) loadConfig() (*config.Config, error) {
	// Check if config file exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		// Return default config
		return &config.Config{
			Version:          1,
			DefaultPlatforms: paths.Platforms(),
			Repos:            make(map[string]config.RepoConfig),
		}, nil
	}

	// Initialize viper with defaults
	config.Init()

	// Load from the specified path
	return config.Load(m.configPath)
}

// saveConfig saves the configuration to the manager's config path.
func (m *Manager) saveConfig(cfg *config.Config) error {
	// Ensure parent directory exists
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return errors.Wrap(err, "creating config directory")
	}

	return fileutil.AtomicWriteYAML(m.configPath, cfg)
}
