package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/ZertGraf/avito-test/internal/domain"
	"github.com/ZertGraf/avito-test/internal/pkg/logger"
	"github.com/ZertGraf/avito-test/internal/repository"
	. "github.com/go-ozzo/ozzo-validation"
)

type TeamService struct {
	repo   repository.TeamRepository
	logger *logger.Logger
}

func NewTeamService(teamRepo repository.TeamRepository, logger *logger.Logger) *TeamService {
	return &TeamService{
		repo:   teamRepo,
		logger: logger,
	}
}

type CreateTeamResponse struct {
	Team *domain.Team `json:"team"`
}

func (s *TeamService) CreateTeam(ctx context.Context, team *domain.Team) (*CreateTeamResponse, error) {
	if err := s.validateTeam(team); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	exists, err := s.repo.TeamExists(ctx, team.TeamName)
	if err != nil {
		return nil, fmt.Errorf("check team exists: %w", err)
	}

	if exists {
		return nil, domain.ErrTeamExists
	}

	created, err := s.repo.CreateTeamWithMembers(ctx, team)
	if err != nil {
		return nil, fmt.Errorf("create team: %w", err)
	}

	s.logger.Info("team created",
		"team_name", created.TeamName,
		"members_count", len(created.Members),
	)

	return &CreateTeamResponse{Team: created}, nil
}

func (s *TeamService) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	team, err := s.repo.GetTeamWithMembers(ctx, teamName)
	if err != nil {
		return nil, fmt.Errorf("get team: %w", err)
	}

	s.logger.Info("team retrieved",
		"team_name", teamName,
		"members_count", len(team.Members),
	)

	return team, nil
}

func (s *TeamService) validateTeam(team *domain.Team) error {
	if team == nil {
		return errors.New("team is nil")
	}

	return ValidateStruct(team,
		Field(&team.TeamName,
			Required,
			Length(1, 255),
		),
		Field(&team.Members,
			Required,
			Length(1, 0),
			Each(By(s.validateMember)),
		),
	)
}

func (s *TeamService) validateMember(value interface{}) error {
	member, ok := value.(domain.TeamMember)
	if !ok {
		return errors.New("invalid member type")
	}

	return ValidateStruct(&member,
		Field(&member.UserID,
			Required,
			Length(1, 255),
		),
		Field(&member.Username,
			Required,
			Length(1, 255),
		),
	)
}
