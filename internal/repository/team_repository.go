package repository

import (
	"context"
	"errors"
	"pull-request-reviewers-service/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TeamRepository struct {
	DB *pgxpool.Pool
}

func NewTeamRepository(DB *pgxpool.Pool) *TeamRepository {
	return &TeamRepository{DB: DB}
}

func (r *TeamRepository) CreateTeam(ctx context.Context, tx pgx.Tx, teamName string) error {
	_, err := tx.Exec(ctx, `INSERT INTO teams (team_name) VALUES ($1) RETURNING team_name`, teamName)
	if err != nil {
		return err
	}
	return nil
}

func (r *TeamRepository) CreateUpdateUser(ctx context.Context, tx pgx.Tx, member models.TeamMember, teamName string) error {
	var exist bool
	err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT user_id FROM users WHERE user_id = $1)", member.Id).Scan(&exist)
	if exist {
		_, err = tx.Exec(ctx, `UPDATE users 
SET username = $1, team_name = $2, is_active = $3 
WHERE user_id = $4`, member.Username, teamName, member.IsActive, member.Id)
	} else {
		_, err = tx.Exec(ctx, `INSERT INTO users (user_id, username, team_name, is_active)
VALUES ($1, $2, $3, $4)`, member.Id, member.Username, teamName, member.IsActive)
	}
	if err != nil {
		return err
	}

	return nil
}

func (r *TeamRepository) GetUser(ctx context.Context, userID string) (models.User, error) {
	var u models.User
	err := r.DB.QueryRow(ctx, `SELECT user_id, username, team_name, is_active
FROM users
WHERE user_id = $1`, userID).Scan(&u.Id, &u.Username, &u.TeamName, &u.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, models.ErrUserNotFound
		}
		return models.User{}, err
	}
	return u, nil
}

func (r *TeamRepository) GetTeam(ctx context.Context, name string) (models.Team, error) {
	querySelectTeamName := `SELECT team_name FROM teams WHERE team_name = $1`
	row := r.DB.QueryRow(ctx, querySelectTeamName, name)
	if err := row.Scan(&name); err != nil {
		return models.Team{}, err
	}

	var members []models.TeamMember
	querySelectTeamMembers := `SELECT user_id, username, is_active 
FROM users 
WHERE team_name = $1`
	rows, err := r.DB.Query(ctx, querySelectTeamMembers, name)
	if err != nil {
		return models.Team{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var member models.TeamMember
		if err = rows.Scan(&member.Id, &member.Username, &member.IsActive); err != nil {
			return models.Team{}, err
		}
		members = append(members, member)
	}
	team := models.Team{
		Name:    name,
		Members: members,
	}
	return team, nil
}

func (r *TeamRepository) SetIsActiveUser(ctx context.Context, userID string, isActive bool) (models.User, error) {
	var user models.User
	err := r.DB.QueryRow(ctx, `UPDATE users 
SET is_active = $1 
WHERE user_id = $2 
RETURNING user_id, username, team_name, is_active`, isActive, userID).Scan(&user.Id, &user.Username, &user.TeamName, &user.IsActive)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}
func (r *TeamRepository) GetPRsByReviewer(ctx context.Context, reviewerID string) ([]models.PullRequestShort, error) {
	rows, err := r.DB.Query(ctx, `SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
FROM pull_requests pr
JOIN reviewers r
ON pr.pull_request_id = r.pull_request_id
WHERE r.reviewer_id = $1`, reviewerID)
	if err != nil {
		return nil, err
	}

	var pullRequestsShort []models.PullRequestShort
	defer rows.Close()

	for rows.Next() {
		var prShort models.PullRequestShort
		err = rows.Scan(&prShort.Id, &prShort.Name, &prShort.AuthorID, &prShort.Status)
		if err != nil {
			return nil, err
		}
		pullRequestsShort = append(pullRequestsShort, prShort)
	}

	return pullRequestsShort, nil
}
