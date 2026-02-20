package main

import (
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kosuke9809/gh-review/internal/github"
	"github.com/kosuke9809/gh-review/internal/git"
	"github.com/kosuke9809/gh-review/internal/tui"
	"golang.org/x/term"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Check prerequisites
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("gh CLI not found — install from https://cli.github.com")
	}

	repoRoot, err := git.RepoRoot()
	if err != nil {
		return fmt.Errorf("not in a git repository")
	}

	remoteURL, err := git.RemoteURL("origin")
	if err != nil {
		return fmt.Errorf("no 'origin' remote found: %w", err)
	}

	owner, repo, err := git.ParseOwnerRepo(remoteURL)
	if err != nil {
		return fmt.Errorf("remote is not a GitHub URL: %w", err)
	}

	token, err := github.Token()
	if err != nil {
		return err
	}

	client := github.NewClient(token)

	ctx := tui.Context()
	currentUser, err := github.CurrentUser(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to get current GitHub user: %w", err)
	}

	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width, height = 120, 40
	}

	m := tui.New(owner, repo, repoRoot, currentUser, client, width, height)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}
	return nil
}
