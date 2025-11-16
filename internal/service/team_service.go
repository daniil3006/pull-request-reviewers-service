package service

import (
	"context"
	"errors"
	"pull-request-reviewers-service/internal/models"
	"pull-request-reviewers-service/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type TeamService struct {
	r *repository.TeamRepository
}

func NewTeamService(r *repository.TeamRepository) *TeamService {
	return &TeamService{r}
}

func (s *TeamService) CreateTeam(ctx context.Context, team models.Team) (models.Team, error) {
	tx, err := s.r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return models.Team{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	err = s.r.CreateTeam(ctx, tx, team.Name)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.Team{}, models.ErrTeamExist
		}
		return models.Team{}, err
	}
	for _, user := range team.Members {
		err = s.r.CreateUpdateUser(ctx, tx, user, team.Name)
		if err != nil {
			return models.Team{}, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return models.Team{}, err
	}

	createdTeam, err := s.r.GetTeam(ctx, team.Name)
	if err != nil {
		return models.Team{}, err
	}

	return createdTeam, nil
}

func (s *TeamService) GetTeam(ctx context.Context, teamName string) (models.Team, error) {
	team, err := s.r.GetTeam(ctx, teamName)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Team{}, models.ErrTeamNotFound
	}
	return team, nil
}

func (s *TeamService) SetIsActive(ctx context.Context, userID string, isActive bool) (models.User, error) {
	user, err := s.r.SetIsActiveUser(ctx, userID, isActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, models.ErrUserNotFound
	}
	return user, nil
}

func (s *TeamService) GetPRsByReviewer(ctx context.Context, reviewerID string) ([]models.PullRequestShort, error) {
	pullRequests, err := s.r.GetPRsByReviewer(ctx, reviewerID)
	if err != nil {
		return nil, err
	}
	return pullRequests, nil
}
