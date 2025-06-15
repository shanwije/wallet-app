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
		if tx.Type == "withdraw" {
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

// TestWalletTransfer tests the complete transfer flow between two wallets
func TestWalletTransfer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	baseURL := getTestURL()

	// Create sender user
	senderPayload := map[string]string{"name": "Transfer Sender"}
	senderJSON, _ := json.Marshal(senderPayload)

	resp, err := http.Post(baseURL+"/api/v1/users", "application/json", bytes.NewBuffer(senderJSON))
	if err != nil {
		t.Fatalf("Failed to create sender user: %v", err)
	}
	defer resp.Body.Close()

	var senderUser models.UserWithWallet
	err = json.NewDecoder(resp.Body).Decode(&senderUser)
	if err != nil {
		t.Fatalf("Failed to decode sender response: %v", err)
	}
	senderWalletID := senderUser.Wallet.ID.String()

	// Create receiver user
	receiverPayload := map[string]string{"name": "Transfer Receiver"}
	receiverJSON, _ := json.Marshal(receiverPayload)

	resp, err = http.Post(baseURL+"/api/v1/users", "application/json", bytes.NewBuffer(receiverJSON))
	if err != nil {
		t.Fatalf("Failed to create receiver user: %v", err)
	}
	defer resp.Body.Close()

	var receiverUser models.UserWithWallet
	err = json.NewDecoder(resp.Body).Decode(&receiverUser)
	if err != nil {
		t.Fatalf("Failed to decode receiver response: %v", err)
	}
	receiverWalletID := receiverUser.Wallet.ID.String()

	// Deposit money to sender wallet
	depositPayload := map[string]float64{"amount": 200.00}
	depositJSON, _ := json.Marshal(depositPayload)

	resp, err = http.Post(fmt.Sprintf("%s/api/v1/wallets/%s/deposit", baseURL, senderWalletID), "application/json", bytes.NewBuffer(depositJSON))
	if err != nil {
		t.Fatalf("Failed to make deposit: %v", err)
	}
	defer resp.Body.Close()

	// Transfer money from sender to receiver
	transferPayload := map[string]interface{}{
		"to_wallet_id": receiverWalletID,
		"amount":       50.75,
	}
	transferJSON, _ := json.Marshal(transferPayload)

	resp, err = http.Post(fmt.Sprintf("%s/api/v1/wallets/%s/transfer", baseURL, senderWalletID), "application/json", bytes.NewBuffer(transferJSON))
	if err != nil {
		t.Fatalf("Failed to make transfer: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Verify sender balance (should be 149.25)
	resp, err = http.Get(fmt.Sprintf("%s/api/v1/wallets/%s/balance", baseURL, senderWalletID))
	if err != nil {
		t.Fatalf("Failed to get sender balance: %v", err)
	}
	defer resp.Body.Close()

	var senderBalance map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&senderBalance)
	if err != nil {
		t.Fatalf("Failed to decode sender balance: %v", err)
	}

	expectedSenderBalance := 149.25
	if senderBalance["balance"].(float64) != expectedSenderBalance {
		t.Errorf("Expected sender balance %f, got %f", expectedSenderBalance, senderBalance["balance"].(float64))
	}

	// Verify receiver balance (should be 50.75)
	resp, err = http.Get(fmt.Sprintf("%s/api/v1/wallets/%s/balance", baseURL, receiverWalletID))
	if err != nil {
		t.Fatalf("Failed to get receiver balance: %v", err)
	}
	defer resp.Body.Close()

	var receiverBalance map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&receiverBalance)
	if err != nil {
		t.Fatalf("Failed to decode receiver balance: %v", err)
	}

	expectedReceiverBalance := 50.75
	if receiverBalance["balance"].(float64) != expectedReceiverBalance {
		t.Errorf("Expected receiver balance %f, got %f", expectedReceiverBalance, receiverBalance["balance"].(float64))
	}
}

