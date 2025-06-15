package integration

import (
	"net/http"
	"testing"
)

// TestSwaggerEndpoint tests that swagger documentation is accessible
func TestSwaggerEndpoint(t *testing.T) {
	// Test the swagger endpoint
	resp, err := http.Get("http://localhost:8082/swagger/index.html")
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
	resp, err := http.Get("http://localhost:8082/")
	if err != nil {
		t.Fatalf("Failed to make request to root endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}
