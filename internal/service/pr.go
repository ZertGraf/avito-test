package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/ZertGraf/avito-test/internal/domain"
	"github.com/ZertGraf/avito-test/internal/pkg/logger"
	"github.com/ZertGraf/avito-test/internal/repository"
	"math/rand"
	"sync"
	"time"
)

type PRService struct {
	prRepo   repository.PRRepository
	userRepo repository.UserRepository
	logger   *logger.Logger
	random   *rand.Rand
	mu       *sync.Mutex
}

func NewPRService(
	prRepo repository.PRRepository,
	userRepo repository.UserRepository,
	logger *logger.Logger,
) *PRService {
	mu := new(sync.Mutex)
	return &PRService{
		prRepo:   prRepo,
		userRepo: userRepo,
		logger:   logger.Component("service/pr"),
		random:   rand.New(rand.NewSource(time.Now().UnixNano())),
		mu:       mu,
	}
}

// GetReviewsByUser retrieves all pull requests assigned to a specific reviewer.
func (s *PRService) GetReviewsByUser(ctx context.Context, userID string) ([]*domain.PullRequestShort, error) {
	prs, err := s.prRepo.GetByReviewer(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get reviews by user: %w", err)
	}

	s.logger.Info("retrieved user reviews",
		"user_id", userID,
		"count", len(prs),
	)

	return prs, nil
}

// CreatePR creates a new pull request and automatically assigns reviewers from author's team.
func (s *PRService) CreatePR(ctx context.Context, prID, prName, authorID string) (*domain.PullRequest, error) {
	exists, err := s.prRepo.Exists(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("check pr exists: %w", err)
	}
	if exists {
		return nil, domain.ErrPRExists
	}

	author, err := s.userRepo.GetByID(ctx, authorID)
	if err != nil {
		return nil, fmt.Errorf("get author: %w", err)
	}

	candidates, err := s.userRepo.GetActiveTeamMembers(ctx, author.TeamName, authorID)
	if err != nil {
		return nil, fmt.Errorf("get team members: %w", err)
	}

	// Randomly select up to 2 reviewers
	reviewers := s.selectReviewers(candidates, 2)

	pr := &domain.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            domain.PRStatusOpen,
		AssignedReviewers: reviewers,
	}

	if err := s.prRepo.Create(ctx, pr); err != nil {
		return nil, fmt.Errorf("create pr: %w", err)
	}

	s.logger.Info("pr created",
		"pr_id", prID,
		"author_id", authorID,
		"reviewers_count", len(reviewers),
	)

	created, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("get created pr: %w", err)
	}

	return created, nil
}

// selectReviewers randomly selects reviewers from a pool of candidates.
// Uses Fisher-Yates shuffle for uniform distribution.
func (s *PRService) selectReviewers(candidates []*domain.User, maxCount int) []string {
	if len(candidates) == 0 {
		return []string{}
	}

	count := min(maxCount, len(candidates))

	// Create a copy to avoid modifying the original slice
	shuffled := make([]*domain.User, len(candidates))
	copy(shuffled, candidates)

	// Thread-safe shuffle
	s.mu.Lock()
	s.random.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	s.mu.Unlock()

	// Extract IDs from shuffled candidates
	reviewers := make([]string, count)
	for i := 0; i < count; i++ {
		reviewers[i] = shuffled[i].UserID
	}

	return reviewers
}

// MergePR marks a pull request as merged. Idempotent operation.
func (s *PRService) MergePR(ctx context.Context, prID string) (*domain.PullRequest, error) {
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("get pr: %w", err)
	}

	if pr.Status == domain.PRStatusMerged {
		s.logger.Info("pr already merged, returning current state",
			"pr_id", prID,
			"merged_at", pr.MergedAt,
		)
		return pr, nil
	}

	if err := s.prRepo.Merge(ctx, prID); err != nil {
		return nil, fmt.Errorf("merge pr: %w", err)
	}

	merged, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("get merged pr: %w", err)
	}

	s.logger.Info("pr merged successfully",
		"pr_id", prID,
		"merged_at", merged.MergedAt,
		"reviewers_count", len(merged.AssignedReviewers),
	)

	return merged, nil
}

// ReassignReviewer replaces an assigned reviewer with a new random team member.
// Returns the updated PR and new reviewer ID.
func (s *PRService) ReassignReviewer(ctx context.Context, prID, oldUserID string) (*domain.PullRequest, string, error) {
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, "", fmt.Errorf("get pr: %w", err)
	}

	if pr.Status == domain.PRStatusMerged {
		return nil, "", domain.ErrPRMerged
	}

	if !s.isAssigned(pr.AssignedReviewers, oldUserID) {
		return nil, "", domain.ErrNotAssigned
	}

	oldUser, err := s.userRepo.GetByID(ctx, oldUserID)
	if err != nil {
		return nil, "", fmt.Errorf("get old reviewer: %w", err)
	}

	// Get all active team members
	candidates, err := s.userRepo.GetActiveTeamMembers(ctx, oldUser.TeamName, "")
	if err != nil {
		return nil, "", fmt.Errorf("get team members: %w", err)
	}

	// Build exclusion list: author + current reviewers
	excluded := make(map[string]bool)
	excluded[pr.AuthorID] = true
	for _, reviewerID := range pr.AssignedReviewers {
		excluded[reviewerID] = true
	}

	// Filter eligible candidates
	eligible := make([]*domain.User, 0)
	for _, candidate := range candidates {
		if !excluded[candidate.UserID] {
			eligible = append(eligible, candidate)
		}
	}

	if len(eligible) == 0 {
		return nil, "", domain.ErrNoCandidate
	}

	// Select random replacement reviewer
	s.mu.Lock()
	newReviewer := eligible[s.random.Intn(len(eligible))]
	s.mu.Unlock()

	// Atomically replace reviewer in database
	if err := s.prRepo.ReplaceReviewer(ctx, prID, oldUserID, newReviewer.UserID); err != nil {
		// Handle concurrent merge scenario
		if errors.Is(err, domain.ErrNotAssigned) {
			checkPR, checkErr := s.prRepo.GetByID(ctx, prID)
			if checkErr == nil && checkPR.Status == domain.PRStatusMerged {
				return nil, "", domain.ErrPRMerged
			}
			return nil, "", domain.ErrNotAssigned
		}
		return nil, "", fmt.Errorf("replace reviewer: %w", err)
	}

	updated, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, "", fmt.Errorf("get updated pr: %w", err)
	}

	s.logger.Info("reviewer reassigned",
		"pr_id", prID,
		"old_reviewer", oldUserID,
		"new_reviewer", newReviewer.UserID,
		"team", oldUser.TeamName,
	)

	return updated, newReviewer.UserID, nil
}

// isAssigned checks if a user is in the reviewers list.
func (s *PRService) isAssigned(reviewers []string, userID string) bool {
	for _, id := range reviewers {
		if id == userID {
			return true
		}
	}
	return false
}
