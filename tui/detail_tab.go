package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kosuke9809/gh-review/model"
)

type detailTabModel struct {
	viewport viewport.Model
	pr       *model.PR
	width    int
	height   int
}

func newDetailTab(width, height int) detailTabModel {
	vp := viewport.New(width, height-4)
	return detailTabModel{viewport: vp, width: width, height: height}
}

func (m detailTabModel) SetPR(pr *model.PR) detailTabModel {
	m.pr = pr
	if pr != nil {
		m.viewport.SetContent(RenderDetailContent(*pr))
	}
	return m
}

// RenderDetailContent builds the text content for the Detail tab.
func RenderDetailContent(pr model.PR) string {
	var b strings.Builder
	sep := lipgloss.NewStyle().Foreground(colorGray).Render(strings.Repeat("─", 60))

	b.WriteString(lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("#%d: %s", pr.Number, pr.Title)))
	b.WriteString("\n\n")

	age := time.Since(pr.UpdatedAt).Round(time.Hour)
	b.WriteString(fmt.Sprintf("Author: %s  |  Created: %s  |  Updated: %s ago\n",
		pr.Author,
		pr.CreatedAt.Format("2006-01-02"),
		formatDuration(age),
	))
	b.WriteString(fmt.Sprintf("Branch: %s ← %s\n", pr.BaseRef, pr.HeadRef))

	if pr.HasWorktree {
		b.WriteString(fmt.Sprintf("Worktree: %s  [o:open] [D:delete]\n", pr.WorktreePath))
	}

	if pr.Body != "" {
		b.WriteString("\n" + sep + "\n")
		b.WriteString(lipgloss.NewStyle().Bold(true).Render("Description"))
		b.WriteString("\n\n")
		b.WriteString(pr.Body)
		b.WriteString("\n")
	}

	b.WriteString("\n" + sep + "\n")
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("CI Checks"))
	if len(pr.CheckRuns) == 0 {
		b.WriteString(fmt.Sprintf(" — %s\n", ciIconStr(string(pr.CIStatus))))
	} else {
		passed := 0
		for _, c := range pr.CheckRuns {
			if c.Status == model.CIStatusPass {
				passed++
			}
		}
		b.WriteString(fmt.Sprintf(" (%d/%d passed)\n", passed, len(pr.CheckRuns)))
		for _, c := range pr.CheckRuns {
			b.WriteString(fmt.Sprintf("  %s %s\n", ciIconStr(string(c.Status)), c.Name))
		}
	}

	b.WriteString("\n" + sep + "\n")
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Reviews"))
	b.WriteString("\n")
	if len(pr.Reviews) == 0 {
		b.WriteString("  No reviews yet\n")
	} else {
		for _, r := range pr.Reviews {
			icon := "○"
			switch r.State {
			case "APPROVED":
				icon = styleCIPass.Render("✓")
			case "CHANGES_REQUESTED":
				icon = styleCIFail.Render("✗")
			}
			b.WriteString(fmt.Sprintf("  %s %s: %s (%s)\n",
				icon, r.Author, strings.ToLower(r.State), r.CreatedAt.Format("2006-01-02")))
		}
	}

	b.WriteString("\n" + sep + "\n")
	unread := 0
	for _, c := range pr.Comments {
		if c.IsUnread {
			unread++
		}
	}
	b.WriteString(lipgloss.NewStyle().Bold(true).Render(
		fmt.Sprintf("Comments (%d total, %d unread)", len(pr.Comments), unread),
	))
	b.WriteString("\n")
	for _, c := range pr.Comments {
		prefix := "  "
		if c.IsUnread {
			prefix = styleUnread.Render("  ●")
		}
		loc := ""
		if c.Path != "" {
			loc = fmt.Sprintf(" [%s:%d]", c.Path, c.Line)
		}
		b.WriteString(fmt.Sprintf("%s %s%s: %s\n", prefix, c.Author, loc, c.Body))
	}

	return b.String()
}

func formatDuration(d time.Duration) string {
	if h := int(d.Hours()); h >= 24 {
		return fmt.Sprintf("%dd", h/24)
	} else if h > 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dm", int(d.Minutes()))
}

func (m detailTabModel) Update(msg tea.Msg) (detailTabModel, tea.Cmd) {
	var cmd tea.Cmd
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "j":
			m.viewport.ScrollDown(1)
			return m, nil
		case "k":
			m.viewport.ScrollUp(1)
			return m, nil
		}
	}
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m detailTabModel) View() string {
	if m.pr == nil {
		return lipgloss.NewStyle().
			Width(m.width).
			Height(m.height - 4).
			Align(lipgloss.Center, lipgloss.Center).
			Render("Select a PR to view details")
	}
	return m.viewport.View()
}
