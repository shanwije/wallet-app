package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/shanwije/wallet-app/internal/models"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, name string) (*models.User, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserWithWallet(ctx context.Context, id uuid.UUID) (*models.UserWithWallet, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserWithWallet), args.Error(1)
}

// MockWalletRepository is a mock implementation of WalletRepository
type MockWalletRepository struct {
	mock.Mock
}

func (m *MockWalletRepository) CreateWallet(ctx context.Context, userID uuid.UUID) (*models.Wallet, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletRepository) GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*models.Wallet, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletRepository) GetWalletByID(ctx context.Context, id uuid.UUID) (*models.Wallet, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletRepository) UpdateBalance(ctx context.Context, id uuid.UUID, balance decimal.Decimal) error {
	args := m.Called(ctx, id, balance)
	return args.Error(0)
}

// Transaction support methods (not used in user tests, but required by interface)
func (m *MockWalletRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	args := m.Called(ctx)
	return nil, args.Error(1) // Return nil for Tx as it's not used in user tests
}

func (m *MockWalletRepository) UpdateBalanceWithTx(ctx context.Context, tx *sql.Tx, id uuid.UUID, balance decimal.Decimal) error {
	args := m.Called(ctx, tx, id, balance)
	return args.Error(0)
}

func (m *MockWalletRepository) GetWalletByIDWithTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) (*models.Wallet, error) {
	args := m.Called(ctx, tx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
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
		Balance:   decimal.Zero,
		CreatedAt: now,
	}

	userRepo.On("CreateUser", mock.Anything, "John Doe").Return(expectedUser, nil)
	walletRepo.On("CreateWallet", mock.Anything, userID).Return(expectedWallet, nil)

	result, err := service.CreateUser(context.Background(), "John Doe")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
	assert.Equal(t, "John Doe", result.Name)
	assert.Equal(t, walletID, result.Wallet.ID)
	assert.True(t, result.Wallet.Balance.Equal(decimal.Zero))

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

	userRepo.On("CreateUser", mock.Anything, "John Doe").Return(nil, errors.New("database error"))

	result, err := service.CreateUser(context.Background(), "John Doe")

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
			Balance:   decimal.NewFromFloat(100.0),
			CreatedAt: now,
		},
		CreatedAt: now,
	}

	userRepo.On("GetUserWithWallet", mock.Anything, userID).Return(expectedUserWithWallet, nil)

	result, err := service.GetUserWithWallet(context.Background(), userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
	assert.Equal(t, "John Doe", result.Name)
	assert.Equal(t, walletID, result.Wallet.ID)
	assert.True(t, result.Wallet.Balance.Equal(decimal.NewFromFloat(100.0)))

	userRepo.AssertExpectations(t)
}

// Tests for assignment requirements - user validation

func TestCreateUserEmptyName(t *testing.T) {
	userRepo := new(MockUserRepository)
	walletRepo := new(MockWalletRepository)
	service := &UserService{
		UserRepo:   userRepo,
		WalletRepo: walletRepo,
	}

	// Test empty name validation - should fail early, no repository calls expected
	result, err := service.CreateUser(context.Background(), "")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "name cannot be empty")

	// No mock expectations needed since validation should fail before repository calls
}

func TestUUIDValidation(t *testing.T) {
	// Test UUID generation and validation
	validID := uuid.New()
	nilID := uuid.Nil

	assert.NotEqual(t, validID, nilID, "Generated UUIDs should not be nil")
	assert.NotEmpty(t, validID.String(), "UUID string representation should not be empty")

	// Test UUID parsing
	parsedID, err := uuid.Parse(validID.String())
	assert.NoError(t, err)
	assert.Equal(t, validID, parsedID)

	// Test invalid UUID parsing
	_, err = uuid.Parse("invalid-uuid")
	assert.Error(t, err, "Invalid UUID should cause parsing error")
}
