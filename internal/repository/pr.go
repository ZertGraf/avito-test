package repository

import (
	"context"
	"fmt"
	"github.com/ZertGraf/avito-test/internal/domain"
	"github.com/ZertGraf/avito-test/internal/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PRRepo struct {
	db     *pgxpool.Pool
	logger *logger.Logger
}

func NewPRRepo(db *pgxpool.Pool, logger *logger.Logger) *PRRepo {
	return &PRRepo{
		db:     db,
		logger: logger.Component("repository/pr"),
	}
}

// Create persists a new pull request and its assigned reviewers.
// Uses transaction to ensure atomicity.
func (r *PRRepo) Create(ctx context.Context, pr *domain.PullRequest) error {
	return r.withTx(ctx, func(tx pgx.Tx) error {
		// Insert PR record
		_, err := tx.Exec(ctx, `
            INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status)
            VALUES ($1, $2, $3, $4)
        `, pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.Status)

		if err != nil {
			return fmt.Errorf("insert pr: %w", err)
		}

		// Insert reviewer assignments
		for _, reviewerID := range pr.AssignedReviewers {
			_, err := tx.Exec(ctx, `
                INSERT INTO pr_reviewers (pull_request_id, reviewer_id)
                VALUES ($1, $2)
            `, pr.PullRequestID, reviewerID)

			if err != nil {
				return fmt.Errorf("insert reviewer %s: %w", reviewerID, err)
			}
		}

		return nil
	})
}

// GetByID retrieves a pull request with all assigned reviewers.
// Returns ErrPRNotFound if PR doesn't exist.
func (r *PRRepo) GetByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	query := `
        SELECT 
            pr.pull_request_id,
            pr.pull_request_name,
            pr.author_id,
            pr.status,
            pr.created_at,
            pr.merged_at,
            COALESCE(r.reviewer_id, '') as reviewer_id
        FROM pull_requests pr
        LEFT JOIN pr_reviewers r ON pr.pull_request_id = r.pull_request_id
        WHERE pr.pull_request_id = $1
        ORDER BY r.assigned_at
    `

	rows, err := r.db.Query(ctx, query, prID)
	if err != nil {
		return nil, fmt.Errorf("query pr: %w", err)
	}
	defer rows.Close()

	var pr *domain.PullRequest
	for rows.Next() {
		var reviewerID string

		// Initialize PR on first row
		if pr == nil {
			pr = &domain.PullRequest{}
			err = rows.Scan(
				&pr.PullRequestID,
				&pr.PullRequestName,
				&pr.AuthorID,
				&pr.Status,
				&pr.CreatedAt,
				&pr.MergedAt,
				&reviewerID,
			)
		} else {
			var tmpPR domain.PullRequest

			// Subsequent rows only need reviewer ID
			err = rows.Scan(
				&tmpPR.PullRequestID,
				&tmpPR.PullRequestName,
				&tmpPR.AuthorID,
				&tmpPR.Status,
				&tmpPR.CreatedAt,
				&tmpPR.MergedAt,
				&reviewerID,
			)
		}

		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		// Collect reviewer IDs
		if reviewerID != "" {
			pr.AssignedReviewers = append(pr.AssignedReviewers, reviewerID)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	if pr == nil {
		return nil, domain.ErrPRNotFound
	}

	// Ensure non-nil slice for consistency
	if pr.AssignedReviewers == nil {
		pr.AssignedReviewers = []string{}
	}

	return pr, nil
}

// Merge marks a pull request as merged with current timestamp.
func (r *PRRepo) Merge(ctx context.Context, prID string) error {
	query := `
        UPDATE pull_requests 
        SET status = $1, merged_at = NOW()
        WHERE pull_request_id = $2
    `

	result, err := r.db.Exec(ctx, query, domain.PRStatusMerged, prID)
	if err != nil {
		return fmt.Errorf("update pr: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrPRNotFound
	}

	return nil
}

// GetByReviewer retrieves all PRs assigned to a specific reviewer.
// Returns empty slice if no PRs found.
func (r *PRRepo) GetByReviewer(ctx context.Context, userID string) ([]*domain.PullRequestShort, error) {
	query := `
		SELECT 
			pr.pull_request_id,
			pr.pull_request_name,
			pr.author_id,
			pr.status
		FROM pull_requests pr
		INNER JOIN pr_reviewers r ON pr.pull_request_id = r.pull_request_id
		WHERE r.reviewer_id = $1
		ORDER BY pr.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query prs: %w", err)
	}
	defer rows.Close()

	var prs []*domain.PullRequestShort
	for rows.Next() {
		pr := &domain.PullRequestShort{}
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status); err != nil {
			return nil, fmt.Errorf("scan pr: %w", err)
		}
		prs = append(prs, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	// Return empty slice instead of nil
	if prs == nil {
		prs = []*domain.PullRequestShort{}
	}

	return prs, nil
}

// ReplaceReviewer atomically replaces a reviewer on an open PR.
// Ensures PR is still open and reviewer is assigned before replacement.
func (r *PRRepo) ReplaceReviewer(ctx context.Context, prID, oldUserID, newUserID string) error {
	query := `
        UPDATE pr_reviewers
        SET reviewer_id = $1
        WHERE pull_request_id = $2 
          AND reviewer_id = $3
          AND EXISTS (
              SELECT 1 FROM pull_requests 
              WHERE pull_request_id = $2 AND status = 'OPEN')
    `

	result, err := r.db.Exec(ctx, query, newUserID, prID, oldUserID)
	if err != nil {
		return fmt.Errorf("replace reviewer: %w", err)
	}

	// No rows affected means either reviewer not assigned or PR not open
	if result.RowsAffected() == 0 {
		return domain.ErrNotAssigned
	}

	return nil
}

// Exists checks if a pull request exists by ID.
func (r *PRRepo) Exists(ctx context.Context, prID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)`

	err := r.db.QueryRow(ctx, query, prID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check pr exists: %w", err)
	}

	return exists, nil
}

// withTx executes a function within a database transaction.
// Automatically handles commit/rollback based on error status.
func (r *PRRepo) withTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				r.logger.Error("failed to rollback transaction",
					"error", rbErr,
					"original_error", err,
				)
			}
		}
	}()

	if err = fn(tx); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
