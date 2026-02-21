package model_test

import (
	"testing"

	"github.com/kosuke9809/gh-review/model"
)

func TestFilterPRs(t *testing.T) {
	prs := []model.PR{
		{Number: 1, Author: "alice", IsReviewRequested: true},
		{Number: 2, Author: "bob", IsReviewRequested: false},
		{Number: 3, Author: "me", IsReviewRequested: false},
		{Number: 4, Author: "me", IsReviewRequested: true},
	}

	tests := []struct {
		name        string
		filter      model.PRFilter
		currentUser string
		wantNums    []int
	}{
		{
			name:        "FilterReviewRequested: IsReviewRequestedがtrueのものだけ",
			filter:      model.FilterReviewRequested,
			currentUser: "me",
			wantNums:    []int{1, 4},
		},
		{
			name:        "FilterAuthored: 自分が作成したものだけ",
			filter:      model.FilterAuthored,
			currentUser: "me",
			wantNums:    []int{3, 4},
		},
		{
			name:        "FilterAll: 全件返す",
			filter:      model.FilterAll,
			currentUser: "me",
			wantNums:    []int{1, 2, 3, 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := model.FilterPRs(prs, tt.filter, tt.currentUser)
			if len(got) != len(tt.wantNums) {
				t.Fatalf("FilterPRs() len = %d, want %d", len(got), len(tt.wantNums))
			}
			for i, pr := range got {
				if pr.Number != tt.wantNums[i] {
					t.Errorf("FilterPRs()[%d].Number = %d, want %d", i, pr.Number, tt.wantNums[i])
				}
			}
		})
	}
}
