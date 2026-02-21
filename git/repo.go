package git

import (
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	scpLikeRe = regexp.MustCompile(`^(?:[^@]+@)?([^:/]+):([^/]+)/([^/]+?)(?:\.git)?$`)
)

// ParseHostOwnerRepo extracts host, owner and repo name from a git remote URL.
func ParseHostOwnerRepo(remote string) (host, owner, repo string, err error) {
	remote = strings.TrimSpace(remote)
	if m := scpLikeRe.FindStringSubmatch(remote); m != nil {
		return m[1], m[2], m[3], nil
	}

	u, err := url.Parse(remote)
	if err != nil {
		return "", "", "", fmt.Errorf("cannot parse remote URL %q: %w", remote, err)
	}
	if u.Host == "" {
		return "", "", "", fmt.Errorf("cannot parse github host/owner/repo from remote: %q", remote)
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", "", fmt.Errorf("cannot parse github host/owner/repo from remote: %q", remote)
	}
	repo = strings.TrimSuffix(parts[1], ".git")
	if repo == "" {
		return "", "", "", fmt.Errorf("cannot parse github host/owner/repo from remote: %q", remote)
	}
	return u.Hostname(), parts[0], repo, nil
}

// ParseOwnerRepo extracts owner and repo name from a git remote URL.
func ParseOwnerRepo(remote string) (owner, repo string, err error) {
	_, owner, repo, err = ParseHostOwnerRepo(remote)
	return owner, repo, err
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

// RemoveWorktree removes the git worktree for the given PR number.
func RemoveWorktree(repoRoot string, prNumber int) error {
	path := WorktreePath(repoRoot, prNumber)
	if err := exec.Command("git", "worktree", "remove", "--force", path).Run(); err != nil {
		return fmt.Errorf("failed to remove worktree at %s: %w", path, err)
	}
	return nil
}
