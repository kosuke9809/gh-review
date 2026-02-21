package github_test

import (
	"testing"
	"time"

	"github.com/kosuke9809/gh-review/github"
	"github.com/kosuke9809/gh-review/model"
)

func TestCalcReviewState(t *testing.T) {
	now := time.Now()
	lastReview := now.Add(-2 * time.Hour)
	afterReview := now.Add(-1 * time.Hour)

	tests := []struct {
		name        string
		reviews     []model.Review
		prUpdatedAt time.Time
		want        model.ReviewState
	}{
		{
			name:        "no reviews → NEW",
			reviews:     nil,
			prUpdatedAt: now,
			want:        model.ReviewStateNew,
		},
		{
			name: "approved → DONE",
			reviews: []model.Review{
				{Author: "me", State: "APPROVED", CreatedAt: lastReview},
			},
			prUpdatedAt: lastReview.Add(-time.Minute),
			want:        model.ReviewStateDone,
		},
		{
			name: "approved but updated after → UPD",
			reviews: []model.Review{
				{Author: "me", State: "APPROVED", CreatedAt: lastReview},
			},
			prUpdatedAt: afterReview,
			want:        model.ReviewStateUpd,
		},
		{
			name: "changes requested → CHG",
			reviews: []model.Review{
				{Author: "me", State: "CHANGES_REQUESTED", CreatedAt: lastReview},
			},
			prUpdatedAt: lastReview.Add(-time.Minute),
			want:        model.ReviewStateChg,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := github.CalcReviewState("me", tt.reviews, tt.prUpdatedAt)
			if got != tt.want {
				t.Errorf("CalcReviewState() = %v, want %v", got, tt.want)
			}
		})
	}
}