// TestErrorScenarios tests various error conditions
func TestErrorScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	baseURL := getTestURL()

	// Create a user with minimal balance
	userPayload := map[string]string{"name": "Error Test User"}
	userJSON, _ := json.Marshal(userPayload)

	resp, err := http.Post(baseURL+"/api/v1/users", "application/json", bytes.NewBuffer(userJSON))
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	defer resp.Body.Close()

	var user models.UserWithWallet
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		t.Fatalf("Failed to decode user response: %v", err)
	}
	walletID := user.Wallet.ID.String()

	// Test insufficient funds withdrawal
	withdrawPayload := map[string]float64{"amount": 100.00} // More than available balance (0)
	withdrawJSON, _ := json.Marshal(withdrawPayload)

	resp, err = http.Post(fmt.Sprintf("%s/api/v1/wallets/%s/withdraw", baseURL, walletID), "application/json", bytes.NewBuffer(withdrawJSON))
	if err != nil {
		t.Fatalf("Failed to make withdrawal request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d for insufficient funds, got %d", http.StatusBadRequest, resp.StatusCode)
	}

	// Test negative amount deposit
	negativeDepositPayload := map[string]float64{"amount": -50.00}
	negativeDepositJSON, _ := json.Marshal(negativeDepositPayload)

	resp, err = http.Post(fmt.Sprintf("%s/api/v1/wallets/%s/deposit", baseURL, walletID), "application/json", bytes.NewBuffer(negativeDepositJSON))
	if err != nil {
		t.Fatalf("Failed to make negative deposit request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d for negative amount, got %d", http.StatusBadRequest, resp.StatusCode)
	}

	// Test invalid wallet ID
	invalidWalletID := "invalid-uuid"
	resp, err = http.Get(fmt.Sprintf("%s/api/v1/wallets/%s/balance", baseURL, invalidWalletID))
	if err != nil {
		t.Fatalf("Failed to make invalid wallet request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid wallet ID, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

// TestIdempotencyWithWalletOperations tests idempotency with wallet operations
func TestIdempotencyWithWalletOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	baseURL := getTestURL()

	// Create a user
	userPayload := map[string]string{"name": "Idempotency Wallet Test User"}
	userJSON, _ := json.Marshal(userPayload)

	resp, err := http.Post(baseURL+"/api/v1/users", "application/json", bytes.NewBuffer(userJSON))
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	defer resp.Body.Close()

	var user models.UserWithWallet
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		t.Fatalf("Failed to decode user response: %v", err)
	}
	walletID := user.Wallet.ID.String()

	// Test idempotent deposit
	depositPayload := map[string]float64{"amount": 100.00}
	depositJSON, _ := json.Marshal(depositPayload)

	client := &http.Client{}

	// First deposit with idempotency key
	req1, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/wallets/%s/deposit", baseURL, walletID), bytes.NewBuffer(depositJSON))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Idempotency-Key", "deposit-test-key-456")

	resp1, err := client.Do(req1)
	if err != nil {
		t.Fatalf("Failed to make first deposit: %v", err)
	}
	defer resp1.Body.Close()

	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, resp1.StatusCode)
	}

	// Second deposit with same idempotency key
	req2, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/wallets/%s/deposit", baseURL, walletID), bytes.NewBuffer(depositJSON))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Idempotency-Key", "deposit-test-key-456")

	resp2, err := client.Do(req2)
	if err != nil {
		t.Fatalf("Failed to make second deposit: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, resp2.StatusCode)
	}

	// Check that balance is only 100 (not 200, proving idempotency worked)
	resp, err = http.Get(fmt.Sprintf("%s/api/v1/wallets/%s/balance", baseURL, walletID))
	if err != nil {
		t.Fatalf("Failed to get balance: %v", err)
	}
	defer resp.Body.Close()

	var balance map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&balance)
	if err != nil {
		t.Fatalf("Failed to decode balance: %v", err)
	}

	expectedBalance := 100.0
	if balance["balance"].(float64) != expectedBalance {
		t.Errorf("Idempotency failed: expected balance %f, got %f", expectedBalance, balance["balance"].(float64))
	}

	// Check transaction history shows only one deposit
	time.Sleep(100 * time.Millisecond)

	resp, err = http.Get(fmt.Sprintf("%s/api/v1/wallets/%s/transactions", baseURL, walletID))
	if err != nil {
		t.Fatalf("Failed to get transactions: %v", err)
	}
	defer resp.Body.Close()

	var transactions []models.Transaction
	err = json.NewDecoder(resp.Body).Decode(&transactions)
	if err != nil {
		t.Fatalf("Failed to decode transactions: %v", err)
	}

	if len(transactions) != 1 {
		t.Errorf("Idempotency failed: expected 1 transaction, got %d", len(transactions))
	}

	if transactions[0].Type != "deposit" {
		t.Errorf("Expected deposit transaction, got %s", transactions[0].Type)
	}
}
