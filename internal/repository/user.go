package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/ZertGraf/avito-test/internal/domain"
	"github.com/ZertGraf/avito-test/internal/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db     *pgxpool.Pool
	logger *logger.Logger
}

func NewUserRepo(db *pgxpool.Pool, logger *logger.Logger) *UserRepo {
	return &UserRepo{
		db:     db,
		logger: logger.Component("repository/postgres"),
	}
}

func (r *UserRepo) SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	query := `
		UPDATE users 
		SET is_active = $1
		WHERE user_id = $2
		RETURNING user_id, username, team_name, is_active
	`

	var user domain.User
	err := r.db.QueryRow(ctx, query, isActive, userID).Scan(
		&user.UserID,
		&user.Username,
		&user.TeamName,
		&user.IsActive,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("update user: %w", err)
	}

	return &user, nil
}

func (r *UserRepo) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	query := `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE user_id = $1
	`

	var user domain.User
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.UserID,
		&user.Username,
		&user.TeamName,
		&user.IsActive,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	return &user, nil
}

func (r *UserRepo) GetActiveTeamMembers(ctx context.Context, teamName string, excludeUserID string) ([]*domain.User, error) {
	query := `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE team_name = $1 
		  AND user_id != $2
		  AND is_active = true
		ORDER BY user_id
	`

	rows, err := r.db.Query(ctx, query, teamName, excludeUserID)
	if err != nil {
		return nil, fmt.Errorf("query active members: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return users, nil
}
