package handler

import (
	"encoding/json"
	"errors"
	"github.com/ZertGraf/avito-test/internal/domain"
	"github.com/ZertGraf/avito-test/internal/pkg/logger"
	"net/http"
)

// ErrorCode - коды из OpenAPI спецификации
type ErrorCode string

const (
	CodeTeamExists  ErrorCode = "TEAM_EXISTS"
	CodePRExists    ErrorCode = "PR_EXISTS"
	CodePRMerged    ErrorCode = "PR_MERGED"
	CodeNotAssigned ErrorCode = "NOT_ASSIGNED"
	CodeNoCandidate ErrorCode = "NO_CANDIDATE"
	CodeNotFound    ErrorCode = "NOT_FOUND"
)

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func WriteError(w http.ResponseWriter, err error, logger *logger.Logger) {
	status, response := mapError(err)

	if isDomainError(err) {
		logger.Warn("domain error",
			"error", err.Error(),
			"code", response.Error.Code,
		)
	} else {
		logger.Error("unexpected error",
			"error", err.Error(),
		)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func mapError(err error) (int, ErrorResponse) {
	switch {
	case errors.Is(err, domain.ErrTeamExists):
		return http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    CodeTeamExists,
				Message: err.Error(),
			},
		}

	case errors.Is(err, domain.ErrPRExists):
		return http.StatusConflict, ErrorResponse{
			Error: ErrorDetail{
				Code:    CodePRExists,
				Message: err.Error(),
			},
		}

	case errors.Is(err, domain.ErrPRMerged):
		return http.StatusConflict, ErrorResponse{
			Error: ErrorDetail{
				Code:    CodePRMerged,
				Message: err.Error(),
			},
		}

	case errors.Is(err, domain.ErrNotAssigned):
		return http.StatusConflict, ErrorResponse{
			Error: ErrorDetail{
				Code:    CodeNotAssigned,
				Message: err.Error(),
			},
		}

	case errors.Is(err, domain.ErrNoCandidate):
		return http.StatusConflict, ErrorResponse{
			Error: ErrorDetail{
				Code:    CodeNoCandidate,
				Message: err.Error(),
			},
		}

	case errors.Is(err, domain.ErrTeamNotFound),
		errors.Is(err, domain.ErrUserNotFound),
		errors.Is(err, domain.ErrPRNotFound):
		return http.StatusNotFound, ErrorResponse{
			Error: ErrorDetail{
				Code:    CodeNotFound,
				Message: err.Error(),
			},
		}

	default:
		return http.StatusInternalServerError, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "internal server error",
			},
		}
	}
}

func isDomainError(err error) bool {
	return errors.Is(err, domain.ErrTeamExists) ||
		errors.Is(err, domain.ErrTeamNotFound) ||
		errors.Is(err, domain.ErrUserNotFound) ||
		errors.Is(err, domain.ErrPRExists) ||
		errors.Is(err, domain.ErrPRNotFound) ||
		errors.Is(err, domain.ErrPRMerged) ||
		errors.Is(err, domain.ErrNotAssigned) ||
		errors.Is(err, domain.ErrNoCandidate)
}
