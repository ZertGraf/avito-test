package handler

import (
	"encoding/json"
	"github.com/ZertGraf/avito-test/internal/domain"
	"github.com/ZertGraf/avito-test/internal/pkg/logger"
	"github.com/ZertGraf/avito-test/internal/service"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type UserHandler struct {
	userService *service.UserService
	prService   *service.PRService
	logger      *logger.Logger
}

func NewUserHandler(
	userService *service.UserService,
	prService *service.PRService,
	logger *logger.Logger,
) *UserHandler {
	return &UserHandler{
		userService: userService,
		prService:   prService,
		logger:      logger.Component("handler/user"),
	}
}

func (h *UserHandler) Routes() http.Handler {
	r := chi.NewRouter()

	r.Get("/health", healthCheck)
	r.Post("/setIsActive", h.SetIsActive)
	r.Get("/getReview", h.GetReview)

	return r
}

type SetIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type SetIsActiveResponse struct {
	User *domain.User `json:"user"`
}

func (h *UserHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var req SetIsActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("invalid request body", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userService.SetIsActive(r.Context(), req.UserID, req.IsActive) // ← вот тут
	if err != nil {
		WriteError(w, err, h.logger)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := SetIsActiveResponse{User: user}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

type GetReviewResponse struct {
	UserID       string                     `json:"user_id"`
	PullRequests []*domain.PullRequestShort `json:"pull_requests"`
}

func (h *UserHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		h.logger.Warn("user_id query parameter is required")
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	// вызываем PRService напрямую
	prs, err := h.prService.GetReviewsByUser(r.Context(), userID)
	if err != nil {
		WriteError(w, err, h.logger)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := GetReviewResponse{
		UserID:       userID,
		PullRequests: prs,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}
