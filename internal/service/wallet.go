package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/shanwije/wallet-app/internal/models"
	"github.com/shanwije/wallet-app/internal/repository"
	"github.com/shopspring/decimal"
)

type WalletService struct {
	WalletRepo repository.WalletRepository
}

func (s *WalletService) Deposit(walletID uuid.UUID, amount decimal.Decimal) (*models.Wallet, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("deposit amount must be positive")
	}

	// Get current wallet
	wallet, err := s.WalletRepo.GetWalletByID(walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	// Update balance
	newBalance := wallet.Balance.Add(amount)
	err = s.WalletRepo.UpdateBalance(walletID, newBalance)
	if err != nil {
		return nil, fmt.Errorf("failed to update wallet balance: %w", err)
	}

	// Return updated wallet
	wallet.Balance = newBalance
	return wallet, nil
}

func (s *WalletService) Withdraw(walletID uuid.UUID, amount decimal.Decimal) (*models.Wallet, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("withdrawal amount must be positive")
	}

	// Get current wallet
	wallet, err := s.WalletRepo.GetWalletByID(walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	// Check sufficient balance
	if wallet.Balance.LessThan(amount) {
		return nil, fmt.Errorf("insufficient balance")
	}

	// Update balance
	newBalance := wallet.Balance.Sub(amount)
	err = s.WalletRepo.UpdateBalance(walletID, newBalance)
	if err != nil {
		return nil, fmt.Errorf("failed to update wallet balance: %w", err)
	}

	// Return updated wallet
	wallet.Balance = newBalance
	return wallet, nil
}

func (s *WalletService) GetBalance(walletID uuid.UUID) (*models.Wallet, error) {
	wallet, err := s.WalletRepo.GetWalletByID(walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return wallet, nil
}

func (s *WalletService) GetWalletByUserID(userID uuid.UUID) (*models.Wallet, error) {
	wallet, err := s.WalletRepo.GetWalletByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet for user: %w", err)
	}

	return wallet, nil
}
