package service

import (
	"context"
	"fmt"
	"github.com/ZertGraf/avito-test/internal/domain"
	"github.com/ZertGraf/avito-test/internal/pkg/logger"
	"github.com/ZertGraf/avito-test/internal/repository"
)

type UserService struct {
	repo   repository.UserRepository
	logger *logger.Logger
}

func NewUserService(repo repository.UserRepository, logger *logger.Logger) *UserService {
	return &UserService{
		repo:   repo,
		logger: logger,
	}
}

type SetIsActiveResponse struct {
	User *domain.User `json:"user"`
}

// SetIsActive updates user's activity status.
// Used to enable/disable users from reviewer assignment pool.
func (s *UserService) SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	user, err := s.repo.SetIsActive(ctx, userID, isActive)
	if err != nil {
		return nil, fmt.Errorf("set is_active: %w", err)
	}

	s.logger.Info("user activity status changed",
		"user_id", userID,
		"is_active", isActive,
		"team", user.TeamName,
	)

	return user, nil
}
