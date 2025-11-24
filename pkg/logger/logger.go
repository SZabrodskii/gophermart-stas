package logger

import (
	"log/slog"
	"os"

	"go.uber.org/fx"
)

type Logger struct {
	*slog.Logger
}

func New() *Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return &Logger{Logger: logger}
}

var Module = fx.Options(
	fx.Provide(New),
)
