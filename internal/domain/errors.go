package domain

import "errors"

var (
	ErrTeamExists   = errors.New("team already exists")
	ErrTeamNotFound = errors.New("team not found")
	ErrUserNotFound = errors.New("user not found")
	ErrPRExists     = errors.New("pull request already exists")
	ErrPRNotFound   = errors.New("pull request not found")
	ErrPRMerged     = errors.New("cannot modify merged pull request")
	ErrNotAssigned  = errors.New("user not assigned as reviewer")
	ErrNoCandidate  = errors.New("no available reviewers in team")
)
