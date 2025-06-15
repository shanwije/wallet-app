// Package service provides business logic for wallet operations
package service

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/shanwije/wallet-app/internal/models"
)

// Test fixtures and helper functions
const (
	testWalletBalance  = 100.0
	testDepositAmount  = 50.0
	testWithdrawAmount = 30.0
)

// setupWalletService creates a test wallet service with mocked dependencies
func setupWalletService() (*WalletService, *MockWalletRepositoryTest, *MockTransactionRepositoryTest) {
	walletRepo := new(MockWalletRepositoryTest)
	transactionRepo := new(MockTransactionRepositoryTest)
	service := &WalletService{
		WalletRepo:      walletRepo,
		TransactionRepo: transactionRepo,
	}
	return service, walletRepo, transactionRepo
}

// createTestWallet creates a wallet for testing
func createTestWallet(id uuid.UUID, balance float64) *models.Wallet {
	return &models.Wallet{
		ID:      id,
		Balance: decimal.NewFromFloat(balance),
	}
}

// MockWalletRepository for testing
type MockWalletRepositoryTest struct {
	mock.Mock
}

func (m *MockWalletRepositoryTest) CreateWallet(ctx context.Context, userID uuid.UUID) (*models.Wallet, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletRepositoryTest) GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*models.Wallet, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletRepositoryTest) GetWalletByID(ctx context.Context, id uuid.UUID) (*models.Wallet, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletRepositoryTest) UpdateBalance(ctx context.Context, id uuid.UUID, balance decimal.Decimal) error {
	args := m.Called(ctx, id, balance)
	return args.Error(0)
}

// Mock transaction methods
func (m *MockWalletRepositoryTest) BeginTx(ctx context.Context) (*sql.Tx, error) {
	args := m.Called(ctx)
	return args.Get(0).(*sql.Tx), args.Error(1)
}

func (m *MockWalletRepositoryTest) UpdateBalanceWithTx(ctx context.Context, tx *sql.Tx, id uuid.UUID, balance decimal.Decimal) error {
	args := m.Called(ctx, tx, id, balance)
	return args.Error(0)
}

func (m *MockWalletRepositoryTest) GetWalletByIDWithTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) (*models.Wallet, error) {
	args := m.Called(ctx, tx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

// MockTransactionRepository for testing
type MockTransactionRepositoryTest struct {
	mock.Mock
}

func (m *MockTransactionRepositoryTest) CreateTransaction(ctx context.Context, transaction *models.Transaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}

func (m *MockTransactionRepositoryTest) CreateTransactionWithTx(ctx context.Context, tx *sql.Tx, transaction *models.Transaction) error {
	args := m.Called(ctx, tx, transaction)
	return args.Error(0)
}

func (m *MockTransactionRepositoryTest) GetTransactionsByWalletID(ctx context.Context, walletID uuid.UUID) ([]*models.Transaction, error) {
	args := m.Called(ctx, walletID)
	return args.Get(0).([]*models.Transaction), args.Error(1)
}

func TestWalletDepositValidAmount(t *testing.T) {
	service, walletRepo, transactionRepo := setupWalletService()

	walletID := uuid.New()
	wallet := createTestWallet(walletID, testWalletBalance)
	depositAmount := decimal.NewFromFloat(testDepositAmount)
	expectedBalance := decimal.NewFromFloat(testWalletBalance + testDepositAmount)

	walletRepo.On("BeginTx", mock.Anything).Return((*sql.Tx)(nil), nil)
	walletRepo.On("GetWalletByIDWithTx", mock.Anything, (*sql.Tx)(nil), walletID).Return(wallet, nil)
	walletRepo.On("UpdateBalanceWithTx", mock.Anything, (*sql.Tx)(nil), walletID, expectedBalance).Return(nil)
	transactionRepo.On("CreateTransactionWithTx", mock.Anything, (*sql.Tx)(nil), mock.AnythingOfType("*models.Transaction")).Return(nil)

	result, err := service.Deposit(context.Background(), walletID, depositAmount)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Balance.Equal(expectedBalance))
	walletRepo.AssertExpectations(t)
	transactionRepo.AssertExpectations(t)
}

func TestWalletWithdrawValidAmount(t *testing.T) {
	service, walletRepo, transactionRepo := setupWalletService()

	walletID := uuid.New()
	wallet := createTestWallet(walletID, testWalletBalance)
	withdrawAmount := decimal.NewFromFloat(testWithdrawAmount)
	expectedBalance := decimal.NewFromFloat(testWalletBalance - testWithdrawAmount)

	walletRepo.On("BeginTx", mock.Anything).Return((*sql.Tx)(nil), nil)
	walletRepo.On("GetWalletByIDWithTx", mock.Anything, (*sql.Tx)(nil), walletID).Return(wallet, nil)
	walletRepo.On("UpdateBalanceWithTx", mock.Anything, (*sql.Tx)(nil), walletID, expectedBalance).Return(nil)
	transactionRepo.On("CreateTransactionWithTx", mock.Anything, (*sql.Tx)(nil), mock.AnythingOfType("*models.Transaction")).Return(nil)

	result, err := service.Withdraw(context.Background(), walletID, withdrawAmount)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Balance.Equal(expectedBalance))
	walletRepo.AssertExpectations(t)
	transactionRepo.AssertExpectations(t)
}

