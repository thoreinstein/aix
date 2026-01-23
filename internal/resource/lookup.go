package resource

import (
	"errors"
	"path/filepath"

	"github.com/thoreinstein/aix/internal/config"
	"github.com/thoreinstein/aix/internal/paths"
	"github.com/thoreinstein/aix/internal/repo"
)

// ErrNoReposConfigured is returned when no repositories are configured.
var ErrNoReposConfigured = errors.New("no repositories configured")

// FindByName scans all configured repositories and returns resources matching
// the given name and type exactly. Returns an empty slice if no matches found.
func FindByName(name string, resourceType ResourceType) ([]Resource, error) {
	configPath := filepath.Join(paths.ConfigHome(), config.AppName, "config.yaml")
	mgr := repo.NewManager(configPath)

	repos, err := mgr.List()
	if err != nil {
		return nil, err
	}

	if len(repos) == 0 {
		return nil, ErrNoReposConfigured
	}

	scanner := NewScanner()
	resources, err := scanner.ScanAll(repos)
	if err != nil {
		return nil, err
	}

	return filterByNameAndType(resources, name, resourceType), nil
}

// FindByNameInRepo scans a specific repository and returns the resource matching
// the given name and type exactly. Returns nil if no match found.
func FindByNameInRepo(name string, resourceType ResourceType, repoName string) (*Resource, error) {
	configPath := filepath.Join(paths.ConfigHome(), config.AppName, "config.yaml")
	mgr := repo.NewManager(configPath)

	repoConfig, err := mgr.Get(repoName)
	if err != nil {
		return nil, err
	}

	scanner := NewScanner()
	resources, err := scanner.ScanRepo(repoConfig.Path, repoConfig.Name, repoConfig.URL)
	if err != nil {
		return nil, err
	}

	matches := filterByNameAndType(resources, name, resourceType)
	if len(matches) == 0 {
		return nil, nil
	}

	return &matches[0], nil
}

// filterByNameAndType filters resources by exact name and type match.
func filterByNameAndType(resources []Resource, name string, resourceType ResourceType) []Resource {
	var matches []Resource
	for _, r := range resources {
		if r.Name == name && r.Type == resourceType {
			matches = append(matches, r)
		}
	}
	return matches
}
