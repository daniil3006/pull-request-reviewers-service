package models

import (
	"errors"
	"time"
)

type PullRequest struct {
	Id                string     `json:"pull_request_id"`
	Name              string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            string     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         time.Time  `json:"createdAt"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
}

type PullRequestShort struct {
	Id       string `json:"pull_request_id"`
	Name     string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
	Status   string `json:"status"`
}

type ReviewerStat struct {
	ReviewerID string `json:"reviewer_id"`
	AssignStat int    `json:"assign_stat"`
}

type PullRequestsByReviewerResponse struct {
	UserID       string             `json:"user_id"`
	PullRequests []PullRequestShort `json:"pull_requests"`
}

type ReassignResponse struct {
	PullRequest PullRequest `json:"pr"`
	ReplacedBy  string      `json:"replaced_by"`
}

type PullRequestResponse struct {
	PullRequest PullRequest `json:"pr"`
}

var ErrPullRequestExist = errors.New("pull request already exists")
var ErrPullRequestNotFound = errors.New("pull request not found")
var ErrPullRequestAlreadyMerged = errors.New("pull request already merged")
var ErrUserNotReviewer = errors.New("reviewer is not assigned to this PR")
