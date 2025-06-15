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
	"github.com/shanwije/wallet-app/internal/repository/postgres"
	"github.com/shanwije/wallet-app/internal/service"
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

	// Create repositories
	userRepo := postgres.NewUserRepository(db)
	walletRepo := postgres.NewWalletRepository(db)

	// Create services
	userService := &service.UserService{UserRepo: userRepo, WalletRepo: walletRepo}
	walletService := &service.WalletService{WalletRepo: walletRepo}

	// Create handlers
	userHandler := &handlers.UserHandler{UserService: userService}
	walletHandler := &handlers.WalletHandler{WalletService: walletService}
	healthHandler := handlers.NewHealthHandler()

	// Routes - using configurable API version
	apiRoute := fmt.Sprintf("/api/%s", cfg.APIVersion)
	r.Route(apiRoute, func(r chi.Router) {
		r.Get("/health", healthHandler.GetHealth)
		r.Post("/users", userHandler.CreateUser)

		// Wallet operations
		r.Route("/wallets/{id}", func(r chi.Router) {
			r.Post("/deposit", walletHandler.Deposit)
			r.Post("/withdraw", walletHandler.Withdraw)
			r.Get("/balance", walletHandler.GetBalance)
		})
	})

	// Health check at root level for simple monitoring
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
