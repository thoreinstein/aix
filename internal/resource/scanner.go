// Package resource provides types and utilities for discovering shareable aix
// resources (skills, commands, agents, and MCP servers) from repositories.
package resource

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/thoreinstein/aix/internal/config"
	"github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/internal/mcp"
	"github.com/thoreinstein/aix/pkg/frontmatter"
)

// Scanner scans cached repositories for resources.
type Scanner struct {
	logger *slog.Logger
}

// NewScanner creates a new Scanner with a default discard logger.
func NewScanner() *Scanner {
	return &Scanner{
		logger: slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelWarn,
		})),
	}
}

// NewScannerWithLogger creates a new Scanner with the given logger.
func NewScannerWithLogger(logger *slog.Logger) *Scanner {
	return &Scanner{logger: logger}
}

// ScanRepo scans a single repository for resources.
// It looks for skills, commands, agents, and MCP servers in their
// respective directories.
func (s *Scanner) ScanRepo(repoPath, repoName, repoURL string) ([]Resource, error) {
	resources := make([]Resource, 0, 4)

	// Scan skills directory
	skillResources, err := s.scanSkills(repoPath, repoName, repoURL)
	if err != nil {
		s.logger.Warn("failed to scan skills directory",
			"repo", repoName,
			"error", err)
	}
	resources = append(resources, skillResources...)

	// Scan commands directory
	cmdResources, err := s.scanCommands(repoPath, repoName, repoURL)
	if err != nil {
		s.logger.Warn("failed to scan commands directory",
			"repo", repoName,
			"error", err)
	}
	resources = append(resources, cmdResources...)

	// Scan agents directory
	agentResources, err := s.scanAgents(repoPath, repoName, repoURL)
	if err != nil {
		s.logger.Warn("failed to scan agents directory",
			"repo", repoName,
			"error", err)
	}
	resources = append(resources, agentResources...)

	// Scan MCP directory
	mcpResources, err := s.scanMCP(repoPath, repoName, repoURL)
	if err != nil {
		s.logger.Warn("failed to scan mcp directory",
			"repo", repoName,
			"error", err)
	}
	resources = append(resources, mcpResources...)

	return resources, nil
}

// ScanAll scans multiple repositories for resources concurrently.
// It uses a worker pool limited to GOMAXPROCS to parallelize scanning.
func (s *Scanner) ScanAll(repos []config.RepoConfig) ([]Resource, error) {
	if len(repos) == 0 {
		return nil, nil
	}

	// Limit concurrency to GOMAXPROCS or number of repos, whichever is smaller
	workers := runtime.GOMAXPROCS(0)
	if len(repos) < workers {
		workers = len(repos)
	}

	// Channel to send work to workers
	work := make(chan config.RepoConfig, len(repos))

	// Collect results from workers
	type scanResult struct {
		resources []Resource
		repoName  string
	}
	results := make(chan scanResult, len(repos))

	// Start workers
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for repo := range work {
				repoResources, err := s.ScanRepo(repo.Path, repo.Name, repo.URL)
				if err != nil {
					s.logger.Warn("failed to scan repository",
						"repo", repo.Name,
						"path", repo.Path,
						"error", err)
					results <- scanResult{repoName: repo.Name}
					continue
				}
				results <- scanResult{resources: repoResources, repoName: repo.Name}
			}
		}()
	}

	// Send work to workers
	for _, repo := range repos {
		work <- repo
	}
	close(work)

	// Close results channel when all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect all results
	var resources []Resource
	for result := range results {
		resources = append(resources, result.resources...)
	}

	return resources, nil
}

// skillMeta holds the frontmatter fields we extract from skills.
type skillMeta struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// scanSkills scans the skills/ directory for SKILL.md files.
// Skills use required frontmatter.
func (s *Scanner) scanSkills(repoPath, repoName, repoURL string) ([]Resource, error) {
	skillsDir := filepath.Join(repoPath, "skills")

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		if os.IsPermission(err) {
			s.logger.Warn("permission denied reading skills directory",
				"path", skillsDir,
				"error", err)
			return nil, nil
		}
		return nil, errors.Wrapf(err, "reading skills directory %s", skillsDir)
	}

	resources := make([]Resource, 0, len(entries))

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillPath := filepath.Join(skillsDir, entry.Name(), "SKILL.md")
		file, err := os.Open(skillPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			if os.IsPermission(err) {
				s.logger.Warn("permission denied reading skill file",
					"path", skillPath)
				continue
			}
			s.logger.Warn("failed to open skill file",
				"path", skillPath,
				"error", err)
			continue
		}

		var meta skillMeta
		if err := frontmatter.ParseHeader(file, &meta); err != nil {
			file.Close()
			s.logger.Warn("failed to parse skill frontmatter",
				"path", skillPath,
				"error", err)
			continue
		}
		file.Close()

		// Use directory name if name not in frontmatter
		name := meta.Name
		if name == "" {
			name = entry.Name()
		}

		resources = append(resources, Resource{
			Name:        name,
			Description: meta.Description,
			Type:        TypeSkill,
			RepoName:    repoName,
			RepoURL:     repoURL,
			Path:        filepath.Join("skills", entry.Name()),
		})
	}

	return resources, nil
}

