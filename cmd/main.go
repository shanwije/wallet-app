package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/shanwije/wallet-app/internal/config"
)

func main() {
	cfg := config.LoadConfig()

	fmt.Printf("Starting server on port %s...\n", cfg.AppPort)
	log.Fatal(http.ListenAndServe(":"+cfg.AppPort, nil))
}
