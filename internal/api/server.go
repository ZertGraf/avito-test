package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/ZertGraf/avito-test/internal/api/handler"
	"github.com/ZertGraf/avito-test/internal/api/middleware"
	"github.com/ZertGraf/avito-test/internal/pkg/logger"
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"
)

type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type HTTPServer struct {
	server *http.Server
	config *ServerConfig
	logger *logger.Logger
}

func NewHTTPServer(config *ServerConfig,
	teamHandler *handler.TeamHandler,
	userHandler *handler.UserHandler,
	prHandler *handler.PRHandler,
	logger *logger.Logger) *HTTPServer {

	router := setupRouter(teamHandler, userHandler, prHandler, logger)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler:      router,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	return &HTTPServer{
		server: server,
		config: config,
		logger: logger.Component("http"),
	}
}

func (s *HTTPServer) Start(_ context.Context) error {
	s.logger.Info("starting http servers",
		"public_addr", s.server.Addr)

	go func() {
		s.logger.Info("public server listening", "addr", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("public server failed", "error", err)
		}
	}()

	s.logger.Info("http servers started successfully")
	return nil
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	s.logger.Info("stopping http servers")
	// shutdown public server
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("public server shutdown failed", "error", err)
		return err
	}

	s.logger.Info("http servers stopped successfully")
	return nil
}

func setupRouter(
	teamHandler *handler.TeamHandler,
	userHandler *handler.UserHandler,
	prHandler *handler.PRHandler,
	logger *logger.Logger,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestLogger(logger))
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.Security())
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"healthy"}`)); err != nil {
			logger.Warn("failed to write health response", "error", err)
		}

	})

	r.Mount("/team", teamHandler.Routes())
	r.Mount("/pullRequest", prHandler.Routes())
	r.Mount("/users", userHandler.Routes())

	return r
}
