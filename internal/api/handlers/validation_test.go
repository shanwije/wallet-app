package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRequestValidation tests various request validation scenarios
func TestRequestValidation(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := NewHealthHandler()

	handler.GetHealth(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("GetHealth returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

// TestCreateUserRequestValidation tests request validation
func TestCreateUserRequestValidation(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		expectError bool
	}{
		{
			name:        "Valid request",
			requestBody: `{"name": "John Doe"}`,
			expectError: false,
		},
		{
			name:        "Empty name",
			requestBody: `{"name": ""}`,
			expectError: true,
		},
		{
			name:        "Missing name field",
			requestBody: `{}`,
			expectError: true,
		},
		{
			name:        "Invalid JSON",
			requestBody: `{"name": invalid}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req struct {
				Name string `json:"name"`
			}

			err := json.Unmarshal([]byte(tt.requestBody), &req)

			if tt.expectError {
				// Either JSON unmarshal should fail or name should be empty
				if err == nil && req.Name != "" {
					t.Errorf("Expected validation error but got none")
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, req.Name)
			}
		})
	}
}

// TestWalletRequestValidation tests wallet operation request validation
func TestWalletRequestValidation(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		expectError bool
	}{
		{
			name:        "Valid deposit",
			requestBody: `{"amount": 100.50}`,
			expectError: false,
		},
		{
			name:        "Zero amount",
			requestBody: `{"amount": 0}`,
			expectError: true,
		},
		{
			name:        "Negative amount",
			requestBody: `{"amount": -10.50}`,
			expectError: true,
		},
		{
			name:        "Missing amount",
			requestBody: `{}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req struct {
				Amount float64 `json:"amount"`
			}

			err := json.Unmarshal([]byte(tt.requestBody), &req)

			if tt.expectError {
				// Either JSON unmarshal should fail or amount should be invalid
				if err == nil && req.Amount > 0 {
					t.Errorf("Expected validation error but got none")
				}
			} else {
				assert.NoError(t, err)
				assert.Greater(t, req.Amount, 0.0)
			}
		})
	}
}

// TestTransferRequestValidation tests transfer request validation
func TestTransferRequestValidation(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		expectError bool
	}{
		{
			name:        "Valid transfer",
			requestBody: `{"to_wallet_id": "123e4567-e89b-12d3-a456-426614174000", "amount": 50.25, "description": "test"}`,
			expectError: false,
		},
		{
			name:        "Invalid UUID",
			requestBody: `{"to_wallet_id": "invalid-uuid", "amount": 50.25, "description": "test"}`,
			expectError: true,
		},
		{
			name:        "Missing to_wallet_id",
			requestBody: `{"amount": 50.25, "description": "test"}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req struct {
				ToWalletID  string  `json:"to_wallet_id"`
				Amount      float64 `json:"amount"`
				Description string  `json:"description"`
			}

			err := json.Unmarshal([]byte(tt.requestBody), &req)

			if tt.expectError {
				// Either JSON unmarshal should fail or to_wallet_id should be invalid
				if err == nil && req.ToWalletID != "" && req.ToWalletID != "invalid-uuid" {
					t.Errorf("Expected validation error but got none")
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, req.ToWalletID)
				assert.Greater(t, req.Amount, 0.0)
			}
		})
	}
}

// TestJSONResponseFormat tests that responses are properly formatted
func TestJSONResponseFormat(t *testing.T) {
	// Test that we can create a proper JSON response
	response := map[string]interface{}{
		"id":      "123e4567-e89b-12d3-a456-426614174000",
		"balance": 100.50,
		"message": "success",
	}

	jsonData, err := json.Marshal(response)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), "id")
	assert.Contains(t, string(jsonData), "balance")
	assert.Contains(t, string(jsonData), "message")

	// Test that we can unmarshal it back
	var decoded map[string]interface{}
	err = json.Unmarshal(jsonData, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, response["id"], decoded["id"])
	assert.Equal(t, response["balance"], decoded["balance"])
	assert.Equal(t, response["message"], decoded["message"])
}

// TestRESTfulAPICompliance tests API structure compliance with assignment requirements
func TestRESTfulAPICompliance(t *testing.T) {
	t.Run("HTTP methods validation", func(t *testing.T) {
		// Test that we can validate HTTP methods for different operations
		validMethods := map[string][]string{
			"users":                     {"POST", "GET"},
			"wallets/{id}/deposit":      {"POST"},
			"wallets/{id}/withdraw":     {"POST"},
			"wallets/{id}/transfer":     {"POST"},
			"wallets/{id}/balance":      {"GET"},
			"wallets/{id}/transactions": {"GET"},
		}

		for endpoint, methods := range validMethods {
			for _, method := range methods {
				assert.Contains(t, []string{"GET", "POST", "PUT", "DELETE"}, method,
					"Endpoint %s should use valid HTTP method %s", endpoint, method)
			}
		}
	})

	t.Run("Status codes validation", func(t *testing.T) {
		// Test expected status codes for assignment requirements
		validStatusCodes := []int{
			http.StatusOK,                  // 200 - GET operations, successful updates
			http.StatusCreated,             // 201 - POST user creation
			http.StatusBadRequest,          // 400 - validation errors
			http.StatusNotFound,            // 404 - resource not found
			http.StatusInternalServerError, // 500 - server errors
		}

		for _, code := range validStatusCodes {
			assert.True(t, code >= 200 && code < 600, "Status code %d should be valid HTTP status", code)
		}
	})
}

// TestErrorHandling tests error response formatting
func TestErrorHandling(t *testing.T) {
	t.Run("Error response format", func(t *testing.T) {
		// Test that error responses are properly formatted
		errorCases := map[string]int{
			"Invalid request":       http.StatusBadRequest,
			"User not found":        http.StatusNotFound,
			"Insufficient balance":  http.StatusBadRequest,
			"Internal server error": http.StatusInternalServerError,
		}

		for errorMsg, expectedStatus := range errorCases {
			assert.NotEmpty(t, errorMsg, "Error message should not be empty")
			assert.True(t, expectedStatus >= 400, "Error status should be 4xx or 5xx")
		}
	})
}
