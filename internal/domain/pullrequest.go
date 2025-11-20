package domain

import "time"

type PullRequest struct {
	PullRequestID     string
	PullRequestName   string
	AuthorID          string
	Status            PRStatus // OPEN или MERGED
	AssignedReviewers []string // до 2 user_id
	CreatedAt         *time.Time
	MergedAt          *time.Time
}

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)
