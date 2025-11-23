package postgres

import (
	"context"
	"fmt"
	"github.com/ZertGraf/avito-test/internal/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Connection struct {
	pool   *pgxpool.Pool
	logger *logger.Logger
	config *Config
}

func New(logger *logger.Logger, config *Config) (*Connection, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid postgres config: %w", err)
	}
	return &Connection{
		config: config,
		logger: logger.Component("database/postgres"),
	}, nil
}

func (c *Connection) Connect(ctx context.Context) error {
	cfg, err := pgxpool.ParseConfig(c.config.DSN())
	if err != nil {
		return fmt.Errorf("failed to parse postgres dsn: %w", err)
	}
	cfg.MaxConns = c.config.MaxConns
	cfg.MinConns = c.config.MinConns
	cfg.MaxConnLifetime = c.config.MaxConnLifetime
	cfg.MaxConnIdleTime = c.config.MaxConnIdleTime
	cfg.HealthCheckPeriod = c.config.HealthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create postgres connection: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to ping postgres: %w", err)
	}

	c.pool = pool

	c.logger.Info("postres connection established",
		"host", c.config.Host,
		"database", c.config.Database,
		"max_conns", c.config.MaxConns)

	return nil
}

func (c *Connection) Pool() *pgxpool.Pool {
	if c.pool == nil {
		panic("postgres connection not established, call Connect() first")
	}
	return c.pool
}

func (c *Connection) Close() {
	if c.pool != nil {
		c.pool.Close()
	}
}

func (c *Connection) Health(ctx context.Context) error {
	if c.pool == nil {
		return fmt.Errorf("postgres pool not initialized")
	}
	return c.pool.Ping(ctx)
}
