package logger

import (
	"log/slog"
	"os"
)

type Logger struct {
	sl *slog.Logger
}

func New(env string) *Logger {
	level := slog.LevelInfo
	if env == "development" || env == "dev" {
		level = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return &Logger{sl: slog.New(handler)}
}

func (l *Logger) Info(msg string, args ...any) {
	l.sl.Info(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.sl.Error(msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.sl.Debug(msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.sl.Warn(msg, args...)
}
