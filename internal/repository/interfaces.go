package repository

import (
	"context"
	"github.com/ZertGraf/avito-test/internal/domain"
)

// TeamRepository - интерфейс для работы с командами
type TeamRepository interface {
	TeamExists(ctx context.Context, teamName string) (bool, error)
	CreateTeamWithMembers(ctx context.Context, team *domain.Team) (*domain.Team, error)
	GetTeamWithMembers(ctx context.Context, teamName string) (*domain.Team, error)
}

// UserRepository - интерфейс для работы с пользователями
type UserRepository interface {
	GetByID(ctx context.Context, userID string) (*domain.User, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
	GetActiveTeamMembers(ctx context.Context, teamName string, excludeUserID string) ([]*domain.User, error)
}

// PRRepository - интерфейс для работы с PR
type PRRepository interface {
	Create(ctx context.Context, pr *domain.PullRequest) error
	GetByID(ctx context.Context, prID string) (*domain.PullRequest, error)
	Merge(ctx context.Context, prID string) error
	GetByReviewer(ctx context.Context, userID string) ([]*domain.PullRequestShort, error)
	ReplaceReviewer(ctx context.Context, prID, oldUserID, newUserID string) error
	Exists(ctx context.Context, prID string) (bool, error)
}
