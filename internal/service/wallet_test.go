package service

import (
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

func TestWalletDeposit(t *testing.T) {
	walletRepo := new(MockWalletRepositoryTest)
	service := &WalletService{WalletRepo: walletRepo}

	walletID := uuid.New()
	wallet := &models.Wallet{
		ID:      walletID,
		Balance: decimal.NewFromFloat(100.0),
	}

	expectedBalance := decimal.NewFromFloat(150.0)
	depositAmount := decimal.NewFromFloat(50.0)

	walletRepo.On("GetWalletByID", walletID).Return(wallet, nil)
	walletRepo.On("UpdateBalance", walletID, expectedBalance).Return(nil)

	result, err := service.Deposit(walletID, depositAmount)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Balance.Equal(expectedBalance))
	walletRepo.AssertExpectations(t)
}

func TestWalletWithdraw(t *testing.T) {
	walletRepo := new(MockWalletRepositoryTest)
	service := &WalletService{WalletRepo: walletRepo}

	walletID := uuid.New()
	wallet := &models.Wallet{
		ID:      walletID,
		Balance: decimal.NewFromFloat(100.0),
	}

	expectedBalance := decimal.NewFromFloat(50.0)
	withdrawAmount := decimal.NewFromFloat(50.0)

	walletRepo.On("GetWalletByID", walletID).Return(wallet, nil)
	walletRepo.On("UpdateBalance", walletID, expectedBalance).Return(nil)

	result, err := service.Withdraw(walletID, withdrawAmount)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Balance.Equal(expectedBalance))
	walletRepo.AssertExpectations(t)
}

func TestWalletWithdrawInsufficientBalance(t *testing.T) {
	walletRepo := new(MockWalletRepositoryTest)
	service := &WalletService{WalletRepo: walletRepo}

	walletID := uuid.New()
	wallet := &models.Wallet{
		ID:      walletID,
		Balance: decimal.NewFromFloat(30.0),
	}

	withdrawAmount := decimal.NewFromFloat(50.0)

	walletRepo.On("GetWalletByID", walletID).Return(wallet, nil)

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
	service := &WalletService{WalletRepo: walletRepo}

	walletID := uuid.New()
	negativeAmount := decimal.NewFromFloat(-10.0)

	result, err := service.Deposit(walletID, negativeAmount)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "deposit amount must be positive")
}
