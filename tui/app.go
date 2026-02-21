package tui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	gogithub "github.com/google/go-github/v68/github"
	"github.com/kosuke9809/gh-review/git"
	"github.com/kosuke9809/gh-review/github"
	"github.com/kosuke9809/gh-review/model"
	"golang.org/x/sync/errgroup"
)

// Context returns a background context for use in main.go.
func Context() context.Context {
	return context.Background()
}
type screen int

const (
	screenList   screen = iota
	screenDetail
)

type detailSubTab int

const (
	subTabDetail detailSubTab = iota
	subTabDiff
)


type fetchedMsg struct {
	prs []model.PR
	err error
}

type detailFetchedMsg struct {
	prNumber int
	reviews  []model.Review
	checks   []model.CheckRun
	ciStatus model.CIStatus
	comments []model.Comment
	files    []model.DiffFile
	err      error
}

type tickMsg time.Time

// AppModel is the root bubbletea model.
type AppModel struct {
	screen        screen
	detailSubTab  detailSubTab
	selectedPR    *model.PR
	filter        model.PRFilter
	prsTab        prsTabModel
	detailTab     detailTabModel
	diffTab       diffTabModel
	allPRs        []model.PR
	prs           []model.PR
	loading       bool
	err           error
	lastSync      time.Time
	repoName      string
	repoOwner     string
	repoRepo      string
	repoRoot      string
	currentUser   string
	ghClient      *gogithub.Client
	width         int
	height        int
	spinner       spinner.Model
	loadingDetail bool
}

// New creates a new AppModel.
func New(owner, repo, repoRoot, currentUser string, client *gogithub.Client, width, height int) AppModel {
	inner := width - 2
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(colorGreen)
	return AppModel{
		screen:      screenList,
		prsTab:      newPRsTab(inner, height),
		detailTab:   newDetailTab(inner, height),
		diffTab:     newDiffTab(inner, height),
		loading:     true,
		repoName:    owner + "/" + repo,
		repoOwner:   owner,
		repoRepo:    repo,
		repoRoot:    repoRoot,
		currentUser: currentUser,
		ghClient:    client,
		width:       width,
		height:      height,
		spinner:     sp,
	}
}

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(m.fetchCmd(), tickCmd(), m.spinner.Tick)
}

func tickCmd() tea.Cmd {
	return tea.Tick(60*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m AppModel) fetchCmd() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		ghPRs, err := github.FetchPRs(ctx, m.ghClient, m.repoOwner, m.repoRepo, m.currentUser)
		if err != nil {
			return fetchedMsg{err: err}
		}
		prs := make([]model.PR, 0, len(ghPRs))
		for _, ghPR := range ghPRs {
			wtPath := git.WorktreePath(m.repoRoot, int(ghPR.GetNumber()))
			hasWt := git.WorktreeExists(m.repoRoot, int(ghPR.GetNumber()))
			prs = append(prs, model.PR{
				Number:            int(ghPR.GetNumber()),
				Title:             ghPR.GetTitle(),
				Author:            ghPR.GetUser().GetLogin(),
				BaseRef:           ghPR.GetBase().GetRef(),
				HeadRef:           ghPR.GetHead().GetRef(),
				HeadSHA:           ghPR.GetHead().GetSHA(),
				Body:              ghPR.GetBody(),
				CreatedAt:         ghPR.GetCreatedAt().Time,
				UpdatedAt:         ghPR.GetUpdatedAt().Time,
				HTMLURL:           ghPR.GetHTMLURL(),
				CIStatus:          model.CIStatusUnknown,
				IsReviewRequested: github.IsReviewRequested(ghPR, m.currentUser),
				HasWorktree:       hasWt,
				WorktreePath:      wtPath,
				DetailLoaded:      false,
			})
		}
		return fetchedMsg{prs: prs}
	}
}

