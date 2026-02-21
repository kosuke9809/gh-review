package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorGreen  = lipgloss.Color("2")
	colorRed    = lipgloss.Color("1")
	colorYellow = lipgloss.Color("3")
	colorBlue   = lipgloss.Color("4")
	colorGray   = lipgloss.Color("8")
	colorWhite  = lipgloss.Color("15")

	styleTabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite).
			Background(colorBlue).
			Padding(0, 1)

	styleTabInactive = lipgloss.NewStyle().
				Foreground(colorGray).
				Padding(0, 1)

	styleStatusBar = lipgloss.NewStyle().
			Foreground(colorGray)

	styleBadgeNew  = lipgloss.NewStyle().Foreground(colorBlue).Bold(true)
	styleBadgeUpd  = lipgloss.NewStyle().Foreground(colorYellow).Bold(true)
	styleBadgeDone = lipgloss.NewStyle().Foreground(colorGreen).Bold(true)
	styleBadgeChg  = lipgloss.NewStyle().Foreground(colorRed).Bold(true)

	styleCIPass    = lipgloss.NewStyle().Foreground(colorGreen)
	styleCIFail    = lipgloss.NewStyle().Foreground(colorRed)
	styleCIPending = lipgloss.NewStyle().Foreground(colorYellow)

	styleUnread = lipgloss.NewStyle().Foreground(colorYellow).Bold(true)

	styleDiffAdd = lipgloss.NewStyle().Foreground(colorGreen)
	styleDiffDel = lipgloss.NewStyle().Foreground(colorRed)
	styleDiffHdr = lipgloss.NewStyle().Foreground(colorBlue)

	styleBorder = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorGray)

	styleSelected = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Bold(true)
)

func badgeForState(state string) string {
	switch state {
	case "NEW":
		return styleBadgeNew.Render("[NEW]")
	case "UPD":
		return styleBadgeUpd.Render("[UPD]")
	case "DONE":
		return styleBadgeDone.Render("[DONE]")
	case "CHG":
		return styleBadgeChg.Render("[CHG]")
	}
	return ""
}

func ciIconStr(s string) string {
	switch s {
	case "pass":
		return styleCIPass.Render("✓")
	case "fail":
		return styleCIFail.Render("✗")
	case "pending":
		return styleCIPending.Render("●")
	}
	return "?"
}
