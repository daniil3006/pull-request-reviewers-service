package repository

import (
	"context"
	"pull-request-reviewers-service/internal/models"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PullRequestRepository struct {
	db *pgxpool.Pool
}

func NewPullRequestRepository(db *pgxpool.Pool) *PullRequestRepository {
	return &PullRequestRepository{db: db}
}

func (r *PullRequestRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.db.BeginTx(ctx, pgx.TxOptions{})
}

func (r *PullRequestRepository) CreatePullRequest(ctx context.Context, tx pgx.Tx, pr models.PullRequest) error {
	_, err := tx.Exec(ctx, `INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at) 
VALUES ($1, $2, $3, $4, $5)`, pr.Id, pr.Name, pr.AuthorID, pr.Status, pr.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *PullRequestRepository) GetUsersTeam(ctx context.Context, userID string) (string, error) {
	var teamName string
	err := r.db.QueryRow(ctx, `SELECT team_name FROM users WHERE user_id = $1`, userID).Scan(&teamName)
	if err != nil {
		return "", err
	}
	return teamName, nil
}

func (r *PullRequestRepository) FindPotentialReviewers(ctx context.Context, teamName, authorID string) ([]string, error) {
	rows, err := r.db.Query(ctx, `SELECT user_id FROM users WHERE team_name = $1 AND user_id != $2 AND is_active IS TRUE`, teamName, authorID)
	if err != nil {
		return nil, err
	}
	var members []string
	defer rows.Close()

	for rows.Next() {
		var memberID string
		if err = rows.Scan(&memberID); err != nil {
			return nil, err
		}
		members = append(members, memberID)
	}
	return members, nil
}

func (r *PullRequestRepository) FindNewReviewer(ctx context.Context, tx pgx.Tx, teamName, authorID, prID string) ([]string, error) {
	rows, err := tx.Query(ctx, `SELECT user_id 
FROM users 
WHERE team_name = $1 AND user_id != $2 AND is_active IS TRUE AND user_id NOT IN (
SELECT reviewer_id FROM reviewers WHERE pull_request_id = $3)`, teamName, authorID, prID)
	if err != nil {
		return nil, err
	}

	var members []string
	defer rows.Close()

	for rows.Next() {
		var memberID string
		if err = rows.Scan(&memberID); err != nil {
			return nil, err
		}
		members = append(members, memberID)
	}
	return members, nil
}

func (r *PullRequestRepository) UpdateReviewer(ctx context.Context, tx pgx.Tx, prID, newReviewerID, oldReviewerID string) error {
	var id string
	err := tx.QueryRow(ctx, `UPDATE reviewers 
SET reviewer_id = $1 
WHERE pull_request_id = $2 AND reviewer_id = $3 RETURNING reviewer_id`, newReviewerID, prID, oldReviewerID).Scan(&id)
	if err != nil {
		return err
	}
	return nil
}

func (r *PullRequestRepository) AddReviewers(ctx context.Context, tx pgx.Tx, prID string, reviewersID []string) error {
	for _, reviewerID := range reviewersID {
		_, err := tx.Exec(ctx, `INSERT INTO reviewers (pull_request_id, reviewer_id) VALUES ($1, $2)`, prID, reviewerID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PullRequestRepository) MergePullRequest(ctx context.Context, tx pgx.Tx, prID string, mergedAt time.Time) error {
	var id string
	err := tx.QueryRow(ctx, `UPDATE pull_requests SET status = $1, merged_at = $2 
WHERE pull_request_id = $3 RETURNING pull_request_id`, "MERGED", mergedAt, prID).Scan(&id)
	if err != nil {
		return err
	}
	return nil
}

func (r *PullRequestRepository) GetPullRequest(ctx context.Context, tx pgx.Tx, prID string) (models.PullRequest, error) {
	var pr models.PullRequest
	err := tx.QueryRow(ctx, `SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
FROM pull_requests
WHERE pull_request_id = $1`, prID).Scan(&pr.Id, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
	if err != nil {
		return models.PullRequest{}, err
	}

	rows, err := tx.Query(ctx, `SELECT reviewer_id FROM reviewers WHERE pull_request_id = $1`, pr.Id)
	if err != nil {
		return models.PullRequest{}, err
	}

	for rows.Next() {
		var reviewerID string
		err = rows.Scan(&reviewerID)
		if err != nil {
			return models.PullRequest{}, err
		}
		pr.AssignedReviewers = append(pr.AssignedReviewers, reviewerID)
	}

	return pr, nil
}

func (r *PullRequestRepository) GetAssignStat(ctx context.Context) ([]models.ReviewerStat, error) {
	rows, err := r.db.Query(ctx, `SELECT reviewer_id, COUNT(*) AS count FROM reviewers GROUP BY reviewer_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stat []models.ReviewerStat
	for rows.Next() {
		var st models.ReviewerStat
		if err = rows.Scan(&st.ReviewerID, &st.AssignStat); err != nil {
			return nil, err
		}
		stat = append(stat, st)
	}
	return stat, nil
}
