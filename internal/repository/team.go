package repository

import (
	"context"
	"fmt"
	"github.com/ZertGraf/avito-test/internal/domain"
	"github.com/ZertGraf/avito-test/internal/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Team struct {
	db     *pgxpool.Pool
	logger *logger.Logger
}

func NewTeamRepo(db *pgxpool.Pool, logger *logger.Logger) *Team {
	return &Team{
		db:     db,
		logger: logger.Component("repository/postgres"),
	}
}

// TeamExists проверяет существование команды
func (r *Team) TeamExists(ctx context.Context, teamName string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`

	err := r.db.QueryRow(ctx, query, teamName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check team exists: %w", err)
	}

	return exists, nil
}

// CreateTeamWithMembers создает команду и участников в одной транзакции
func (r *Team) CreateTeamWithMembers(ctx context.Context, team *domain.Team) (*domain.Team, error) {
	err := r.withTx(ctx, func(tx pgx.Tx) error {
		// 1. Создаем команду
		_, err := tx.Exec(ctx,
			`INSERT INTO teams (team_name, created_at) VALUES ($1, NOW())`,
			team.TeamName,
		)
		if err != nil {
			return fmt.Errorf("insert team: %w", err)
		}

		// 2. Upsert участников
		for _, member := range team.Members {
			_, err := tx.Exec(ctx, `
				INSERT INTO users (user_id, username, team_name, is_active, created_at)
				VALUES ($1, $2, $3, $4, NOW())
				ON CONFLICT (user_id) 
				DO UPDATE SET 
					username = EXCLUDED.username,
					team_name = EXCLUDED.team_name,
					is_active = EXCLUDED.is_active
			`,
				member.UserID,
				member.Username,
				team.TeamName,
				member.IsActive,
			)
			if err != nil {
				return fmt.Errorf("upsert user %s: %w", member.UserID, err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return team, nil
}

// GetTeamWithMembers получает команду со всеми участниками
func (r *Team) GetTeamWithMembers(ctx context.Context, teamName string) (*domain.Team, error) {
	// Используем LEFT JOIN на случай если команда пустая
	query := `
        SELECT 
            t.team_name,
            COALESCE(u.user_id, '') as user_id,
            COALESCE(u.username, '') as username,
            COALESCE(u.is_active, false) as is_active
        FROM teams t
        LEFT JOIN users u ON u.team_name = t.team_name
        WHERE t.team_name = $1
        ORDER BY u.user_id
    `

	rows, err := r.db.Query(ctx, query, teamName)
	if err != nil {
		return nil, fmt.Errorf("query team: %w", err)
	}
	defer rows.Close()

	var team *domain.Team
	for rows.Next() {
		var (
			tName    string
			userID   string
			username string
			isActive bool
		)

		if err := rows.Scan(&tName, &userID, &username, &isActive); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		if team == nil {
			team = &domain.Team{
				TeamName: tName,
				Members:  []domain.TeamMember{},
			}
		}

		// Пропускаем пустые записи (когда у команды нет участников)
		if userID != "" {
			team.Members = append(team.Members, domain.TeamMember{
				UserID:   userID,
				Username: username,
				IsActive: isActive,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	if team == nil {
		return nil, domain.ErrTeamNotFound
	}

	return team, nil
}

// withTx выполняет функцию в транзакции
func (r *Team) withTx(ctx context.Context, fn func(pgx.Tx) error) error {
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