func (m AppModel) detailFetchCmd(pr model.PR) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		var (
			reviews  []model.Review
			checks   []model.CheckRun
			ciStatus model.CIStatus
			comments []model.Comment
			files    []model.DiffFile
		)
		eg, ctx := errgroup.WithContext(ctx)
		eg.Go(func() error {
			var err error
			reviews, err = github.FetchReviews(ctx, m.ghClient, m.repoOwner, m.repoRepo, pr.Number)
			return err
		})
		eg.Go(func() error {
			var err error
			checks, ciStatus, err = github.FetchCheckRuns(ctx, m.ghClient, m.repoOwner, m.repoRepo, pr.HeadSHA)
			return err
		})
		eg.Go(func() error {
			var err error
			comments, err = github.FetchComments(ctx, m.ghClient, m.repoOwner, m.repoRepo, pr.Number)
			return err
		})
		eg.Go(func() error {
			var err error
			files, err = github.FetchDiff(ctx, m.ghClient, m.repoOwner, m.repoRepo, pr.Number)
			return err
		})
		if err := eg.Wait(); err != nil {
			return detailFetchedMsg{prNumber: pr.Number, err: err}
		}
		return detailFetchedMsg{
			prNumber: pr.Number,
			reviews:  reviews,
			checks:   checks,
			ciStatus: ciStatus,
			comments: comments,
			files:    files,
		}
	}
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		inner := msg.Width - 2
		m.prsTab = newPRsTab(inner, msg.Height).SetPRs(m.prs)
		m.detailTab = newDetailTab(inner, msg.Height)
		m.diffTab = newDiffTab(inner, msg.Height)
		if m.selectedPR != nil {
			m.detailTab = m.detailTab.SetPR(m.selectedPR)
			m.diffTab = m.diffTab.SetFiles(m.selectedPR.DiffFiles)
		}

	case fetchedMsg:
		m.loading = false
		m.lastSync = time.Now()
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.err = nil
			m.allPRs = msg.prs
			m = m.applyFilter()
			// Re-fetch details for currently viewed PR
			if m.selectedPR != nil {
				for _, pr := range m.allPRs {
					if pr.Number == m.selectedPR.Number {
						pr := pr
						m.selectedPR = &pr
						if !pr.DetailLoaded {
							m.loadingDetail = true
							return m, m.detailFetchCmd(pr)
						}
					}
				}
			}
		}

	case detailFetchedMsg:
		m.loadingDetail = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			for i, pr := range m.allPRs {
				if pr.Number == msg.prNumber {
					m.allPRs[i].Reviews = msg.reviews
					m.allPRs[i].CheckRuns = msg.checks
					m.allPRs[i].CIStatus = msg.ciStatus
					m.allPRs[i].Comments = msg.comments
					m.allPRs[i].DiffFiles = msg.files
					m.allPRs[i].DetailLoaded = true
					m.allPRs[i].ReviewState = github.CalcReviewState(m.currentUser, msg.reviews, m.allPRs[i].UpdatedAt)
					// Update selectedPR if this is the PR being viewed
					if m.selectedPR != nil && m.selectedPR.Number == msg.prNumber {
						updated := m.allPRs[i]
						m.selectedPR = &updated
						m.detailTab = m.detailTab.SetPR(m.selectedPR)
						m.diffTab = m.diffTab.SetFiles(m.selectedPR.DiffFiles)
					}
					break
				}
			}
			m = m.applyFilter()
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tickMsg:
		m.loading = true
		return m, tea.Batch(m.fetchCmd(), tickCmd())

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc", "b":
			if m.screen == screenDetail {
				m.screen = screenList
				return m, nil
			}
		case "tab":
			if m.screen == screenDetail {
				if m.detailSubTab == subTabDetail {
					m.detailSubTab = subTabDiff
				} else {
					m.detailSubTab = subTabDetail
				}
				return m, nil
			}
		case "f":
			if m.screen == screenList {
				m.filter = m.filter.Next()
				m = m.applyFilter()
				return m, nil
			}
		case "r":
			m.loading = true
			return m, m.fetchCmd()
		case "enter":
			if m.screen == screenList {
				if pr := m.prsTab.SelectedPR(); pr != nil {
					pr := *pr
					m.selectedPR = &pr
					m.screen = screenDetail
					m.detailSubTab = subTabDetail
					m.detailTab = m.detailTab.SetPR(m.selectedPR)
					m.diffTab = m.diffTab.SetFiles(m.selectedPR.DiffFiles)
					if !pr.DetailLoaded {
						m.loadingDetail = true
						return m, m.detailFetchCmd(pr)
					}
					return m, nil
				}
			}
		case "w":
			if m.screen == screenList {
				return m, m.worktreeCmd()
			}
		case "o":
			if m.screen == screenList {
				if pr := m.prsTab.SelectedPR(); pr != nil && pr.HasWorktree {
					return m, openEditorCmd(pr.WorktreePath)
				}
			}
		case "D":
			if m.screen == screenList {
				return m, m.removeWorktreeCmd()
			}
		}
	}

	var cmd tea.Cmd
	if m.screen == screenList {
		m.prsTab, cmd = m.prsTab.Update(msg)
	} else {
		switch m.detailSubTab {
		case subTabDetail:
			m.detailTab, cmd = m.detailTab.Update(msg)
		case subTabDiff:
			m.diffTab, cmd = m.diffTab.Update(msg)
		}
	}
	return m, cmd
}

// applyFilter filters allPRs client-side and updates the PRs tab.
func (m AppModel) applyFilter() AppModel {
	m.prs = model.FilterPRs(m.allPRs, m.filter, m.currentUser)
	m.prsTab = m.prsTab.SetPRs(m.prs)
	return m
}

