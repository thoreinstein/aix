// Package git provides Git operation wrappers for cloning and updating repositories.
package git

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/thoreinstein/aix/internal/errors"
)

// scpLikeURL matches git@host:path/to/repo.git style URLs.
var scpLikeURL = regexp.MustCompile(`^[\w-]+@[\w.-]+:[\w./-]+\.git$`)

// ValidateURL checks if the provided string is a safe and valid git URL.
// It allows:
//   - HTTP/HTTPS: https://github.com/user/repo.git
//   - SSH: ssh://git@github.com/user/repo.git
//   - Git: git://github.com/user/repo.git
//   - SCP-like SSH: git@github.com:user/repo.git
//
// It rejects:
//   - Strings starting with "-" (argument injection)
//   - "ext::" protocol (remote code execution risk)
//   - Unknown schemes
func ValidateURL(s string) error {
	if s == "" {
		return errors.New("git URL cannot be empty")
	}

	// Prevent argument injection
	if strings.HasPrefix(s, "-") {
		return errors.Newf("git URL cannot start with '-': %s", s)
	}

	// Block ext:: protocol explicitly (RCE risk)
	if strings.HasPrefix(s, "ext::") {
		return errors.Newf("ext:: protocol is not allowed: %s", s)
	}

	// Check SCP-like syntax (git@host:path)
	if scpLikeURL.MatchString(s) {
		return nil
	}

	// Parse as URL
	u, err := url.Parse(s)
	if err != nil {
		return errors.Wrapf(err, "parsing git URL %s", s)
	}

	// Validate scheme
	switch u.Scheme {
	case "http", "https", "ssh", "git", "file":
		return nil
	case "":
		// If no scheme and not matched by scpLikeURL, assume invalid or local path (which we block for now to be safe)
		// We could allow file:// if needed, but for now stick to remote protocols per requirement.
		return errors.Newf("missing protocol scheme in git URL: %s", s)
	default:
		return errors.Newf("unsupported protocol scheme %q in git URL: %s", u.Scheme, s)
	}
}

// IsURL returns true if s looks like a valid git repository URL.
// This acts as a loose check, but callers should prefer ValidateURL for security.
func IsURL(s string) bool {
	return ValidateURL(s) == nil
}

// Clone clones a git repository from url to dest with the specified depth.
// It validates the URL before execution to prevent injection attacks.
// Output is streamed to os.Stdout and os.Stderr. Stdin is connected to os.Stdin
// to support interactive authentication (e.g., SSH passphrase, credentials).
func Clone(url, dest string, depth int) error {
	if err := ValidateURL(url); err != nil {
		return errors.Wrap(err, "validating git URL")
	}

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
