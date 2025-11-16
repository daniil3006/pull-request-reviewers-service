package service

import (
	"context"
	"errors"
	"math/rand/v2"
	"pull-request-reviewers-service/internal/models"
	"pull-request-reviewers-service/internal/repository"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	openPullRequest   = "OPEN"
	mergedPullRequest = "MERGED"
)

type PullRequestService struct {
	r        *repository.PullRequestRepository
	teamRepo *repository.TeamRepository
}

func NewPullRequestService(r *repository.PullRequestRepository, teamRepo *repository.TeamRepository) *PullRequestService {
	return &PullRequestService{r, teamRepo}
}

func (s *PullRequestService) CreatePullRequest(ctx context.Context, prShort models.PullRequestShort) (models.PullRequest, error) {
	pullRequest := models.PullRequest{
		Id:        prShort.Id,
		Name:      prShort.Name,
		AuthorID:  prShort.AuthorID,
		Status:    "OPEN",
		CreatedAt: time.Now(),
	}

	teamName, err := s.r.GetUsersTeam(ctx, prShort.AuthorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.PullRequest{}, models.ErrAuthorNotFound
		}
		return models.PullRequest{}, err
	}

	reviewers, err := s.r.FindPotentialReviewers(ctx, teamName, prShort.AuthorID)
	if err != nil {
		return models.PullRequest{}, err
	}

	if len(reviewers) > 2 {
		rand.Shuffle(len(reviewers), func(i, j int) {
			reviewers[i], reviewers[j] = reviewers[j], reviewers[i]
		})
		reviewers = reviewers[:2]
	}

	tx, err := s.r.BeginTx(ctx)
	if err != nil {
		return models.PullRequest{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	err = s.r.CreatePullRequest(ctx, tx, pullRequest)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.PullRequest{}, models.ErrPullRequestExist
		}
		return models.PullRequest{}, err
	}

	err = s.r.AddReviewers(ctx, tx, prShort.Id, reviewers)
	if err != nil {
		return models.PullRequest{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return models.PullRequest{}, err
	}

	for _, reviewer := range reviewers {
		pullRequest.AssignedReviewers = append(pullRequest.AssignedReviewers, reviewer)
	}

	return pullRequest, nil
}

func (s *PullRequestService) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (models.PullRequest, string, error) {
	tx, err := s.r.BeginTx(ctx)
	if err != nil {
		return models.PullRequest{}, "", err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	pullRequest, err := s.r.GetPullRequest(ctx, tx, prID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.PullRequest{}, "", models.ErrPullRequestNotFound
		}
		return models.PullRequest{}, "", err
	}

	if pullRequest.Status == mergedPullRequest {
		return models.PullRequest{}, "", models.ErrPullRequestAlreadyMerged
	}

	if !slices.Contains(pullRequest.AssignedReviewers, oldReviewerID) {
		return models.PullRequest{}, "", models.ErrUserNotReviewer
	}

	oldReviewer, err := s.teamRepo.GetUser(ctx, oldReviewerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.PullRequest{}, "", models.ErrUserNotFound
		}
		return models.PullRequest{}, "", err
	}

	reviewers, err := s.r.FindNewReviewer(ctx, tx, oldReviewer.TeamName, pullRequest.AuthorID, prID)
	if err != nil {
		return models.PullRequest{}, "", err
	}

	if len(reviewers) == 0 {
		return models.PullRequest{}, "", models.ErrNotEnoughMembersInTeam
	}

	newReviewerID := reviewers[rand.IntN(len(reviewers))]

	err = s.r.UpdateReviewer(ctx, tx, prID, newReviewerID, oldReviewerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.PullRequest{}, "", models.ErrUserNotReviewer
		}
		return models.PullRequest{}, "", err
	}

	pullRequest, err = s.r.GetPullRequest(ctx, tx, prID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.PullRequest{}, "", models.ErrPullRequestNotFound
		}
		return models.PullRequest{}, "", err
	}

	if err = tx.Commit(ctx); err != nil {
		return models.PullRequest{}, "", err
	}

	return pullRequest, newReviewerID, nil
}

func (s *PullRequestService) MergePullRequest(ctx context.Context, prID string) (models.PullRequest, error) {
	tx, err := s.r.BeginTx(ctx)
	if err != nil {
		return models.PullRequest{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	pullRequest, err := s.r.GetPullRequest(ctx, tx, prID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.PullRequest{}, models.ErrPullRequestNotFound
		}
		return models.PullRequest{}, err
	}

	if pullRequest.Status == mergedPullRequest {
		return pullRequest, nil
	}

	if pullRequest.Status == openPullRequest {
		mergedAt := time.Now()
		err = s.r.MergePullRequest(ctx, tx, prID, mergedAt)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return models.PullRequest{}, models.ErrPullRequestNotFound
			}
			return models.PullRequest{}, err
		}
	}

	updatedPr, err := s.r.GetPullRequest(ctx, tx, prID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.PullRequest{}, models.ErrPullRequestNotFound
		}
		return models.PullRequest{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return models.PullRequest{}, err
	}

	return updatedPr, nil
}

func (s *PullRequestService) GetAssignStat(ctx context.Context) ([]models.ReviewerStat, error) {
	return s.r.GetAssignStat(ctx)
}
