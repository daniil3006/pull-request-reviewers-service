package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"pull-request-reviewers-service/internal/models"
	"pull-request-reviewers-service/internal/service"
)

type PullRequestHandler struct {
	s *service.PullRequestService
}

func NewPullRequestHandler(s *service.PullRequestService) *PullRequestHandler {
	return &PullRequestHandler{s: s}
}
func (h *PullRequestHandler) CreatePullRequest(w http.ResponseWriter, r *http.Request) {
	var prShort models.PullRequestShort
	err := json.NewDecoder(r.Body).Decode(&prShort)
	if err != nil {
		writeHTTPError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON")
		return
	}

	pullRequest, err := h.s.CreatePullRequest(r.Context(), prShort)
	if err != nil {
		if errors.Is(err, models.ErrAuthorNotFound) {
			writeHTTPError(w, http.StatusNotFound, "NOT_FOUND", "author not found")
			return
		}
		if errors.Is(err, models.ErrPullRequestExist) {
			writeHTTPError(w, http.StatusConflict, "PR_EXISTS", "pull request already exists")
			return
		}

		writeHTTPError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}

	prResp := models.PullRequestResponse{PullRequest: pullRequest}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(prResp)
}

func (h *PullRequestHandler) MergePullRequest(w http.ResponseWriter, r *http.Request) {
	var prID struct {
		PullRequestID string `json:"pull_request_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&prID)
	if err != nil {
		writeHTTPError(w, http.StatusBadRequest, "BAD_JSON", "internal JSON")
		return
	}

	pullRequest, err := h.s.MergePullRequest(r.Context(), prID.PullRequestID)
	if err != nil {
		if errors.Is(err, models.ErrPullRequestNotFound) {
			writeHTTPError(w, http.StatusNotFound, "NOT_FOUND", "pull request not found")
			return
		}
		writeHTTPError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(pullRequest)
}

func (h *PullRequestHandler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	var reassign struct {
		PullRequestID string `json:"pull_request_id"`
		OldUserID     string `json:"old_user_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&reassign)
	if err != nil {
		writeHTTPError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON")
		return
	}

	pullRequest, newReviewerID, err := h.s.ReassignReviewer(r.Context(), reassign.PullRequestID, reassign.OldUserID)
	if err != nil {
		if errors.Is(err, models.ErrPullRequestNotFound) {
			writeHTTPError(w, http.StatusNotFound, "NOT_FOUND", "pull request not found")
			return
		}
		if errors.Is(err, models.ErrUserNotFound) {
			writeHTTPError(w, http.StatusNotFound, "NOT_FOUND", "user not found")
			return
		}
		if errors.Is(err, models.ErrPullRequestAlreadyMerged) {
			writeHTTPError(w, http.StatusConflict, "PR_MERGED", "cannot reassign on merged PR")
			return
		}
		if errors.Is(err, models.ErrUserNotReviewer) {
			writeHTTPError(w, http.StatusConflict, "NOT_ASSIGNED", "reviewer is not assigned to this PR")
			return
		}
		if errors.Is(err, models.ErrNotEnoughMembersInTeam) {
			writeHTTPError(w, http.StatusConflict, "NO_CANDIDATE", "no active replacement candidate in team")
			return
		}
		writeHTTPError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	reassignResp := models.ReassignResponse{
		PullRequest: pullRequest,
		ReplacedBy:  newReviewerID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(reassignResp)
}

func (h *PullRequestHandler) GetAssignStat(w http.ResponseWriter, r *http.Request) {
	stat, err := h.s.GetAssignStat(r.Context())
	if err != nil {
		writeHTTPError(w, http.StatusInternalServerError, "INTERNAL", "server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(stat)
}
