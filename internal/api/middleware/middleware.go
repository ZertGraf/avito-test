package middleware

import (
	"encoding/json"
	"github.com/ZertGraf/avito-test/internal/api/handler"
	"github.com/ZertGraf/avito-test/internal/pkg/logger"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// RequestLogger creates HTTP request logging middleware
func RequestLogger(logger *logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}

// Security adds basic security headers
func Security() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// for file downloads, allow content sniffing
			if r.URL.Path != "/health" && r.Method == "GET" {
				w.Header().Del("X-Content-Type-Options")
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Recovery recovers from panics and logs them
func Recovery(logger *logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("panic recovered",
						"error", err,
						"stack", string(debug.Stack()),
					)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					if err := json.NewEncoder(w).Encode(handler.ErrorResponse{
						Error: handler.ErrorDetail{
							Code:    "INTERNAL_ERROR",
							Message: "internal server error",
						},
					}); err != nil {
						logger.Warn("failed to write recovered response", "error", err)
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// Timeout adds request timeout
func Timeout(timeout time.Duration) func(next http.Handler) http.Handler {
	return middleware.Timeout(timeout)
}
