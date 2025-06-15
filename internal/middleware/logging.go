package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/shanwije/wallet-app/pkg/logger"
)

// RequestIDMiddleware adds request ID to context and response headers
func RequestIDMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get request ID from header or generate new one
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Add request ID to response header
			w.Header().Set("X-Request-ID", requestID)

			// Add request ID to context with logger
			ctx := logger.WithRequestID(r.Context(), requestID)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// LoggingMiddleware provides structured request logging
func LoggingMiddleware() func(http.Handler) http.Handler {
	return middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger: &ChiZapLogger{},
	})
}

// ChiZapLogger adapts zap logger for chi middleware
type ChiZapLogger struct{}

func (l *ChiZapLogger) Print(v ...interface{}) {
	logger.Log.Sugar().Info(v...)
}
