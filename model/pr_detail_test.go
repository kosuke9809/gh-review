package model

import "testing"

func TestDetailLoaded_DefaultFalse(t *testing.T) {
	pr := PR{Number: 1, Title: "test"}
	if pr.DetailLoaded {
		t.Error("DetailLoaded should default to false")
	}
}

func TestFilterPRs_PreservesDetailLoaded(t *testing.T) {
	prs := []PR{
		{Number: 1, Author: "alice", DetailLoaded: true},
		{Number: 2, Author: "bob", DetailLoaded: false},
	}
	filtered := FilterPRs(prs, FilterAll, "alice")
	if len(filtered) != 2 {
		t.Fatalf("expected 2 PRs, got %d", len(filtered))
	}
	if !filtered[0].DetailLoaded {
		t.Error("DetailLoaded should be preserved through FilterPRs")
	}
}
