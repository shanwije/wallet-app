package main

import (
	"log"
	"net/http"

	"github.com/shanwije/wallet-app/internal/api"
	"github.com/shanwije/wallet-app/internal/config"
	_ "github.com/shanwije/wallet-app/docs"
)

// @title Wallet API
// @version 1.0
// @description This is a simple wallet API that allows users to deposit, withdraw, and transfer money between their wallets.
func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Setup router
	router := api.NewRouter(cfg)

	// Start server
	log.Printf("Starting server on port %s...", cfg.AppPort)
	log.Fatal(http.ListenAndServe(":"+cfg.AppPort, router))
}
