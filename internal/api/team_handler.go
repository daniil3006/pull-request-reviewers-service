package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"pull-request-reviewers-service/internal/models"
	"pull-request-reviewers-service/internal/service"
)

type TeamHandler struct {
	s *service.TeamService
}

func NewTeamHandler(s *service.TeamService) *TeamHandler {
	return &TeamHandler{s: s}
}

func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var team models.Team
	err := json.NewDecoder(r.Body).Decode(&team)
	if err != nil {
		writeHTTPError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON")
		return
	}
	team, err = h.s.CreateTeam(r.Context(), team)
	if err != nil {
		if errors.Is(err, models.ErrTeamExist) {
			writeHTTPError(w, http.StatusBadRequest, "TEAM_EXISTS", "team_name already exists")
			return
		}
		writeHTTPError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	teamResp := models.TeamResponse{Team: team}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(teamResp)
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	team, err := h.s.GetTeam(r.Context(), teamName)
	if err != nil {
		if errors.Is(err, models.ErrTeamNotFound) {
			writeHTTPError(w, http.StatusNotFound, "NOT_FOUND", "resource not found")
			return
		}
		writeHTTPError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(team)
}

func (h *TeamHandler) SetIsActiveUser(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}

	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		writeHTTPError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON")
		return
	}

	user, err := h.s.SetIsActive(r.Context(), reqBody.UserID, reqBody.IsActive)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			writeHTTPError(w, http.StatusNotFound, "NOT_FOUND", "user not found")
			return
		}

		writeHTTPError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	userResp := models.UserResponse{User: user}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(userResp)
}

func (h *TeamHandler) GetPRsByReviewer(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	pullRequests, err := h.s.GetPRsByReviewer(r.Context(), userID)
	if err != nil {
		writeHTTPError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	usersPRs := models.PullRequestsByReviewerResponse{
		UserID:       userID,
		PullRequests: pullRequests,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(usersPRs)
}

func writeHTTPError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(models.NewErrorResponse(code, message))
}
