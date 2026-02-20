package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kosuke9809/gh-review/internal/model"
)

// ColorDiffLine applies color to a single diff line.
func ColorDiffLine(line string) string {
	switch {
	case strings.HasPrefix(line, "+"):
		return styleDiffAdd.Render(line)
	case strings.HasPrefix(line, "-"):
		return styleDiffDel.Render(line)
	case strings.HasPrefix(line, "@@"):
		return styleDiffHdr.Render(line)
	}
	return line
}

func colorDiff(patch string) string {
	lines := strings.Split(patch, "\n")
	for i, line := range lines {
		lines[i] = ColorDiffLine(line)
	}
	return strings.Join(lines, "\n")
}

type fileItem struct{ name string }

func (f fileItem) Title() string       { return f.name }
func (f fileItem) Description() string { return "" }
func (f fileItem) FilterValue() string { return f.name }

type diffTabModel struct {
	fileList  list.Model
	diffView  viewport.Model
	files     []model.DiffFile
	focusLeft bool
	width     int
	height    int
}

func newDiffTab(width, height int) diffTabModel {
	leftW := width / 4
	rightW := width - leftW - 3

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	fl := list.New(nil, delegate, leftW, height-4)
	fl.SetShowStatusBar(false)
	fl.SetFilteringEnabled(false)
	fl.SetShowHelp(false)
	fl.Title = "Files"

	dv := viewport.New(rightW, height-4)

	return diffTabModel{
		fileList:  fl,
		diffView:  dv,
		focusLeft: true,
		width:     width,
		height:    height,
	}
}

func (m diffTabModel) SetFiles(files []model.DiffFile) diffTabModel {
	m.files = files
	items := make([]list.Item, len(files))
	for i, f := range files {
		items[i] = fileItem{name: f.Filename}
	}
	m.fileList.SetItems(items)
	if len(files) > 0 {
		m.diffView.SetContent(colorDiff(files[0].Patch))
	}
	return m
}

func (m diffTabModel) updateDiffView() diffTabModel {
	idx := m.fileList.Index()
	if idx >= 0 && idx < len(m.files) {
		m.diffView.SetContent(colorDiff(m.files[idx].Patch))
		m.diffView.GotoTop()
	}
	return m
}

func (m diffTabModel) Update(msg tea.Msg) (diffTabModel, tea.Cmd) {
	var cmd tea.Cmd
	if m.focusLeft {
		prev := m.fileList.Index()
		m.fileList, cmd = m.fileList.Update(msg)
		if m.fileList.Index() != prev {
			m = m.updateDiffView()
		}
		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "tab" {
			m.focusLeft = false
		}
	} else {
		m.diffView, cmd = m.diffView.Update(msg)
		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "tab" {
			m.focusLeft = true
		}
	}
	return m, cmd
}

func (m diffTabModel) View() string {
	leftW := m.width / 4

	leftBorder := styleBorder
	rightBorder := styleBorder
	if m.focusLeft {
		leftBorder = leftBorder.BorderForeground(colorBlue)
	} else {
		rightBorder = rightBorder.BorderForeground(colorBlue)
	}

	left := leftBorder.Width(leftW).Render(m.fileList.View())
	right := rightBorder.Width(m.width - leftW - 4).Render(m.diffView.View())

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}
