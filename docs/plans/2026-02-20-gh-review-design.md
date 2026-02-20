# gh-review Design

Date: 2026-02-20

## Overview

`gh-review` is a `gh` CLI extension that centralizes GitHub PR review management for the current repository. It provides a multi-view TUI for visualizing PR status, CI results, and review progress, with git worktree-based fast switching between PRs.

## Goals

- Show all PRs with review requested to the current user in the current repository
- Visualize: PR title, CI status, review progress, unread comments, and review state (new/updated/done)
- Support git worktree-based workflow for switching between PRs
- Fast keyboard-driven navigation

## Non-Goals

- Multi-repository support (only the repository from which the tool is launched)
- Posting review comments from within the TUI (read-only)

## Tech Stack

| Role | Library |
|------|---------|
| TUI framework | [bubbletea](https://github.com/charmbracelet/bubbletea) |
| Styling | [lipgloss](https://github.com/charmbracelet/lipgloss) |
| UI components | [bubbles](https://github.com/charmbracelet/bubbles) (list, viewport) |
| GitHub API | [go-github](https://github.com/google/go-github) |
| Auth | `gh auth token` (reuse existing gh authentication) |
| Language | Go |

## Tab Structure

```
[1: PRs]  [2: Detail]  [3: Diff]
```

## Views

### Tab 1: PRs

Displays all open PRs with review requested to the current user.

```
┌─ gh-review — myorg/backend ──────────────────────────────────────────┐
│ [1:PRs] [2:Detail] [3:Diff]                       Last sync: 30s ago │
├──────────────────────────────────────────────────────────────────────┤
│  PR#   TITLE                              CI    REVIEW  STATE        │
│ ▶#142  Fix auth bug                       ✓     1/2     [UPD]  [wt] │
│  #89   Add dark mode                      ✗     0/2     [NEW]        │
│  #203  Refactor DB layer                  ✓     2/2     [DONE] [wt] │
│  #15   Update terraform                   ●     1/1     [NEW]        │
├──────────────────────────────────────────────────────────────────────┤
│ [Enter]worktree+detail [d]iff [o]open in editor [r]efresh [q]quit    │
└──────────────────────────────────────────────────────────────────────┘
```

**STATE badge definitions:**

| Badge | Meaning |
|-------|---------|
| `[NEW]` | Not yet reviewed |
| `[UPD]` | New commits pushed after your last review |
| `[DONE]` | You have approved |
| `[CHG]` | You requested changes |

`[wt]` indicates a git worktree has been created for this PR.

### Tab 2: Detail + Comments

Shows detailed information for the selected PR including CI checks, review status, and comments.

```
┌─ #142: Fix auth bug ─────────────────────────────────────────────────┐
│ Author: alice  |  Created: 2026-02-15  |  Updated: 2h ago            │
│ Worktree: .worktrees/pr-142  [o: open in editor]                     │
│                                                                       │
│ CI Checks (5/5 passed)                                               │
│  ✓ unit-tests  ✓ lint  ✓ build  ✓ e2e  ✓ security                   │
│                                                                       │
│ Reviews                                                              │
│  ✓ bob: approved (2026-02-18)                                        │
│  ● you: reviewed (2026-02-17) → 2 commits added since               │
│                                                                       │
│ Comments (3 total, 2 unread)                                         │
│  ● alice [auth/jwt.go:42]: Should we also check iss claim?           │
│    you: Good point, added to follow-up issue                         │
│  ● bob [README.md:10]: Typo in description                           │
└──────────────────────────────────────────────────────────────────────┘
```

Unread comments are highlighted with `●`.

### Tab 3: Diff

Split pane with file tree on the left and diff on the right.

```
┌─ Files ──────────┬─── Diff ───────────────────────────────────────────┐
│ ▶ auth/jwt.go    │ @@ -42,7 +42,12 @@                                 │
│   auth/jwt_test  │  func Validate(token string) error {               │
│   config/config  │ -    if token.ExpiresAt.IsZero() {                 │
│                  │ +    if token.ExpiresAt.Before(time.Now()) {       │
│                  │ +        return ErrTokenExpired                    │
│                  │ +    }                                             │
└──────────────────┴────────────────────────────────────────────────────┘
```

## Worktree Workflow

When Enter is pressed on a PR in the PRs tab:

1. If worktree does not exist: `git worktree add .worktrees/pr-{number} refs/pull/{number}/head`
2. If worktree already exists: switch to it immediately
3. Navigate to Detail tab
4. Worktree path is shown in Detail view; `[o]` opens it in `$EDITOR` or prints the path to stdout

Worktrees are stored at `.worktrees/pr-{number}/` relative to the repository root.

## Data Flow

```
Launch
 → Identify GitHub repository from git remote
 → Obtain auth token via `gh auth token`
 → GitHub API: fetch open PRs with review_requested=@me
 → Fetch CI status and review status for each PR (concurrent)
 → Render TUI

Background refresh: every 60 seconds
```

### GitHub API Endpoints (REST via go-github)

- `GET /repos/{owner}/{repo}/pulls?state=open` filtered by review requested
- `GET /repos/{owner}/{repo}/pulls/{number}/reviews`
- `GET /repos/{owner}/{repo}/commits/{sha}/check-runs`
- `GET /repos/{owner}/{repo}/pulls/{number}/comments`

## State Model

```go
type Model struct {
    activeTab     Tab
    prs           []PR
    selectedPR    int
    prsModel      list.Model
    detailModel   viewport.Model
    diffModel     diffPane
    loading       bool
    err           error
    lastSync      time.Time
}
```

## Error Handling

| Case | Behavior |
|------|----------|
| `gh` not installed | Print error message and exit on startup |
| Not in a git repository | Print error message and exit on startup |
| GitHub API failure | Show error in status bar, TUI continues |
| Worktree creation failure | Show inline error message |
| No PRs awaiting review | Show "No PRs awaiting review" in PRs tab |