func TestWalletWithdrawInsufficientBalance(t *testing.T) {
	walletRepo := new(MockWalletRepositoryTest)
	transactionRepo := new(MockTransactionRepositoryTest)

	service := &WalletService{
		WalletRepo:      walletRepo,
		TransactionRepo: transactionRepo,
	}

	walletID := uuid.New()
	wallet := &models.Wallet{
		ID:      walletID,
		Balance: decimal.NewFromFloat(30.0),
	}

	withdrawAmount := decimal.NewFromFloat(50.0)

	walletRepo.On("BeginTx", mock.Anything).Return((*sql.Tx)(nil), nil)
	walletRepo.On("GetWalletByIDWithTx", mock.Anything, (*sql.Tx)(nil), walletID).Return(wallet, nil)

	result, err := service.Withdraw(context.Background(), walletID, withdrawAmount)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "insufficient balance")
	walletRepo.AssertExpectations(t)
}

func TestWalletGetBalance(t *testing.T) {
	walletRepo := new(MockWalletRepositoryTest)
	service := &WalletService{WalletRepo: walletRepo}

	walletID := uuid.New()
	wallet := &models.Wallet{
		ID:      walletID,
		Balance: decimal.NewFromFloat(100.0),
	}

	walletRepo.On("GetWalletByID", mock.Anything, walletID).Return(wallet, nil)

	result, err := service.GetBalance(context.Background(), walletID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Balance.Equal(decimal.NewFromFloat(100.0)))
	walletRepo.AssertExpectations(t)
}

func TestWalletDepositNegativeAmount(t *testing.T) {
	walletRepo := new(MockWalletRepositoryTest)
	transactionRepo := new(MockTransactionRepositoryTest)
	service := &WalletService{
		WalletRepo:      walletRepo,
		TransactionRepo: transactionRepo,
	}

	walletID := uuid.New()
	negativeAmount := decimal.NewFromFloat(-10.0)

	result, err := service.Deposit(context.Background(), walletID, negativeAmount)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "deposit amount must be positive")
}

func TestWalletTransferValidation(t *testing.T) {
	walletRepo := new(MockWalletRepositoryTest)
	transactionRepo := new(MockTransactionRepositoryTest)
	service := &WalletService{
		WalletRepo:      walletRepo,
		TransactionRepo: transactionRepo,
	}

	fromWalletID := uuid.New()
	toWalletID := uuid.New()

	// Test negative amount
	err := service.Transfer(context.Background(), fromWalletID, toWalletID, decimal.NewFromFloat(-10.0), "Test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transfer amount must be positive")

	// Test same wallet transfer
	err = service.Transfer(context.Background(), fromWalletID, fromWalletID, decimal.NewFromFloat(10.0), "Test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot transfer to the same wallet")
}

func TestWalletTransferInsufficientBalance(t *testing.T) {
	walletRepo := new(MockWalletRepositoryTest)
	transactionRepo := new(MockTransactionRepositoryTest)

	service := &WalletService{
		WalletRepo:      walletRepo,
		TransactionRepo: transactionRepo,
	}

	fromWalletID := uuid.New()
	toWalletID := uuid.New()
	transferAmount := decimal.NewFromFloat(150.0)

	fromWallet := &models.Wallet{
		ID:      fromWalletID,
		Balance: decimal.NewFromFloat(100.0),
	}

	toWallet := &models.Wallet{
		ID:      toWalletID,
		Balance: decimal.NewFromFloat(25.0),
	}

	walletRepo.On("BeginTx", mock.Anything).Return((*sql.Tx)(nil), nil)
	walletRepo.On("GetWalletByIDWithTx", mock.Anything, (*sql.Tx)(nil), fromWalletID).Return(fromWallet, nil)
	walletRepo.On("GetWalletByIDWithTx", mock.Anything, (*sql.Tx)(nil), toWalletID).Return(toWallet, nil)

	err := service.Transfer(context.Background(), fromWalletID, toWalletID, transferAmount, "Test transfer")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient balance")
	walletRepo.AssertExpectations(t)
}

