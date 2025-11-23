package handler

import (
	"encoding/json"
	"github.com/ZertGraf/avito-test/internal/domain"
	"github.com/ZertGraf/avito-test/internal/pkg/logger"
	"github.com/ZertGraf/avito-test/internal/service"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type TeamHandler struct {
	teamService *service.TeamService
	logger      *logger.Logger
}

func NewTeamHandler(teamService *service.TeamService, logger *logger.Logger) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
		logger:      logger,
	}
}

func (h *TeamHandler) Routes() http.Handler {
	r := chi.NewRouter()

	// health check
	r.Get("/health", healthCheck)

	r.Post("/add", h.CreateTeam)
	r.Get("/get", h.GetTeam)
	return r
}

func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var team domain.Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		h.logger.Warn("invalid request body", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	res, err := h.teamService.CreateTeam(r.Context(), &team)
	if err != nil {
		WriteError(w, err, h.logger)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		h.logger.Error("failed to encode response", "error", err)
		return
	}
	return
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		h.logger.Warn("team_name query parameter is required")
		http.Error(w, "team_name is required", http.StatusBadRequest)
		return
	}

	team, err := h.teamService.GetTeam(r.Context(), teamName)
	if err != nil {
		WriteError(w, err, h.logger)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(team); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"healthy","service":"auth-service"}`))
}