func (m AppModel) worktreeCmd() tea.Cmd {
	pr := m.prsTab.SelectedPR()
	if pr == nil {
		return nil
	}
	prNum := pr.Number
	repoRoot := m.repoRoot
	return func() tea.Msg {
		if !git.WorktreeExists(repoRoot, prNum) {
			if err := git.CreateWorktree(repoRoot, prNum); err != nil {
				return fetchedMsg{err: fmt.Errorf("worktree: %w", err)}
			}
		}
		return tickMsg(time.Now())
	}
}

func (m AppModel) removeWorktreeCmd() tea.Cmd {
	pr := m.prsTab.SelectedPR()
	if pr == nil || !pr.HasWorktree {
		return nil
	}
	prNum := pr.Number
	repoRoot := m.repoRoot
	return func() tea.Msg {
		if err := git.RemoveWorktree(repoRoot, prNum); err != nil {
			return fetchedMsg{err: fmt.Errorf("remove worktree: %w", err)}
		}
		return tickMsg(time.Now())
	}
}

func openEditorCmd(path string) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	return tea.ExecProcess(exec.Command(editor, path), func(err error) tea.Msg {
		return nil
	})
}

func (m AppModel) View() string {
	top := m.buildTopBorder()
	body := m.renderBody()
	bottom := m.buildBottomBorder()
	return lipgloss.JoinVertical(lipgloss.Left, top, body, bottom)
}

func (m AppModel) buildTopBorder() string {
	var inner string
	if m.screen == screenList {
		title := fmt.Sprintf("[gh-review — %s]", m.repoName)
		filter := lipgloss.NewStyle().Foreground(colorYellow).Render("[f] " + m.filter.Label())
		inner = "─" + title + "─" + filter
	} else {
		prTitle := ""
		if m.selectedPR != nil {
			prTitle = fmt.Sprintf(" — #%d %s", m.selectedPR.Number, m.selectedPR.Title)
		}
		title := fmt.Sprintf("[gh-review — %s%s]", m.repoName, prTitle)
		subTabs := m.renderSubTabsStr()
		inner = "─" + title + "─" + subTabs
	}
	innerW := lipgloss.Width(inner)
	pad := m.width - 2 - innerW
	if pad < 0 {
		pad = 0
	}
	line := "┌" + inner + strings.Repeat("─", pad) + "┐"
	return lipgloss.NewStyle().Foreground(colorGreen).Render(line)
}

func (m AppModel) buildBottomBorder() string {
	help := m.helpStr()
	sync := m.syncStr()
	helpW := lipgloss.Width(help)
	syncW := lipgloss.Width(sync)
	pad := m.width - 2 - helpW - syncW - 1
	if pad < 0 {
		pad = 0
	}
	line := "└─" + help + strings.Repeat("─", pad) + sync + "─┘"
	return lipgloss.NewStyle().Foreground(colorGreen).Render(line)
}

func (m AppModel) renderSubTabsStr() string {
	var detail, diff string
	if m.detailSubTab == subTabDetail {
		detail = styleTabActive.Render("Detail")
		diff = styleTabInactive.Render("Diff")
	} else {
		detail = styleTabInactive.Render("Detail")
		diff = styleTabActive.Render("Diff")
	}
	return detail + diff
}

func (m AppModel) helpStr() string {
	if m.screen == screenList {
		return "[Enter]detail [w]worktree [o]open [D]delete [f]filter [r]efresh [q]quit"
	}
	switch m.detailSubTab {
	case subTabDiff:
		return "[tab]switch [enter]focus [j/k]scroll [Esc/b]back [q]quit"
	default:
		return "[tab]switch [j/k]scroll [Esc/b]back [q]quit"
	}
}

func (m AppModel) syncStr() string {
	if m.err != nil {
		return lipgloss.NewStyle().Foreground(colorRed).Render("Error: " + m.err.Error())
	}
	if m.loading {
		return "Syncing..."
	}
	if !m.lastSync.IsZero() {
		age := time.Since(m.lastSync).Round(time.Second)
		return fmt.Sprintf("%s ago", age)
	}
	return ""
}

func (m AppModel) renderBody() string {
	if m.loading && len(m.allPRs) == 0 {
		return lipgloss.NewStyle().
			Width(m.width - 2).
			Height(m.height - 4).
			Align(lipgloss.Center, lipgloss.Center).
			Render(m.spinner.View() + " Loading PRs...")
	}
	if m.screen == screenList {
		return m.prsTab.View()
	}
	// screenDetail
	if m.loadingDetail {
		return lipgloss.NewStyle().
			Width(m.width - 2).
			Height(m.height - 4).
			Align(lipgloss.Center, lipgloss.Center).
			Render(m.spinner.View() + " Loading details...")
	}
	switch m.detailSubTab {
	case subTabDetail:
		return m.detailTab.View()
	case subTabDiff:
		return m.diffTab.View()
	}
	return ""
}