func TestWalletGetTransactionHistory(t *testing.T) {
	walletRepo := new(MockWalletRepositoryTest)
	transactionRepo := new(MockTransactionRepositoryTest)
	service := &WalletService{
		WalletRepo:      walletRepo,
		TransactionRepo: transactionRepo,
	}

	walletID := uuid.New()
	wallet := &models.Wallet{
		ID:      walletID,
		Balance: decimal.NewFromFloat(100.0),
	}

	// Mock transactions
	transactions := []*models.Transaction{
		{
			ID:       uuid.New(),
			WalletID: walletID,
			Type:     "deposit",
			Amount:   decimal.NewFromFloat(50.0),
		},
		{
			ID:       uuid.New(),
			WalletID: walletID,
			Type:     "withdraw",
			Amount:   decimal.NewFromFloat(25.0),
		},
	}

	walletRepo.On("GetWalletByID", mock.Anything, walletID).Return(wallet, nil)
	transactionRepo.On("GetTransactionsByWalletID", mock.Anything, walletID).Return(transactions, nil)

	result, err := service.GetTransactionHistory(context.Background(), walletID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, "deposit", result[0].Type)
	assert.Equal(t, "withdraw", result[1].Type)
	walletRepo.AssertExpectations(t)
	transactionRepo.AssertExpectations(t)
}

// Tests for assignment requirements - edge cases and validation

func TestWalletDepositZeroAmount(t *testing.T) {
	walletRepo := new(MockWalletRepositoryTest)
	transactionRepo := new(MockTransactionRepositoryTest)
	service := &WalletService{
		WalletRepo:      walletRepo,
		TransactionRepo: transactionRepo,
	}

	walletID := uuid.New()
	zeroAmount := decimal.Zero

	result, err := service.Deposit(context.Background(), walletID, zeroAmount)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "deposit amount must be positive")
}

func TestWalletWithdrawZeroAmount(t *testing.T) {
	walletRepo := new(MockWalletRepositoryTest)
	transactionRepo := new(MockTransactionRepositoryTest)
	service := &WalletService{
		WalletRepo:      walletRepo,
		TransactionRepo: transactionRepo,
	}

	walletID := uuid.New()
	wallet := createTestWallet(walletID, testWalletBalance)
	zeroAmount := decimal.Zero

	walletRepo.On("BeginTx", mock.Anything).Return((*sql.Tx)(nil), nil)
	walletRepo.On("GetWalletByIDWithTx", mock.Anything, (*sql.Tx)(nil), walletID).Return(wallet, nil)

	result, err := service.Withdraw(context.Background(), walletID, zeroAmount)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "withdraw amount must be positive")
	walletRepo.AssertExpectations(t)
}

func TestWalletTransferZeroAmount(t *testing.T) {
	walletRepo := new(MockWalletRepositoryTest)
	transactionRepo := new(MockTransactionRepositoryTest)
	service := &WalletService{
		WalletRepo:      walletRepo,
		TransactionRepo: transactionRepo,
	}

	fromWalletID := uuid.New()
	toWalletID := uuid.New()
	zeroAmount := decimal.Zero

	err := service.Transfer(context.Background(), fromWalletID, toWalletID, zeroAmount, "Test")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transfer amount must be positive")
}

func TestWalletDecimalPrecision(t *testing.T) {
	// Test that decimal calculations maintain precision
	amount1 := decimal.NewFromFloat(0.1)
	amount2 := decimal.NewFromFloat(0.2)
	sum := amount1.Add(amount2)
	expected := decimal.NewFromFloat(0.3)

	assert.True(t, sum.Equal(expected), "Decimal precision must be maintained")

	// Test large numbers
	large := decimal.NewFromFloat(999999999.99)
	small := decimal.NewFromFloat(0.01)
	result := large.Add(small)
	expectedLarge := decimal.NewFromFloat(1000000000.00)
	assert.True(t, result.Equal(expectedLarge), "Large number precision must be maintained")
}

func TestMoneyFormatting(t *testing.T) {
	// Test money formatting for display purposes
	amount := decimal.NewFromFloat(123.456789)

	rounded := amount.Round(2)
	expected := decimal.NewFromFloat(123.46)
	assert.True(t, rounded.Equal(expected))

	// Test zero handling
	zero := decimal.Zero
	assert.True(t, zero.IsZero())
	assert.Equal(t, "0", zero.String())
}
