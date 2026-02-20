package github

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	gogithub "github.com/google/go-github/v68/github"
	"github.com/kosuke9809/gh-review/internal/model"
	"golang.org/x/oauth2"
)

// Token obtains a GitHub token using `gh auth token`.
func Token() (string, error) {
	out, err := exec.Command("gh", "auth", "token").Output()
	if err != nil {
		return "", fmt.Errorf("gh auth token failed (is gh installed and authenticated?): %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// NewClient creates an authenticated go-github client.
func NewClient(token string) *gogithub.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	return gogithub.NewClient(tc)
}

// CurrentUser returns the login of the authenticated user.
func CurrentUser(ctx context.Context, client *gogithub.Client) (string, error) {
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return "", err
	}
	return user.GetLogin(), nil
}

// FetchPRs fetches open PRs where the current user is a requested reviewer,
// using GitHub Search API to reliably handle both individual and team review requests.
func FetchPRs(ctx context.Context, client *gogithub.Client, owner, repo, currentUser string) ([]*gogithub.PullRequest, error) {
	query := fmt.Sprintf("is:open is:pr review-requested:%s repo:%s/%s", currentUser, owner, repo)
	searchOpts := &gogithub.SearchOptions{
		ListOptions: gogithub.ListOptions{PerPage: 100},
	}

	var result []*gogithub.PullRequest
	for {
		searchResult, resp, err := client.Search.Issues(ctx, query, searchOpts)
		if err != nil {
			return nil, fmt.Errorf("search PRs: %w", err)
		}
		for _, issue := range searchResult.Issues {
			pr, _, err := client.PullRequests.Get(ctx, owner, repo, issue.GetNumber())
			if err != nil {
				continue
			}
			result = append(result, pr)
		}
		if resp.NextPage == 0 {
			break
		}
		searchOpts.Page = resp.NextPage
	}
	return result, nil
}

// FetchReviews fetches all reviews for a PR.
func FetchReviews(ctx context.Context, client *gogithub.Client, owner, repo string, prNumber int) ([]model.Review, error) {
	reviews, _, err := client.PullRequests.ListReviews(ctx, owner, repo, prNumber, nil)
	if err != nil {
		return nil, err
	}
	var result []model.Review
	for _, r := range reviews {
		result = append(result, model.Review{
			Author:    r.GetUser().GetLogin(),
			State:     r.GetState(),
			CreatedAt: r.GetSubmittedAt().Time,
		})
	}
	return result, nil
}

// FetchCheckRuns fetches CI check runs for a commit SHA.
func FetchCheckRuns(ctx context.Context, client *gogithub.Client, owner, repo, sha string) ([]model.CheckRun, model.CIStatus, error) {
	checks, _, err := client.Checks.ListCheckRunsForRef(ctx, owner, repo, sha, nil)
	if err != nil {
		return nil, model.CIStatusUnknown, err
	}
	var runs []model.CheckRun
	overall := model.CIStatusPass
	for _, c := range checks.CheckRuns {
		status := model.CIStatusUnknown
		switch c.GetConclusion() {
		case "success":
			status = model.CIStatusPass
		case "failure", "cancelled", "timed_out":
			status = model.CIStatusFail
			overall = model.CIStatusFail
		default:
			if c.GetStatus() == "in_progress" || c.GetStatus() == "queued" {
				status = model.CIStatusPending
				if overall == model.CIStatusPass {
					overall = model.CIStatusPending
				}
			}
		}
		runs = append(runs, model.CheckRun{Name: c.GetName(), Status: status})
	}
	if len(runs) == 0 {
		overall = model.CIStatusUnknown
	}
	return runs, overall, nil
}

// FetchComments fetches review comments for a PR.
func FetchComments(ctx context.Context, client *gogithub.Client, owner, repo string, prNumber int) ([]model.Comment, error) {
	comments, _, err := client.PullRequests.ListComments(ctx, owner, repo, prNumber, nil)
	if err != nil {
		return nil, err
	}
	var result []model.Comment
	for _, c := range comments {
		result = append(result, model.Comment{
			Author:   c.GetUser().GetLogin(),
			Body:     c.GetBody(),
			Path:     c.GetPath(),
			Line:     int(c.GetLine()),
			IsUnread: true,
		})
	}
	return result, nil
}

// FetchDiff fetches the diff files for a PR.
func FetchDiff(ctx context.Context, client *gogithub.Client, owner, repo string, prNumber int) ([]model.DiffFile, error) {
	files, _, err := client.PullRequests.ListFiles(ctx, owner, repo, prNumber, nil)
	if err != nil {
		return nil, err
	}
	var result []model.DiffFile
	for _, f := range files {
		result = append(result, model.DiffFile{
			Filename: f.GetFilename(),
			Patch:    f.GetPatch(),
		})
	}
	return result, nil
}

// CalcReviewState determines the review state for the given user based on reviews and PR updated time.
func CalcReviewState(currentUser string, reviews []model.Review, prUpdatedAt time.Time) model.ReviewState {
	if len(reviews) == 0 {
		return model.ReviewStateNew
	}
	var lastReview *model.Review
	for i := range reviews {
		r := &reviews[i]
		if r.Author != currentUser {
			continue
		}
		if r.State == "COMMENTED" {
			continue
		}
		if lastReview == nil || r.CreatedAt.After(lastReview.CreatedAt) {
			lastReview = r
		}
	}
	if lastReview == nil {
		return model.ReviewStateNew
	}
	if prUpdatedAt.After(lastReview.CreatedAt) {
		return model.ReviewStateUpd
	}
	if lastReview.State == "APPROVED" {
		return model.ReviewStateDone
	}
	return model.ReviewStateChg
}
