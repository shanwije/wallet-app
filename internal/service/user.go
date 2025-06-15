package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/shanwije/wallet-app/internal/models"
	"github.com/shanwije/wallet-app/internal/repository"
)

type UserService struct {
	UserRepo   repository.UserRepository
	WalletRepo repository.WalletRepository
}

func (s *UserService) CreateUser(name string) (*models.UserWithWallet, error) {
	// Create user
	user, err := s.UserRepo.CreateUser(name)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create wallet for the user
	wallet, err := s.WalletRepo.CreateWallet(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet for user: %w", err)
	}

	// Return user with wallet
	return &models.UserWithWallet{
		ID:   user.ID,
		Name: user.Name,

		Wallet:    *wallet,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *UserService) GetUserWithWallet(id uuid.UUID) (*models.UserWithWallet, error) {
	userWithWallet, err := s.UserRepo.GetUserWithWallet(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user with wallet: %w", err)
	}

	return userWithWallet, nil
}
