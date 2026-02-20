package tui_test

import (
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
