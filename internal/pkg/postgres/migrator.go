package postgres

import (
	"context"
	"fmt"
	"github.com/ZertGraf/avito-test/internal/pkg/logger"
	"github.com/ZertGraf/avito-test/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/tern/v2/migrate"
	"time"
)

type MigrationConfig struct {
	Timeout   time.Duration `json:"timeout"`
	TableName string        `json:"table_name"`
	Enabled   bool          `json:"enabled"`
}

type Migrator struct {
	pool   *pgxpool.Pool
	logger *logger.Logger
	config *MigrationConfig
}

func NewMigrator(pool *pgxpool.Pool, config *MigrationConfig, logger *logger.Logger) *Migrator {
	return &Migrator{
		pool:   pool,
		logger: logger.Component("postgres/migrator"),
		config: config,
	}
}

func (m *Migrator) RunMigrations(ctx context.Context) error {
	if !m.config.Enabled {
		m.logger.Info("migrations disabled, skipping")
		return nil
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, m.config.Timeout)
	defer cancel()

	conn, err := m.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	migrator, err := migrate.NewMigrator(ctx, conn.Conn(), m.config.TableName)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	if err = migrator.LoadMigrations(migrations.MigrationFiles); err != nil {
		return fmt.Errorf("load migrations: %w", err)
	}

	currentVersion, err := migrator.GetCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("get current version: %w", err)
	}

	// load available migrations
	maxVersion := int32(0)
	for _, migration := range migrator.Migrations {
		if migration.Sequence > maxVersion {
			maxVersion = migration.Sequence
		}
	}

	pendingCount := maxVersion - currentVersion
	if pendingCount <= 0 {
		m.logger.Info("database schema up to date",
			"current_version", currentVersion,
			"latest_version", maxVersion)
		return nil
	}

	m.logger.Info("applying database migrations",
		"current_version", currentVersion,
		"target_version", maxVersion,
		"pending_migrations", pendingCount)

	// apply them
	if err = migrator.Migrate(ctx); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	finalVersion, err := migrator.GetCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("get final version: %w", err)
	}

	duration := time.Since(start)
	m.logger.Info("migrations completed successfully",
		"from_version", currentVersion,
		"to_version", finalVersion,
		"applied_count", finalVersion-currentVersion,
		"duration", duration)

	return nil
}

func (m *Migrator) GetCurrentVersion(ctx context.Context) (int32, error) {
	conn, err := m.pool.Acquire(ctx)
	if err != nil {
		return 0, fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	migrator, err := migrate.NewMigrator(ctx, conn.Conn(), m.config.TableName)
	if err != nil {
		return 0, fmt.Errorf("create migrator: %w", err)
	}

	return migrator.GetCurrentVersion(ctx)
}

func (m *Migrator) Health(ctx context.Context) error {
	_, err := m.GetCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("migration health check failed: %w", err)
	}
	return nil
}
