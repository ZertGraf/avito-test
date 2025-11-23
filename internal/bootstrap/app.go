package bootstrap

import (
	"context"
	"fmt"
	"github.com/ZertGraf/avito-test/internal/api"
	"github.com/ZertGraf/avito-test/internal/api/handler"
	"github.com/ZertGraf/avito-test/internal/pkg/config"
	"github.com/ZertGraf/avito-test/internal/pkg/logger"
	"github.com/ZertGraf/avito-test/internal/pkg/postgres"
	"github.com/ZertGraf/avito-test/internal/repository"
	"github.com/ZertGraf/avito-test/internal/service"
)

type Application struct {
	Config   *config.Config
	Logger   *logger.Logger
	Postgres *postgres.Connection
	Migrator *postgres.Migrator

	TeamRepo repository.TeamRepository
	UserRepo repository.UserRepository
	PRRepo   repository.PRRepository

	TeamService *service.TeamService
	UserService *service.UserService
	PRService   *service.PRService

	TeamHandler *handler.TeamHandler
	UserHandler *handler.UserHandler
	PRHandler   *handler.PRHandler

	HTTPServer *api.HTTPServer
}

func New() (*Application, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	log, err := logger.New(&logger.Config{
		Level:     cfg.LogLevel,
		Format:    cfg.LogFormat,
		AddSource: cfg.LogAddSource,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	pg, err := postgres.New(log, &postgres.Config{
		Host:              cfg.DatabaseHost,
		Port:              cfg.DatabasePort,
		Username:          cfg.DatabaseUser,
		Password:          cfg.DatabasePassword,
		Database:          cfg.DatabaseName,
		Schema:            cfg.DatabaseSchema,
		SSLMode:           cfg.DatabaseSSLMode,
		MaxConns:          cfg.DatabaseMaxConns,
		MinConns:          cfg.DatabaseMinConns,
		MaxConnLifetime:   cfg.DatabaseMaxConnLifetime,
		MaxConnIdleTime:   cfg.DatabaseMaxConnIdleTime,
		HealthCheckPeriod: cfg.DatabaseHealthCheckPeriod,
		ConnectTimeout:    cfg.DatabaseConnectTimeout,
		AcquireTimeout:    cfg.DatabaseAcquireTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres connection: %w", err)
	}

	return &Application{
		Config:   cfg,
		Logger:   log,
		Postgres: pg,
	}, nil
}

func (app *Application) Init(ctx context.Context) error {
	app.Logger.Info("initializing application")

	if err := app.Postgres.Connect(ctx); err != nil {
		return fmt.Errorf("postgres connection failed: %w", err)
	}

	app.Migrator = postgres.NewMigrator(app.Postgres.Pool(), &postgres.MigrationConfig{
		Timeout:   app.Config.DatabaseMigrationTimeout,
		TableName: app.Config.DatabaseMigrationTable,
		Enabled:   app.Config.DatabaseMigrationEnabled,
	}, app.Logger)

	if err := app.Migrator.RunMigrations(ctx); err != nil {
		return fmt.Errorf("database migrations failed: %w", err)
	}

	app.TeamRepo = repository.NewTeamRepo(app.Postgres.Pool(), app.Logger)
	app.UserRepo = repository.NewUserRepo(app.Postgres.Pool(), app.Logger)
	app.PRRepo = repository.NewPRRepo(app.Postgres.Pool(), app.Logger)

	app.TeamService = service.NewTeamService(app.TeamRepo, app.Logger)
	app.UserService = service.NewUserService(app.UserRepo, app.Logger)
	app.PRService = service.NewPRService(app.PRRepo, app.UserRepo, app.Logger)

	app.TeamHandler = handler.NewTeamHandler(app.TeamService, app.Logger)
	app.UserHandler = handler.NewUserHandler(app.UserService, app.PRService, app.Logger)
	app.PRHandler = handler.NewPRHandler(app.PRService, app.Logger)

	serverConfig := &api.ServerConfig{
		Host:         app.Config.ServerHost,
		Port:         app.Config.ServerPort,
		ReadTimeout:  app.Config.ServerReadTimeout,
		WriteTimeout: app.Config.ServerWriteTimeout,
		IdleTimeout:  app.Config.ServerIdleTimeout,
	}

	app.HTTPServer = api.NewHTTPServer(
		serverConfig,
		app.TeamHandler,
		app.UserHandler,
		app.PRHandler,
		app.Logger,
	)

	if err := app.HTTPServer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start http server: %w", err)
	}

	app.Logger.Info("application initialized successfully")
	return nil
}

func (app *Application) Shutdown(ctx context.Context) error {
	app.Logger.Info("shutting down application")

	if app.HTTPServer != nil {
		if err := app.HTTPServer.Stop(ctx); err != nil {
			app.Logger.Error("error stopping http server", "error", err)
		}
	}

	app.Postgres.Close()

	app.Logger.Info("application shutdown completed")
	return nil
}

func (app *Application) Health(ctx context.Context) error {
	if err := app.Postgres.Health(ctx); err != nil {
		return fmt.Errorf("postgres health check failed: %w", err)
	}
	if err := app.Migrator.Health(ctx); err != nil {
		return fmt.Errorf("migrator health check failed: %w", err)
	}
	return nil
}
