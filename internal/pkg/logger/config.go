package logger

import (
	. "github.com/go-ozzo/ozzo-validation"
	"log/slog"
)

type Config struct {
	Level     string
	Format    string
	AddSource bool
}

func (c *Config) Validate() error {
	return ValidateStruct(c,
		Field(&c.Level, Required, In("debug", "info", "warn", "error", "fatal")),
		Field(&c.Format, Required, In("json", "text")),
	)
}

func (c *Config) GetSlogLevel() slog.Level {
	switch c.Level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
