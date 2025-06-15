package main

import (
	"log"
	"net/http"

	"github.com/shanwije/wallet-app/internal/api"
	"github.com/shanwije/wallet-app/internal/config"
	"github.com/shanwije/wallet-app/pkg/db"
	_ "github.com/shanwije/wallet-app/docs"
)

func main() {
	// Load env config
	cfg := config.LoadConfig()

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
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer dbConn.Close()
	log.Println("Database connection established")
	
	// Setup router and inject DB
	router := api.NewRouter(cfg, dbConn)

	// Start server
	log.Printf("Server started on port %s", cfg.AppPort)
	log.Fatal(http.ListenAndServe(":"+cfg.AppPort, router))
}
