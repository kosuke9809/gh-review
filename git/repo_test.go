package git_test

import (
	"testing"

	"github.com/kosuke9809/gh-review/git"
)

func TestParseOwnerRepo(t *testing.T) {
	tests := []struct {
		remote    string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{"https://github.com/myorg/myrepo.git", "myorg", "myrepo", false},
		{"git@github.com:myorg/myrepo.git", "myorg", "myrepo", false},
		{"https://github.com/myorg/myrepo", "myorg", "myrepo", false},
		{"ssh://git@github.com/myorg/myrepo.git", "myorg", "myrepo", false},
		{"not-a-url", "", "", true},
	}
	for _, tt := range tests {
		owner, repo, err := git.ParseOwnerRepo(tt.remote)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseOwnerRepo(%q) error = %v, wantErr %v", tt.remote, err, tt.wantErr)
			continue
		}
		if owner != tt.wantOwner || repo != tt.wantRepo {
			t.Errorf("ParseOwnerRepo(%q) = (%q, %q), want (%q, %q)", tt.remote, owner, repo, tt.wantOwner, tt.wantRepo)
		}
	}
}

func TestWorktreePath(t *testing.T) {
	path := git.WorktreePath("/repo/root", 142)
	want := "/repo/root/.worktrees/pr-142"
	if path != want {
		t.Errorf("WorktreePath = %q, want %q", path, want)
	}
}

func TestRemoveWorktree_NonExistent(t *testing.T) {
	err := git.RemoveWorktree("/tmp/nonexistent-repo-gh-review-test", 99999)
	if err == nil {
		t.Error("expected error for non-existent worktree, got nil")
	}
}
