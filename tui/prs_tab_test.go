package tui_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/kosuke9809/gh-review/model"
	"github.com/kosuke9809/gh-review/tui"
	"github.com/muesli/termenv"
)

func init() {
	lipgloss.DefaultRenderer().SetColorProfile(termenv.TrueColor)
}

func TestFormatPRRow(t *testing.T) {
	pr := model.PR{
		Number:      142,
		Title:       "Fix auth bug",
		CIStatus:    model.CIStatusPass,
		ReviewState: model.ReviewStateUpd,
		HasWorktree: true,
	}
	row := tui.FormatPRRow(pr, 2, false)
	if row == "" {
		t.Error("FormatPRRow returned empty string")
	}
}

func TestRenderDetail(t *testing.T) {
	pr := model.PR{
		Number:    142,
		Title:     "Fix auth bug",
		Author:    "alice",
		CIStatus:  model.CIStatusPass,
		CheckRuns: []model.CheckRun{{Name: "unit-tests", Status: model.CIStatusPass}},
		Reviews:   []model.Review{{Author: "bob", State: "APPROVED"}},
		Comments:  []model.Comment{{Author: "alice", Body: "LGTM", IsUnread: true}},
	}
	content := tui.RenderDetailContent(pr)
	if content == "" {
		t.Error("RenderDetailContent returned empty string")
	}
	if !strings.Contains(content, "alice") {
		t.Error("expected author name in detail content")
	}
}

func TestColorDiffLine(t *testing.T) {
	tests := []struct {
		line     string
		hasColor bool
	}{
		{"+added line", true},
		{"-removed line", true},
		{"@@ -1,5 +1,5 @@", true},
		{" context line", false},
	}
	for _, tt := range tests {
		result := tui.ColorDiffLine(tt.line)
		if tt.hasColor && result == tt.line {
			t.Errorf("ColorDiffLine(%q) should apply color but got unchanged", tt.line)
		}
	}
}
