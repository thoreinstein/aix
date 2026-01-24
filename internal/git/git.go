// Package git provides Git operation wrappers for cloning and updating repositories.
package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/errors"
)

// IsURL returns true if s looks like a git repository URL.
// It checks for:
//   - URLs containing "://" (e.g., https://, git://)
//   - URLs ending with ".git"
//   - SSH-style URLs starting with "git@"
func IsURL(s string) bool {
	if strings.Contains(s, "://") {
		return true
	}
	if strings.HasSuffix(s, ".git") {
		return true
	}
	if strings.HasPrefix(s, "git@") {
		return true
	}
	return false
}

// Clone clones a git repository from url to dest with the specified depth.
// Output is streamed to os.Stdout and os.Stderr. Stdin is connected to os.Stdin
// to support interactive authentication (e.g., SSH passphrase, credentials).
func Clone(url, dest string, depth int) error {
	depthArg := fmt.Sprintf("--depth=%d", depth)
	cmd := exec.Command("git", "clone", depthArg, url, dest)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "git clone failed")
	}
	return nil
}

// Pull performs a fast-forward-only pull in the specified repository directory.
// Output is streamed to os.Stdout and os.Stderr. Stdin is connected to os.Stdin
// to support interactive authentication (e.g., SSH passphrase, credentials).
func Pull(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "pull", "--ff-only")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "git pull failed")
	}
	return nil
}

// ValidateRemote checks if repoPath is a valid git repository by verifying
// the existence of a .git directory.
func ValidateRemote(repoPath string) error {
	gitDir := filepath.Join(repoPath, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.Newf("not a git repository: %s", repoPath)
		}
		return errors.Wrap(err, "checking git directory")
	}
	if !info.IsDir() {
		return errors.Newf(".git is not a directory: %s", gitDir)
	}
	return nil
}
