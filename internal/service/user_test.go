package service

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/shanwije/wallet-app/internal/models"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(name string) (*models.User, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserWithWallet(id uuid.UUID) (*models.UserWithWallet, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserWithWallet), args.Error(1)
}

// MockWalletRepository is a mock implementation of WalletRepository
type MockWalletRepository struct {
	mock.Mock
}

func (m *MockWalletRepository) CreateWallet(userID uuid.UUID) (*models.Wallet, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletRepository) GetWalletByUserID(userID uuid.UUID) (*models.Wallet, error) {
	args := m.Called(userID)
	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletRepository) GetWalletByID(id uuid.UUID) (*models.Wallet, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletRepository) UpdateBalance(id uuid.UUID, balance float64) error {
	args := m.Called(id, balance)
	return args.Error(0)
}

// Core functionality test: Successful user creation with wallet
func TestCreateUser(t *testing.T) {
	userRepo := new(MockUserRepository)
	walletRepo := new(MockWalletRepository)
	service := &UserService{
		UserRepo:   userRepo,
		WalletRepo: walletRepo,
	}

	userID := uuid.New()
	walletID := uuid.New()
	now := time.Now().UTC()

	expectedUser := &models.User{
		ID:        userID,
		Name:      "John Doe",
		CreatedAt: now,
	}

	expectedWallet := &models.Wallet{
		ID:        walletID,
		UserID:    userID,
		Balance:   0.0,
		CreatedAt: now,
	}

	userRepo.On("CreateUser", "John Doe").Return(expectedUser, nil)
	walletRepo.On("CreateWallet", userID).Return(expectedWallet, nil)

	result, err := service.CreateUser("John Doe")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
	assert.Equal(t, "John Doe", result.Name)
	assert.Equal(t, walletID, result.Wallet.ID)
	assert.Equal(t, 0.0, result.Wallet.Balance)

	userRepo.AssertExpectations(t)
	walletRepo.AssertExpectations(t)
}

// Core functionality test: User creation failure
func TestCreateUserError(t *testing.T) {
	userRepo := new(MockUserRepository)
	walletRepo := new(MockWalletRepository)
	service := &UserService{
		UserRepo:   userRepo,
		WalletRepo: walletRepo,
	}

	userRepo.On("CreateUser", "John Doe").Return(nil, errors.New("database error"))

	result, err := service.CreateUser("John Doe")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create user")

	userRepo.AssertExpectations(t)
}

// Core functionality test: Get user with wallet
func TestGetUserWithWallet(t *testing.T) {
	userRepo := new(MockUserRepository)
	walletRepo := new(MockWalletRepository)
	service := &UserService{
		UserRepo:   userRepo,
		WalletRepo: walletRepo,
	}

	userID := uuid.New()
	walletID := uuid.New()
	now := time.Now().UTC()

	expectedUserWithWallet := &models.UserWithWallet{
		ID:   userID,
		Name: "John Doe",
		Wallet: models.Wallet{
			ID:        walletID,
			UserID:    userID,
			Balance:   100.0,
			CreatedAt: now,
		},
		CreatedAt: now,
	}

	userRepo.On("GetUserWithWallet", userID).Return(expectedUserWithWallet, nil)

	result, err := service.GetUserWithWallet(userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
	assert.Equal(t, "John Doe", result.Name)
	assert.Equal(t, walletID, result.Wallet.ID)
	assert.Equal(t, 100.0, result.Wallet.Balance)

	userRepo.AssertExpectations(t)
}
