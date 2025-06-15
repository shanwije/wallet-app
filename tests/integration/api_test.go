package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/shanwije/wallet-app/internal/models"
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

// TestWalletTransactionHistory tests the complete flow of wallet operations and transaction history
func TestWalletTransactionHistory(t *testing.T) {
	// Skip if not integration test
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	baseURL := getTestURL()

	// 1. Create a user
	userPayload := map[string]string{
		"name": "Transaction Test User",
	}
	userJSON, _ := json.Marshal(userPayload)

	resp, err := http.Post(baseURL+"/api/v1/users", "application/json", bytes.NewBuffer(userJSON))
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	var userWithWallet models.UserWithWallet
	err = json.NewDecoder(resp.Body).Decode(&userWithWallet)
	if err != nil {
		t.Fatalf("Failed to decode user response: %v", err)
	}

	walletID := userWithWallet.Wallet.ID.String()

	// 2. Make a deposit
	depositPayload := map[string]float64{
		"amount": 100.50,
	}
	depositJSON, _ := json.Marshal(depositPayload)

	resp, err = http.Post(fmt.Sprintf("%s/api/v1/wallets/%s/deposit", baseURL, walletID), "application/json", bytes.NewBuffer(depositJSON))
	if err != nil {
		t.Fatalf("Failed to make deposit: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// 3. Make a withdrawal
	withdrawPayload := map[string]float64{
		"amount": 25.25,
	}
	withdrawJSON, _ := json.Marshal(withdrawPayload)

	resp, err = http.Post(fmt.Sprintf("%s/api/v1/wallets/%s/withdraw", baseURL, walletID), "application/json", bytes.NewBuffer(withdrawJSON))
	if err != nil {
		t.Fatalf("Failed to make withdrawal: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// 4. Wait a bit for transactions to be recorded
	time.Sleep(100 * time.Millisecond)

	// 5. Get transaction history
	resp, err = http.Get(fmt.Sprintf("%s/api/v1/wallets/%s/transactions", baseURL, walletID))
	if err != nil {
		t.Fatalf("Failed to get transaction history: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var transactions []models.Transaction
	err = json.NewDecoder(resp.Body).Decode(&transactions)
	if err != nil {
		t.Fatalf("Failed to decode transaction history: %v", err)
	}

	// 6. Verify transaction history
	if len(transactions) != 2 {
		t.Fatalf("Expected 2 transactions, got %d", len(transactions))
	}

	// Transactions should be in descending order (newest first)
	foundDeposit := false
	foundWithdrawal := false

	for _, tx := range transactions {
		if tx.Type == "deposit" {
			foundDeposit = true
		}
		if tx.Type == "withdrawal" {
			foundWithdrawal = true
		}

		// Verify transaction belongs to the correct wallet
		if tx.WalletID != userWithWallet.Wallet.ID {
			t.Errorf("Transaction wallet ID mismatch: expected %s, got %s", userWithWallet.Wallet.ID, tx.WalletID)
		}
	}

	if !foundDeposit {
		t.Error("Deposit transaction not found in history")
	}
	if !foundWithdrawal {
		t.Error("Withdrawal transaction not found in history")
	}
}

// TestIdempotencyMiddleware tests the idempotency functionality
func TestIdempotencyMiddleware(t *testing.T) {
	// Skip if not integration test
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	baseURL := getTestURL()

	// Create a user with idempotency key
	userPayload := map[string]string{
		"name": "Idempotency Test User",
	}
	userJSON, _ := json.Marshal(userPayload)

	// First request with idempotency key
	req, _ := http.NewRequest("POST", baseURL+"/api/v1/users", bytes.NewBuffer(userJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "test-key-123")

	client := &http.Client{}
	resp1, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make first request: %v", err)
	}
	defer resp1.Body.Close()

	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status %d, got %d", http.StatusCreated, resp1.StatusCode)
	}

	var user1 models.UserWithWallet
	err = json.NewDecoder(resp1.Body).Decode(&user1)
	if err != nil {
		t.Fatalf("Failed to decode first response: %v", err)
	}

	// Second request with same idempotency key
	req2, _ := http.NewRequest("POST", baseURL+"/api/v1/users", bytes.NewBuffer(userJSON))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Idempotency-Key", "test-key-123")

	resp2, err := client.Do(req2)
	if err != nil {
		t.Fatalf("Failed to make second request: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status %d, got %d", http.StatusCreated, resp2.StatusCode)
	}

	var user2 models.UserWithWallet
	err = json.NewDecoder(resp2.Body).Decode(&user2)
	if err != nil {
		t.Fatalf("Failed to decode second response: %v", err)
	}

	// Both responses should be identical (same user ID)
	if user1.ID != user2.ID {
		t.Errorf("Idempotency failed: user IDs differ - %s vs %s", user1.ID, user2.ID)
	}
}
