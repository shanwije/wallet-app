package logger

import (
	"context"
	"os"

	"go.uber.org/zap"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	LoggerKey ContextKey = "logger"
)

var (
	// Global logger instance
	Log *zap.Logger
)

// Initialize sets up the global logger
func Initialize(env string) error {
	var config zap.Config

	if env == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	logger, err := config.Build()
	if err != nil {
		return err
	}

	Log = logger
	return nil
}

// FromContext extracts request-scoped logger from context
func FromContext(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(LoggerKey).(*zap.Logger); ok {
		return logger
	}
	return Log
}

// WithRequestID adds request ID to logger
func WithRequestID(ctx context.Context, requestID string) context.Context {
	logger := Log.With(zap.String("request_id", requestID))
	return context.WithValue(ctx, LoggerKey, logger)
}

// Close gracefully shuts down the logger
func Close() error {
	if Log != nil {
		return Log.Sync()
	}
	return nil
}

// GetEnvironment returns the current environment
func GetEnvironment() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}
	return env
}
