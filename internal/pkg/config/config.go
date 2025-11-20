package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	// application settings
	Environment string `env:"ENVIRONMENT" env-default:"development"`
	ServiceName string `env:"SERVICE_NAME" env-default:"pr-reviewer-service"`

	// logging configuration
	LogLevel     string `env:"LOG_LEVEL" env-default:"info"`
	LogFormat    string `env:"LOG_FORMAT" env-default:"text"`
	LogAddSource bool   `env:"LOG_ADD_SOURCE" env-default:"false"`

	// database connection settings
	DatabaseHost     string `env:"DATABASE_HOST" env-default:"localhost"`
	DatabasePort     int    `env:"DATABASE_PORT" env-default:"5432"`
	DatabaseUser     string `env:"DATABASE_USER" env-default:"postgres"`
	DatabasePassword string `env:"DATABASE_PASSWORD" env-required:"true"`
	DatabaseName     string `env:"DATABASE_NAME" env-default:"postgres"`
	DatabaseSchema   string `env:"DATABASE_SCHEMA" env-default:"public"`
	DatabaseSSLMode  string `env:"DATABASE_SSL_MODE" env-default:"require"`

	// database connection pool settings
	DatabaseMaxConns          int32         `env:"DATABASE_MAX_CONNS" env-default:"25"`
	DatabaseMinConns          int32         `env:"DATABASE_MIN_CONNS" env-default:"5"`
	DatabaseMaxConnLifetime   time.Duration `env:"DATABASE_MAX_CONN_LIFETIME" env-default:"1h"`
	DatabaseMaxConnIdleTime   time.Duration `env:"DATABASE_MAX_CONN_IDLE_TIME" env-default:"30m"`
	DatabaseHealthCheckPeriod time.Duration `env:"DATABASE_HEALTH_CHECK_PERIOD" env-default:"1m"`
	DatabaseConnectTimeout    time.Duration `env:"DATABASE_CONNECT_TIMEOUT" env-default:"30s"`
	DatabaseAcquireTimeout    time.Duration `env:"DATABASE_ACQUIRE_TIMEOUT" env-default:"10s"`

	// database migrations settings
	DatabaseMigrationEnabled bool          `env:"DATABASE_MIGRATION_ENABLED" env-default:"true"`
	DatabaseMigrationTimeout time.Duration `env:"DATABASE_MIGRATION_TIMEOUT" env-default:"5m"`
	DatabaseMigrationTable   string        `env:"DATABASE_MIGRATION_TABLE" env-default:"schema_version"`

	// http server configuration
	ServerHost         string        `env:"SERVER_HOST" env-default:"0.0.0.0"`
	ServerPort         int           `env:"SERVER_PORT" env-default:"8081"`
	ServerReadTimeout  time.Duration `env:"SERVER_READ_TIMEOUT" env-default:"30s"`
	ServerWriteTimeout time.Duration `env:"SERVER_WRITE_TIMEOUT" env-default:"30s"`
	ServerIdleTimeout  time.Duration `env:"SERVER_IDLE_TIMEOUT" env-default:"60s"`
}

func New() (*Config, error) {
	var cfg Config

	// read from .env file if exists (optional)
	if err := cleanenv.ReadConfig(".env", &cfg); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read dotenv file: %w", err)
	}

	// read from environment variables (required)
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read environment variables: %w", err)
	}

	return &cfg, nil
}
