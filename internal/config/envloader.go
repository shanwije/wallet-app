package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string `validate:"required" env:"DB_HOST"`
	DBPort     string `validate:"required,numeric" env:"DB_PORT"`
	DBUser     string `validate:"required" env:"DB_USER"`
	DBPassword string `validate:"required" env:"DB_PASSWORD"`
	DBName     string `validate:"required" env:"DB_NAME"`
	DBSSLMode  string `validate:"required,oneof=disable require verify-ca verify-full" env:"DB_SSL_MODE"`

	AppPort     string `validate:"required,numeric" env:"APP_PORT"`
	APIVersion  string `validate:"required" env:"API_VERSION"`
	Environment string `validate:"required,oneof=development staging production" env:"ENVIRONMENT"`
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load() // Only loads from .env in dev

	config := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "wallet"),
		DBPassword: getEnv("DB_PASSWORD", "walletpass"),
		DBName:     getEnv("DB_NAME", "wallet_db"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),

		AppPort:     getEnv("APP_PORT", "8082"),
		APIVersion:  getEnv("API_VERSION", "v1"),
		Environment: getEnv("ENVIRONMENT", "development"),
	}

	// Validate configuration
	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

func getEnv(key string, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
