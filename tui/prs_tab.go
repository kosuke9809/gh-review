package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kosuke9809/gh-review/model"
)

type prItem struct {
	pr             model.PR
	totalReviewers int
}

func (p prItem) Title() string       { return fmt.Sprintf("#%d  %s", p.pr.Number, p.pr.Title) }
func (p prItem) Description() string { return p.renderMeta() }
func (p prItem) FilterValue() string { return p.pr.Title }

func (p prItem) renderMeta() string {
	ci := ciIconStr(string(p.pr.CIStatus))
	approved := 0
	for _, r := range p.pr.Reviews {
		if r.State == "APPROVED" {
			approved++
		}
	}
	review := fmt.Sprintf("%d/%d", approved, p.totalReviewers)
	badge := badgeForState(string(p.pr.ReviewState))
	wt := ""
	if p.pr.HasWorktree {
		wt = lipgloss.NewStyle().Foreground(colorGreen).Render(" [wt]")
	}
	return fmt.Sprintf("CI:%s  Review:%s  %s%s", ci, review, badge, wt)
}

// FormatPRRow returns a text representation of a PR row (used in tests).
func FormatPRRow(pr model.PR, totalReviewers int, selected bool) string {
	item := prItem{pr: pr, totalReviewers: totalReviewers}
	base := item.Title() + "  " + item.renderMeta()
	if selected {
		return styleSelected.Render(base)
	}
	return base
}

type prsTabModel struct {
	list   list.Model
	prs    []model.PR
	width  int
	height int
}

func newPRsTab(width, height int) prsTabModel {
	delegate := list.NewDefaultDelegate()
	l := list.New(nil, delegate, width, height-4)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	return prsTabModel{list: l, width: width, height: height}
}

func (m prsTabModel) SetPRs(prs []model.PR) prsTabModel {
	m.prs = prs
	items := make([]list.Item, len(prs))
	for i, pr := range prs {
		items[i] = prItem{pr: pr, totalReviewers: len(pr.Reviews) + 1}
	}
	m.list.SetItems(items)
	return m
}

func (m prsTabModel) SelectedPR() *model.PR {
	if item, ok := m.list.SelectedItem().(prItem); ok {
		return &item.pr
	}
	return nil
}

func (m prsTabModel) Update(msg tea.Msg) (prsTabModel, tea.Cmd) {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m prsTabModel) View() string {
	if len(m.prs) == 0 {
		return lipgloss.NewStyle().
			Width(m.width).
			Height(m.height - 4).
			Align(lipgloss.Center, lipgloss.Center).
			Render("No PRs awaiting review")
	}
	return m.list.View()
}
