package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/shanwije/wallet-app/internal/api/handlers"
	"github.com/shanwije/wallet-app/internal/config"
)

// Router sets up the HTTP router with all routes
func NewRouter(cfg *config.Config, db *sqlx.DB) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Compress(5))

	// CORS middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	})

	// Handlers
	healthHandler := handlers.NewHealthHandler()

	// Routes - using configurable API version
	apiRoute := fmt.Sprintf("/api/%s", cfg.APIVersion)
	r.Route(apiRoute, func(r chi.Router) {
		r.Get("/health", healthHandler.GetHealth)
	})

	// Health check at root level as well
	r.Get("/health", healthHandler.GetHealth)

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// Root endpoint
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Wallet API is running","swagger":"/swagger/index.html"}`))
	})

	log.Println("Router configured with Swagger documentation at /swagger/index.html")
	return r
}
