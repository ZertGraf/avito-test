package logger

import (
	"fmt"
	"github.com/lmittmann/tint"
	"log/slog"
	"os"
)

type Logger struct {
	*slog.Logger
}

func New(cfg *Config) (*Logger, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	handler := createHandler(cfg)
	logger := slog.New(handler)
	return &Logger{logger}, nil
}

func createHandler(cfg *Config) slog.Handler {
	opts := &slog.HandlerOptions{
		Level:     cfg.GetSlogLevel(),
		AddSource: cfg.AddSource,
	}

	switch cfg.Format {
	case "text":
		return tint.NewHandler(os.Stdout, &tint.Options{
			Level:      opts.Level,
			AddSource:  opts.AddSource,
			TimeFormat: "15:04:05",
		})
	case "json":
		fallthrough
	default:
		return slog.NewJSONHandler(os.Stdout, opts)
	}
}

func (l *Logger) Component(name string) *Logger {
	return &Logger{l.Logger.With("component", name)}
}
