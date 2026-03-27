// Package logger provides a structured JSON logger backed by log/slog.
// Call Init() once at startup. In release mode the level is Info;
// in debug mode (SERVER_MODE != "release") the level is Debug.
package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

var defaultLogger *slog.Logger

// Init initialises the global logger. Must be called after config.Init().
func Init() {
	level := slog.LevelInfo
	if os.Getenv("SERVER_MODE") != "release" {
		level = slog.LevelDebug
	}
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	defaultLogger = slog.New(h)
	slog.SetDefault(defaultLogger)
}

func logger() *slog.Logger {
	if defaultLogger == nil {
		// fallback before Init() is called (e.g. in tests)
		return slog.Default()
	}
	return defaultLogger
}

func Info(msg string)                          { logger().InfoContext(context.Background(), msg) }
func Infof(format string, args ...any)         { logger().InfoContext(context.Background(), sprintf(format, args...)) }
func Warn(msg string)                          { logger().WarnContext(context.Background(), msg) }
func Warnf(format string, args ...any)         { logger().WarnContext(context.Background(), sprintf(format, args...)) }
func Error(msg string)                         { logger().ErrorContext(context.Background(), msg) }
func Errorf(format string, args ...any)        { logger().ErrorContext(context.Background(), sprintf(format, args...)) }
func Debug(msg string)                         { logger().DebugContext(context.Background(), msg) }
func Debugf(format string, args ...any)        { logger().DebugContext(context.Background(), sprintf(format, args...)) }

// WithFields returns a child logger with additional key-value pairs.
func WithFields(args ...any) *slog.Logger {
	return logger().With(args...)
}

func sprintf(format string, args ...any) string {
	if len(args) == 0 {
		return format
	}
	return fmt.Sprintf(format, args...)
}
