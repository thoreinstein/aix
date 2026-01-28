package install

import (
	"fmt"
	"os"

	"github.com/thoreinstein/aix/internal/cli"
	cliprompt "github.com/thoreinstein/aix/internal/cli/prompt"
	"github.com/thoreinstein/aix/internal/config"
	"github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/internal/git"
	"github.com/thoreinstein/aix/internal/repo"
	"github.com/thoreinstein/aix/internal/resource"
)

// LocalInstaller is a function that installs a resource from a local path.
type LocalInstaller func(sourcePath string, scope cli.Scope) error

// Installer handles shared logic for installing resources from repositories.
type Installer struct {
	resourceType resource.ResourceType
	resourceName string // e.g., "skill", "agent", "MCP server"
	localInstall LocalInstaller
}

// NewInstaller creates a new Installer for a specific resource type.
func NewInstaller(t resource.ResourceType, name string, local LocalInstaller) *Installer {
	return &Installer{
		resourceType: t,
		resourceName: name,
		localInstall: local,
	}
}

// InstallFromRepo installs a resource from a list of matches (usually from repo lookup).
func (i *Installer) InstallFromRepo(name string, matches []resource.Resource, scope cli.Scope) error {
	var selected *resource.Resource

	if len(matches) == 1 {
		selected = &matches[0]
	} else {
		// Multiple matches - prompt user to select
		choice, err := cliprompt.SelectResourceDefault(name, matches)
		if err != nil {
			return errors.Wrap(err, "selecting resource")
		}
		selected = choice
	}

	fmt.Printf("Installing from repository: %s\n", selected.RepoName)
	return i.localInstall(selected.SourcePath(), scope)
}

// InstallAllFromRepo installs all resources of the configured type from a specific repository.
func (i *Installer) InstallAllFromRepo(repoName string, scope cli.Scope) error {
	// 1. Get repo config
	configPath := config.DefaultConfigPath()
	mgr := repo.NewManager(configPath)

	rConfig, err := mgr.Get(repoName)
	if err != nil {
		return errors.Wrapf(err, "getting repository %q", repoName)
	}

	// 2. Scan repo for resources
	scanner := resource.NewScanner()
	resources, err := scanner.ScanRepo(rConfig.Path, rConfig.Name, rConfig.URL)
	if err != nil {
		return errors.Wrapf(err, "scanning repository %q", repoName)
	}

	// 3. Filter for matching resource type
	var matches []resource.Resource
	for _, res := range resources {
		if res.Type == i.resourceType {
			matches = append(matches, res)
		}
	}

	if len(matches) == 0 {
		fmt.Printf("No %ss found in repository %q\n", i.resourceName, repoName)
		return nil
	}

	fmt.Printf("Found %d %ss in repository %q. Installing...\n", len(matches), i.resourceName, repoName)

	// 4. Install each resource
	successCount := 0
	for _, res := range matches {
		fmt.Printf("\nInstalling %s...\n", res.Name)

		// Create a synthetic match list (size 1) for InstallFromRepo logic
		if err := i.InstallFromRepo(res.Name, []resource.Resource{res}, scope); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to install %s: %v\n", res.Name, err)
		} else {
			successCount++
		}
	}

	fmt.Printf("\nSuccessfully installed %d/%d %ss from %q.\n", successCount, len(matches), i.resourceName, repoName)

	if successCount < len(matches) {
		return errors.New(fmt.Sprintf("some %ss failed to install", i.resourceName))
	}

	return nil
}

// InstallFromGit clones a git repository and installs the resource from it.
func (i *Installer) InstallFromGit(url string, scope cli.Scope) error {
	fmt.Println("Cloning repository...")

	// Create temp directory for clone
	prefix := fmt.Sprintf("aix-%s-*", i.resourceName)
	tempDir, err := os.MkdirTemp("", prefix)
	if err != nil {
		return errors.Wrap(err, "creating temp directory")
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to clean up temp dir: %v\n", removeErr)
		}
	}()

	// Clone the repository
	if err := git.Clone(url, tempDir, 1); err != nil {
		return errors.Wrap(err, "cloning repository")
	}

	return i.localInstall(tempDir, scope)
}
