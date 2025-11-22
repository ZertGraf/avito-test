package domain

import "time"

type PullRequest struct {
	PullRequestID     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            PRStatus   `json:"status"`             // OPEN или MERGED
	AssignedReviewers []string   `json:"assigned_reviewers"` // до 2 user_id
	CreatedAt         *time.Time `json:"created_at"`
	MergedAt          *time.Time `json:"merged_at"`
}

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

type PullRequestShort struct {
	PullRequestID   string   `json:"pull_request_id"`
	PullRequestName string   `json:"pull_request_name"`
	AuthorID        string   `json:"author_id"`
	Status          PRStatus `json:"status"`
}