// commandMeta holds the frontmatter fields we extract from commands.
type commandMeta struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// scanCommands scans the commands/ directory for command.md or *.md files.
// Commands use optional frontmatter.
func (s *Scanner) scanCommands(repoPath, repoName, repoURL string) ([]Resource, error) {
	commandsDir := filepath.Join(repoPath, "commands")

	entries, err := os.ReadDir(commandsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		if os.IsPermission(err) {
			s.logger.Warn("permission denied reading commands directory",
				"path", commandsDir,
				"error", err)
			return nil, nil
		}
		return nil, errors.Wrapf(err, "reading commands directory %s", commandsDir)
	}

	var resources []Resource

	for _, entry := range entries {
		if entry.IsDir() {
			// Look for command.md in subdirectory
			resource, err := s.scanCommandDir(commandsDir, entry.Name(), repoName, repoURL)
			if err != nil {
				s.logger.Warn("failed to scan command directory",
					"dir", entry.Name(),
					"error", err)
				continue
			}
			if resource != nil {
				resources = append(resources, *resource)
			}
		} else if strings.HasSuffix(entry.Name(), ".md") {
			// Direct .md file in commands/
			resource, err := s.scanCommandFile(commandsDir, entry.Name(), repoName, repoURL)
			if err != nil {
				s.logger.Warn("failed to scan command file",
					"file", entry.Name(),
					"error", err)
				continue
			}
			if resource != nil {
				resources = append(resources, *resource)
			}
		}
	}

	return resources, nil
}

// scanCommandDir scans a subdirectory for command.md.
func (s *Scanner) scanCommandDir(commandsDir, dirName, repoName, repoURL string) (*Resource, error) {
	cmdPath := filepath.Join(commandsDir, dirName, "command.md")
	file, err := os.Open(cmdPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "opening command file %s", cmdPath)
	}
	defer file.Close()

	var meta commandMeta
	if err := frontmatter.ParseHeader(file, &meta); err != nil {
		return nil, errors.Wrap(err, "parsing command frontmatter")
	}

	name := meta.Name
	if name == "" {
		name = dirName
	}

	return &Resource{
		Name:        name,
		Description: meta.Description,
		Type:        TypeCommand,
		RepoName:    repoName,
		RepoURL:     repoURL,
		Path:        filepath.Join("commands", dirName),
	}, nil
}

// scanCommandFile scans a direct .md file in the commands directory.
func (s *Scanner) scanCommandFile(commandsDir, fileName, repoName, repoURL string) (*Resource, error) {
	cmdPath := filepath.Join(commandsDir, fileName)
	file, err := os.Open(cmdPath)
	if err != nil {
		return nil, errors.Wrapf(err, "opening command file %s", cmdPath)
	}
	defer file.Close()

	var meta commandMeta
	if err := frontmatter.ParseHeader(file, &meta); err != nil {
		return nil, errors.Wrap(err, "parsing command frontmatter")
	}

	// Derive name from filename (strip .md extension)
	name := meta.Name
	if name == "" {
		name = strings.TrimSuffix(fileName, ".md")
	}

	return &Resource{
		Name:        name,
		Description: meta.Description,
		Type:        TypeCommand,
		RepoName:    repoName,
		RepoURL:     repoURL,
		Path:        filepath.Join("commands", fileName),
	}, nil
}

