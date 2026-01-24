package resource

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/thoreinstein/aix/internal/paths"
)

// ErrResourceNotFound is returned when a resource's source path does not exist.
var ErrResourceNotFound = errors.New("resource not found")

// CopyToTemp copies a resource from the repository cache to a temporary directory.
// The caller is responsible for cleanup (e.g., defer os.RemoveAll(tempPath)).
//
// For skills (directory with SKILL.md), the entire directory is copied.
// For commands and agents, the behavior depends on whether the path is a directory
// or a flat file:
//   - Directory-based (e.g., commands/foo with command.md): copies the directory
//   - Flat file (e.g., commands/foo.md): copies just the file
//
// Returns the path to the temporary directory containing the copied resource.
func CopyToTemp(res *Resource) (string, error) {
	return CopyToTempFromCache(res, paths.ReposCacheDir())
}

// CopyToTempFromCache copies a resource from the specified cache directory to a
// temporary directory. This variant allows overriding the cache directory for testing.
//
// For directory-based resources (skills, directory commands, directory agents),
// the original directory name is preserved within the temp directory. This ensures
// that validators can verify the resource name matches its directory name.
// For example, copying "skills/implement" creates "/tmp/aix-install-xyz/implement/".
//
// For flat files, the file is copied directly to the temp directory root.
func CopyToTempFromCache(res *Resource, cacheDir string) (string, error) {
	srcPath := filepath.Join(cacheDir, res.RepoName, res.Path)

	// Check if source exists
	info, err := os.Stat(srcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("%w: %s", ErrResourceNotFound, srcPath)
		}
		return "", fmt.Errorf("checking source path: %w", err)
	}

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "aix-install-*")
	if err != nil {
		return "", fmt.Errorf("creating temp directory: %w", err)
	}

	// Copy based on resource type and source structure
	var copyErr error
	var resultPath string

	if info.IsDir() {
		// Directory-based resource (skill directory, command directory, agent directory)
		// Preserve the original directory name for validation (e.g., skill name must match dir name)
		resourceDir := filepath.Join(tempDir, res.Name)
		if err := os.MkdirAll(resourceDir, 0o755); err != nil {
			_ = os.RemoveAll(tempDir)
			return "", fmt.Errorf("creating resource directory: %w", err)
		}
		copyErr = copyDir(srcPath, resourceDir)
		resultPath = resourceDir
	} else {
		// Flat file (command.md or agent.md directly in the type directory)
		copyErr = copyFile(srcPath, filepath.Join(tempDir, filepath.Base(srcPath)))
		resultPath = tempDir
	}

	if copyErr != nil {
		// Clean up temp directory on copy failure
		_ = os.RemoveAll(tempDir)
		return "", fmt.Errorf("copying resource: %w", copyErr)
	}

	return resultPath, nil
}

// copyDir recursively copies a directory from src to dst.
// dst is expected to already exist.
func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("reading directory %s: %w", src, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Create subdirectory
			if err := os.MkdirAll(dstPath, 0o755); err != nil {
				return fmt.Errorf("creating directory %s: %w", dstPath, err)
			}
			// Recurse into subdirectory
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file from src to dst.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening source file %s: %w", src, err)
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("stating source file %s: %w", src, err)
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("creating destination file %s: %w", dst, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copying content from %s to %s: %w", src, dst, err)
	}

	return nil
}

// IsDirectoryResource returns true if the resource path represents a directory
// (e.g., skills/foo, commands/bar) rather than a flat file (e.g., commands/bar.md).
func IsDirectoryResource(res *Resource) bool {
	// Skills are always directories
	if res.Type == TypeSkill {
		return true
	}

	// MCP servers are always flat files (JSON)
	if res.Type == TypeMCP {
		return false
	}

	// Commands and agents can be either directory-based or flat files
	// Flat files have a .md extension in the path
	return !strings.HasSuffix(res.Path, ".md")
}
