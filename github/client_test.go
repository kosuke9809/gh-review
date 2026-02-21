package github_test

import (
	"testing"
	"time"

	"github.com/kosuke9809/gh-review/github"
	"github.com/kosuke9809/gh-review/model"
)

func TestNewClientForHost(t *testing.T) {
	t.Run("github.com default", func(t *testing.T) {
		client, err := github.NewClientForHost("dummy-token", "github.com")
		if err != nil {
			t.Fatalf("NewClientForHost(github.com) error = %v", err)
		}
		if got := client.BaseURL.String(); got != "https://api.github.com/" {
			t.Fatalf("BaseURL = %q, want %q", got, "https://api.github.com/")
		}
	})

	t.Run("enterprise host", func(t *testing.T) {
		client, err := github.NewClientForHost("dummy-token", "github.example.com")
		if err != nil {
			t.Fatalf("NewClientForHost(github.example.com) error = %v", err)
		}
		if got := client.BaseURL.String(); got != "https://github.example.com/api/v3/" {
			t.Fatalf("BaseURL = %q, want %q", got, "https://github.example.com/api/v3/")
		}
		if got := client.UploadURL.String(); got != "https://github.example.com/api/uploads/" {
			t.Fatalf("UploadURL = %q, want %q", got, "https://github.example.com/api/uploads/")
		}
	})
}

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
