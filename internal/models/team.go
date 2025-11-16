package models

import "errors"

type Team struct {
	Name    string       `json:"team_name"`
	Members []TeamMember `json:"members"`
}

type TeamMember struct {
	Id       string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type TeamResponse struct {
	Team Team `json:"team"`
}

var ErrTeamExist = errors.New("team already exist")
var ErrTeamNotFound = errors.New("team not found")