// agentMeta holds the frontmatter fields we extract from agents.
type agentMeta struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// scanAgents scans the agents/ directory for AGENT.md or *.md files.
// Agents use optional frontmatter.
func (s *Scanner) scanAgents(repoPath, repoName, repoURL string) ([]Resource, error) {
	agentsDir := filepath.Join(repoPath, "agents")

	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		if os.IsPermission(err) {
			s.logger.Warn("permission denied reading agents directory",
				"path", agentsDir,
				"error", err)
			return nil, nil
		}
		return nil, errors.Wrapf(err, "reading agents directory %s", agentsDir)
	}

	var resources []Resource

	for _, entry := range entries {
		if entry.IsDir() {
			// Look for AGENT.md in subdirectory
			resource, err := s.scanAgentDir(agentsDir, entry.Name(), repoName, repoURL)
			if err != nil {
				s.logger.Warn("failed to scan agent directory",
					"dir", entry.Name(),
					"error", err)
				continue
			}
			if resource != nil {
				resources = append(resources, *resource)
			}
		} else if strings.HasSuffix(entry.Name(), ".md") {
			// Direct .md file in agents/
			resource, err := s.scanAgentFile(agentsDir, entry.Name(), repoName, repoURL)
			if err != nil {
				s.logger.Warn("failed to scan agent file",
					"file", entry.Name(),
					"error", err)
				continue
			}
			if resource != nil {
				resources = append(resources, *resource)
			}
		}
	}

	return resources, nil
}

// scanAgentDir scans a subdirectory for AGENT.md.
func (s *Scanner) scanAgentDir(agentsDir, dirName, repoName, repoURL string) (*Resource, error) {
	agentPath := filepath.Join(agentsDir, dirName, "AGENT.md")
	file, err := os.Open(agentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "opening agent file %s", agentPath)
	}
	defer file.Close()

	var meta agentMeta
	if err := frontmatter.ParseHeader(file, &meta); err != nil {
		return nil, errors.Wrap(err, "parsing agent frontmatter")
	}

	name := meta.Name
	if name == "" {
		name = dirName
	}

	return &Resource{
		Name:        name,
		Description: meta.Description,
		Type:        TypeAgent,
		RepoName:    repoName,
		RepoURL:     repoURL,
		Path:        filepath.Join("agents", dirName),
	}, nil
}

// scanAgentFile scans a direct .md file in the agents directory.
func (s *Scanner) scanAgentFile(agentsDir, fileName, repoName, repoURL string) (*Resource, error) {
	agentPath := filepath.Join(agentsDir, fileName)
	file, err := os.Open(agentPath)
	if err != nil {
		return nil, errors.Wrapf(err, "opening agent file %s", agentPath)
	}
	defer file.Close()

	var meta agentMeta
	if err := frontmatter.ParseHeader(file, &meta); err != nil {
		return nil, errors.Wrap(err, "parsing agent frontmatter")
	}

	// Derive name from filename (strip .md extension)
	name := meta.Name
	if name == "" {
		name = strings.TrimSuffix(fileName, ".md")
	}

	return &Resource{
		Name:        name,
		Description: meta.Description,
		Type:        TypeAgent,
		RepoName:    repoName,
		RepoURL:     repoURL,
		Path:        filepath.Join("agents", fileName),
	}, nil
}

// scanMCP scans the mcp/ directory for *.json files.
func (s *Scanner) scanMCP(repoPath, repoName, repoURL string) ([]Resource, error) {
	mcpDir := filepath.Join(repoPath, "mcp")

	entries, err := os.ReadDir(mcpDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		if os.IsPermission(err) {
			s.logger.Warn("permission denied reading mcp directory",
				"path", mcpDir,
				"error", err)
			return nil, nil
		}
		return nil, errors.Wrapf(err, "reading mcp directory %s", mcpDir)
	}

	resources := make([]Resource, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		mcpPath := filepath.Join(mcpDir, entry.Name())
		data, err := os.ReadFile(mcpPath)
		if err != nil {
			s.logger.Warn("failed to read MCP file",
				"path", mcpPath,
				"error", err)
			continue
		}

		var server mcp.Server
		if err := json.Unmarshal(data, &server); err != nil {
			s.logger.Warn("failed to parse MCP JSON",
				"path", mcpPath,
				"error", err)
			continue
		}

		// Derive name from filename (strip .json extension) if not in file
		name := server.Name
		if name == "" {
			name = strings.TrimSuffix(entry.Name(), ".json")
		}

		// Build description from server properties
		description := buildMCPDescription(&server)

		resources = append(resources, Resource{
			Name:        name,
			Description: description,
			Type:        TypeMCP,
			RepoName:    repoName,
			RepoURL:     repoURL,
			Path:        filepath.Join("mcp", entry.Name()),
		})
	}

	return resources, nil
}

// buildMCPDescription creates a human-readable description from server config.
func buildMCPDescription(server *mcp.Server) string {
	if server.IsLocal() {
		return "Local MCP server: " + server.Command
	}
	if server.IsRemote() {
		return "Remote MCP server: " + server.URL
	}
	return "MCP server"
}
