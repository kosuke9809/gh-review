package model

import "time"

type Tab int

const (
	TabPRs Tab = iota
	TabDetail
	TabDiff
)

type PRFilter int

const (
	FilterReviewRequested PRFilter = iota // review-requested:@me
	FilterAuthored                        // author:@me
	FilterAll                             // all open PRs
)

func (f PRFilter) Label() string {
	switch f {
	case FilterReviewRequested:
		return "Review Requested"
	case FilterAuthored:
		return "Authored"
	case FilterAll:
		return "All Open"
	}
	return ""
}

func (f PRFilter) Next() PRFilter {
	return (f + 1) % 3
}

type ReviewState string

const (
	ReviewStateNew  ReviewState = "NEW"  // never reviewed
	ReviewStateUpd  ReviewState = "UPD"  // commits after your review
	ReviewStateDone ReviewState = "DONE" // you approved
	ReviewStateChg  ReviewState = "CHG"  // you requested changes
)

type CIStatus string

const (
	CIStatusPass    CIStatus = "pass"
	CIStatusFail    CIStatus = "fail"
	CIStatusPending CIStatus = "pending"
	CIStatusUnknown CIStatus = "unknown"
)

type CheckRun struct {
	Name   string
	Status CIStatus
}

type Review struct {
	Author    string
	State     string // "APPROVED", "CHANGES_REQUESTED", "COMMENTED"
	CreatedAt time.Time
}

type Comment struct {
	Author   string
	Body     string
	Path     string
	Line     int
	IsUnread bool
	Replies  []Comment
}

type DiffFile struct {
	Filename string
	Patch    string
}

type PR struct {
	Number       int
	Title        string
	Author       string
	BaseRef      string
	HeadRef      string
	HeadSHA      string
	Body         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	HTMLURL      string
	CIStatus     CIStatus
	CheckRuns    []CheckRun
	Reviews      []Review
	Comments     []Comment
	DiffFiles    []DiffFile
	ReviewState  ReviewState
	HasWorktree  bool
	WorktreePath string
}
