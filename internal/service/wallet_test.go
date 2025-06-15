package service

import (
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/shanwije/wallet-app/internal/models"
)

// MockWalletRepository for testing
type MockWalletRepositoryTest struct {
	mock.Mock
}

func (m *MockWalletRepositoryTest) CreateWallet(userID uuid.UUID) (*models.Wallet, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletRepositoryTest) GetWalletByUserID(userID uuid.UUID) (*models.Wallet, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletRepositoryTest) GetWalletByID(id uuid.UUID) (*models.Wallet, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletRepositoryTest) UpdateBalance(id uuid.UUID, balance decimal.Decimal) error {
	args := m.Called(id, balance)
	return args.Error(0)
}

// Mock transaction methods
func (m *MockWalletRepositoryTest) BeginTx() (*sql.Tx, error) {
	args := m.Called()
	return args.Get(0).(*sql.Tx), args.Error(1)
}

func (m *MockWalletRepositoryTest) UpdateBalanceWithTx(tx *sql.Tx, id uuid.UUID, balance decimal.Decimal) error {
	args := m.Called(tx, id, balance)
	return args.Error(0)
}

func (m *MockWalletRepositoryTest) GetWalletByIDWithTx(tx *sql.Tx, id uuid.UUID) (*models.Wallet, error) {
	args := m.Called(tx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

// MockTransactionRepository for testing
type MockTransactionRepositoryTest struct {
	mock.Mock
}

func (m *MockTransactionRepositoryTest) CreateTransaction(transaction *models.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *MockTransactionRepositoryTest) CreateTransactionWithTx(tx *sql.Tx, transaction *models.Transaction) error {
	args := m.Called(tx, transaction)
	return args.Error(0)
}

func (m *MockTransactionRepositoryTest) GetTransactionsByWalletID(walletID uuid.UUID) ([]*models.Transaction, error) {
	args := m.Called(walletID)
	return args.Get(0).([]*models.Transaction), args.Error(1)
}

func TestWalletDeposit(t *testing.T) {
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

	expectedBalance := decimal.NewFromFloat(150.0)
	depositAmount := decimal.NewFromFloat(50.0)

	// Use nil transaction for simplicity in unit tests
	walletRepo.On("BeginTx").Return((*sql.Tx)(nil), nil)
	walletRepo.On("GetWalletByIDWithTx", (*sql.Tx)(nil), walletID).Return(wallet, nil)
	walletRepo.On("UpdateBalanceWithTx", (*sql.Tx)(nil), walletID, expectedBalance).Return(nil)
	transactionRepo.On("CreateTransactionWithTx", (*sql.Tx)(nil), mock.AnythingOfType("*models.Transaction")).Return(nil)

	result, err := service.Deposit(walletID, depositAmount)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Balance.Equal(expectedBalance))
	walletRepo.AssertExpectations(t)
	transactionRepo.AssertExpectations(t)
}

func TestWalletWithdraw(t *testing.T) {
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

	expectedBalance := decimal.NewFromFloat(50.0)
	withdrawAmount := decimal.NewFromFloat(50.0)

	walletRepo.On("BeginTx").Return((*sql.Tx)(nil), nil)
	walletRepo.On("GetWalletByIDWithTx", (*sql.Tx)(nil), walletID).Return(wallet, nil)
	walletRepo.On("UpdateBalanceWithTx", (*sql.Tx)(nil), walletID, expectedBalance).Return(nil)
	transactionRepo.On("CreateTransactionWithTx", (*sql.Tx)(nil), mock.AnythingOfType("*models.Transaction")).Return(nil)

	result, err := service.Withdraw(walletID, withdrawAmount)

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

	walletRepo.On("BeginTx").Return((*sql.Tx)(nil), nil)
	walletRepo.On("GetWalletByIDWithTx", (*sql.Tx)(nil), walletID).Return(wallet, nil)

	result, err := service.Withdraw(walletID, withdrawAmount)

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

	walletRepo.On("GetWalletByID", walletID).Return(wallet, nil)

	result, err := service.GetBalance(walletID)

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

	result, err := service.Deposit(walletID, negativeAmount)

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
	err := service.Transfer(fromWalletID, toWalletID, decimal.NewFromFloat(-10.0), "Test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transfer amount must be positive")

	// Test same wallet transfer
	err = service.Transfer(fromWalletID, fromWalletID, decimal.NewFromFloat(10.0), "Test")
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

	walletRepo.On("BeginTx").Return((*sql.Tx)(nil), nil)
	walletRepo.On("GetWalletByIDWithTx", (*sql.Tx)(nil), fromWalletID).Return(fromWallet, nil)
	walletRepo.On("GetWalletByIDWithTx", (*sql.Tx)(nil), toWalletID).Return(toWallet, nil)

	err := service.Transfer(fromWalletID, toWalletID, transferAmount, "Test transfer")

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
			Type:     "withdrawal",
			Amount:   decimal.NewFromFloat(25.0),
		},
	}

	walletRepo.On("GetWalletByID", walletID).Return(wallet, nil)
	transactionRepo.On("GetTransactionsByWalletID", walletID).Return(transactions, nil)

	result, err := service.GetTransactionHistory(walletID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, "deposit", result[0].Type)
	assert.Equal(t, "withdrawal", result[1].Type)
	walletRepo.AssertExpectations(t)
	transactionRepo.AssertExpectations(t)
}
