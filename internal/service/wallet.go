package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/shanwije/wallet-app/internal/models"
	"github.com/shanwije/wallet-app/internal/repository"
	"github.com/shopspring/decimal"
)

type WalletService struct {
	WalletRepo      repository.WalletRepository
	TransactionRepo repository.TransactionRepository
}

func (s *WalletService) Deposit(walletID uuid.UUID, amount decimal.Decimal) (*models.Wallet, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("deposit amount must be positive")
	}

	// Begin database transaction for atomicity
	tx, err := s.WalletRepo.BeginTx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if tx != nil {
				tx.Rollback()
			}
		}
	}()

	// Get current wallet
	wallet, err := s.WalletRepo.GetWalletByIDWithTx(tx, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	// Update balance
	newBalance := wallet.Balance.Add(amount)
	err = s.WalletRepo.UpdateBalanceWithTx(tx, walletID, newBalance)
	if err != nil {
		return nil, fmt.Errorf("failed to update wallet balance: %w", err)
	}

	// Record transaction
	transaction := &models.Transaction{
		WalletID:    walletID,
		Type:        "deposit",
		Amount:      amount,
		Description: nil, // Optional description can be added later
	}

	err = s.TransactionRepo.CreateTransactionWithTx(tx, transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to record transaction: %w", err)
	}

	// Commit transaction
	if tx != nil {
		err = tx.Commit()
		if err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}
	}

	// Return updated wallet
	wallet.Balance = newBalance
	return wallet, nil
}

func (s *WalletService) Withdraw(walletID uuid.UUID, amount decimal.Decimal) (*models.Wallet, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("withdrawal amount must be positive")
	}

	// Begin database transaction for atomicity
	tx, err := s.WalletRepo.BeginTx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if tx != nil {
				tx.Rollback()
			}
		}
	}()

	// Get current wallet
	wallet, err := s.WalletRepo.GetWalletByIDWithTx(tx, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	// Check sufficient balance
	if wallet.Balance.LessThan(amount) {
		return nil, fmt.Errorf("insufficient balance")
	}

	// Update balance
	newBalance := wallet.Balance.Sub(amount)
	err = s.WalletRepo.UpdateBalanceWithTx(tx, walletID, newBalance)
	if err != nil {
		return nil, fmt.Errorf("failed to update wallet balance: %w", err)
	}

	// Record transaction
	transaction := &models.Transaction{
		WalletID:    walletID,
		Type:        "withdrawal",
		Amount:      amount,
		Description: nil, // Optional description can be added later
	}

	err = s.TransactionRepo.CreateTransactionWithTx(tx, transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to record transaction: %w", err)
	}

	// Commit transaction
	if tx != nil {
		err = tx.Commit()
		if err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}
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

// Transfer money between wallets atomically
func (s *WalletService) Transfer(fromWalletID, toWalletID uuid.UUID, amount decimal.Decimal, description string) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("transfer amount must be positive")
	}

	if fromWalletID == toWalletID {
		return fmt.Errorf("cannot transfer to the same wallet")
	}

	// Begin database transaction for atomicity
	tx, err := s.WalletRepo.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if tx != nil {
				tx.Rollback()
			}
		}
	}()

	// Lock and get source wallet (FOR UPDATE to prevent race conditions)
	fromWallet, err := s.WalletRepo.GetWalletByIDWithTx(tx, fromWalletID)
	if err != nil {
		return fmt.Errorf("failed to get source wallet: %w", err)
	}

	// Lock and get destination wallet
	toWallet, err := s.WalletRepo.GetWalletByIDWithTx(tx, toWalletID)
	if err != nil {
		return fmt.Errorf("failed to get destination wallet: %w", err)
	}

	// Check sufficient balance
	if fromWallet.Balance.LessThan(amount) {
		return fmt.Errorf("insufficient balance")
	}

	// Calculate new balances
	newFromBalance := fromWallet.Balance.Sub(amount)
	newToBalance := toWallet.Balance.Add(amount)

	// Update balances
	err = s.WalletRepo.UpdateBalanceWithTx(tx, fromWalletID, newFromBalance)
	if err != nil {
		return fmt.Errorf("failed to update source wallet balance: %w", err)
	}

	err = s.WalletRepo.UpdateBalanceWithTx(tx, toWalletID, newToBalance)
	if err != nil {
		return fmt.Errorf("failed to update destination wallet balance: %w", err)
	}

	// Create reference ID for linking both transaction records
	referenceID := uuid.New()

	// Create outbound transaction record
	outTransaction := &models.Transaction{
		WalletID:    fromWalletID,
		Type:        "transfer_out",
		Amount:      amount,
		ReferenceID: &referenceID,
		Description: &description,
	}

	err = s.TransactionRepo.CreateTransactionWithTx(tx, outTransaction)
	if err != nil {
		return fmt.Errorf("failed to create outbound transaction: %w", err)
	}

	// Create inbound transaction record
	inTransaction := &models.Transaction{
		WalletID:    toWalletID,
		Type:        "transfer_in",
		Amount:      amount,
		ReferenceID: &referenceID,
		Description: &description,
	}

	err = s.TransactionRepo.CreateTransactionWithTx(tx, inTransaction)
	if err != nil {
		return fmt.Errorf("failed to create inbound transaction: %w", err)
	}

	// Commit transaction
	if tx != nil {
		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}

	return nil
}

// GetTransactionHistory gets transaction history for a wallet
func (s *WalletService) GetTransactionHistory(walletID uuid.UUID) ([]*models.Transaction, error) {
	// First verify the wallet exists
	_, err := s.WalletRepo.GetWalletByID(walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	// Get transaction history
	transactions, err := s.TransactionRepo.GetTransactionsByWalletID(walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction history: %w", err)
	}

	return transactions, nil
}
