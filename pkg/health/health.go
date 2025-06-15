package health

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Status represents the health status
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

// HealthCheck represents a single health check
type HealthCheck struct {
	Name        string            `json:"name"`
	Status      Status            `json:"status"`
	Message     string            `json:"message,omitempty"`
	Duration    time.Duration     `json:"duration"`
	LastChecked time.Time         `json:"last_checked"`
	Details     map[string]string `json:"details,omitempty"`
}

// HealthReport represents the overall health status
type HealthReport struct {
	Status      Status                 `json:"status"`
	Version     string                 `json:"version"`
	Timestamp   time.Time              `json:"timestamp"`
	Checks      map[string]HealthCheck `json:"checks"`
	Environment string                 `json:"environment"`
}

// Checker is the interface that health checks must implement
type Checker interface {
	Check(ctx context.Context) HealthCheck
}

// DatabaseChecker checks database connectivity
type DatabaseChecker struct {
	DB   *sql.DB
	Name string
}

func (d *DatabaseChecker) Check(ctx context.Context) HealthCheck {
	start := time.Now()

	check := HealthCheck{
		Name:        d.Name,
		LastChecked: start,
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := d.DB.PingContext(timeoutCtx); err != nil {
		check.Status = StatusUnhealthy
		check.Message = "Database connection failed"
		check.Details = map[string]string{"error": err.Error()}
	} else {
		check.Status = StatusHealthy
		check.Message = "Database connection successful"
	}

	check.Duration = time.Since(start)
	return check
}

// Handler provides HTTP health check endpoints
type Handler struct {
	checkers    map[string]Checker
	version     string
	environment string
	logger      *zap.Logger
}

// NewHandler creates a new health check handler
func NewHandler(version, environment string, logger *zap.Logger) *Handler {
	return &Handler{
		checkers:    make(map[string]Checker),
		version:     version,
		environment: environment,
		logger:      logger,
	}
}

// AddChecker adds a health checker
func (h *Handler) AddChecker(name string, checker Checker) {
	h.checkers[name] = checker
}

// HealthHandler returns the overall health status
func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	report := HealthReport{
		Version:     h.version,
		Timestamp:   time.Now(),
		Environment: h.environment,
		Checks:      make(map[string]HealthCheck),
	}

	overallStatus := StatusHealthy

	// Run all health checks
	for name, checker := range h.checkers {
		check := checker.Check(ctx)
		report.Checks[name] = check

		// Determine overall status
		if check.Status == StatusUnhealthy {
			overallStatus = StatusUnhealthy
		} else if check.Status == StatusDegraded && overallStatus == StatusHealthy {
			overallStatus = StatusDegraded
		}
	}

	report.Status = overallStatus

	// Set appropriate HTTP status code
	var httpStatus int
	switch overallStatus {
	case StatusHealthy:
		httpStatus = http.StatusOK
	case StatusDegraded:
		httpStatus = http.StatusOK // Still return 200 for degraded
	case StatusUnhealthy:
		httpStatus = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	if err := json.NewEncoder(w).Encode(report); err != nil {
		h.logger.Error("Failed to encode health report", zap.Error(err))
	}
}

// ReadinessHandler returns readiness status (simpler check for k8s)
func (h *Handler) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Only check critical services for readiness
	for name, checker := range h.checkers {
		if name == "database" { // Only database is critical for readiness
			check := checker.Check(ctx)
			if check.Status == StatusUnhealthy {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("Not Ready"))
				return
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ready"))
}
