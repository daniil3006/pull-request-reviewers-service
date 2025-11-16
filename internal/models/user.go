package models

import "errors"

type User struct {
	Id       string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type UserResponse struct {
	User User `json:"user"`
}

var ErrUserNotFound = errors.New("user not found")
var ErrAuthorNotFound = errors.New("author not found")
var ErrNotEnoughMembersInTeam = errors.New("not enough members in team")
