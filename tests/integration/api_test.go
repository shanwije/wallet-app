package integration

import (
	"fmt"
	"net/http"
	"os"
	"testing"
)

// getTestURL returns the base URL for integration tests
func getTestURL() string {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8082" // Default port
	}
	return fmt.Sprintf("http://localhost:%s", port)
}

// TestSwaggerEndpoint tests that swagger documentation is accessible
func TestSwaggerEndpoint(t *testing.T) {
	// Test the swagger endpoint
	resp, err := http.Get(getTestURL() + "/swagger/index.html")
	if err != nil {
		t.Fatalf("Failed to make request to swagger endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

// TestRootEndpoint tests the root endpoint
func TestRootEndpoint(t *testing.T) {
	// Test the root endpoint
	resp, err := http.Get(getTestURL() + "/")
	if err != nil {
		t.Fatalf("Failed to make request to root endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}
