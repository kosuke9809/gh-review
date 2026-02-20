package tui_test

import (
	"strings"
	"testing"

	"github.com/kosuke9809/gh-review/internal/model"
	"github.com/kosuke9809/gh-review/internal/tui"
)

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
