package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/shanwije/wallet-app/docs"
	"github.com/shanwije/wallet-app/internal/api"
	"github.com/shanwije/wallet-app/internal/config"
	"github.com/shanwije/wallet-app/pkg/db"
	"github.com/shanwije/wallet-app/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger first
	if err := logger.Initialize(logger.GetEnvironment()); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Close()

	log := logger.Log

	// Load and validate config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config", zap.Error(err))
	}

	log.Info("Starting wallet service",
		zap.String("version", cfg.APIVersion),
		zap.String("environment", cfg.Environment),
		zap.String("port", cfg.AppPort),
	)

	// Setup DB connection
	pgCfg := db.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		Name:     cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	}

	dbConn, err := db.New(pgCfg)
	if err != nil {
		log.Fatal("Failed to connect to DB", zap.Error(err))
	}
	defer dbConn.Close()

	log.Info("Database connection established")

	// Setup router and inject dependencies
	router := api.NewRouter(cfg, dbConn, log)

	// Setup HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Info("Server starting", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Server shutting down...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server gracefully
	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", zap.Error(err))
		return
	}

	log.Info("Server exited")
}
