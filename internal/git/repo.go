package git

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	httpsRe = regexp.MustCompile(`https://github\.com/([^/]+)/([^/]+?)(?:\.git)?$`)
	sshRe   = regexp.MustCompile(`git@github\.com:([^/]+)/([^/]+?)(?:\.git)?$`)
)

// ParseOwnerRepo extracts owner and repo name from a git remote URL.
func ParseOwnerRepo(remote string) (owner, repo string, err error) {
	if m := httpsRe.FindStringSubmatch(remote); m != nil {
		return m[1], m[2], nil
	}
	if m := sshRe.FindStringSubmatch(remote); m != nil {
		return m[1], m[2], nil
	}
	return "", "", fmt.Errorf("cannot parse github owner/repo from remote: %q", remote)
}

// RepoRoot returns the root directory of the current git repository.
func RepoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", errors.New("not in a git repository")
	}
	return strings.TrimSpace(string(out)), nil
}

// RemoteURL returns the URL for the given remote (usually "origin").
func RemoteURL(remote string) (string, error) {
	out, err := exec.Command("git", "remote", "get-url", remote).Output()
	if err != nil {
		return "", fmt.Errorf("cannot get remote URL for %q: %w", remote, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// WorktreePath returns the path for a PR's worktree.
func WorktreePath(repoRoot string, prNumber int) string {
	return filepath.Join(repoRoot, ".worktrees", fmt.Sprintf("pr-%d", prNumber))
}

// WorktreeExists reports whether the worktree for the given PR number exists.
func WorktreeExists(repoRoot string, prNumber int) bool {
	path := WorktreePath(repoRoot, prNumber)
	out, err := exec.Command("git", "worktree", "list", "--porcelain").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), path)
}

// CreateWorktree creates a git worktree for the given PR.
func CreateWorktree(repoRoot string, prNumber int) error {
	path := WorktreePath(repoRoot, prNumber)
	ref := fmt.Sprintf("refs/pull/%d/head", prNumber)
	// Fetch the ref first
	if err := exec.Command("git", "fetch", "origin", ref).Run(); err != nil {
		return fmt.Errorf("failed to fetch %s: %w", ref, err)
	}
	if err := exec.Command("git", "worktree", "add", "--detach", path, "FETCH_HEAD").Run(); err != nil {
		return fmt.Errorf("failed to create worktree at %s: %w", path, err)
	}
	return nil
}
