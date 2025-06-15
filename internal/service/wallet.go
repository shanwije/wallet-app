package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shanwije/wallet-app/internal/models"
	"github.com/shanwije/wallet-app/internal/repository"
	"github.com/shopspring/decimal"
)

// Transaction type constants for better readability
const (
	TransactionTypeDeposit  = "deposit"
	TransactionTypeWithdraw = "withdraw"
	TransactionTypeTransfer = "transfer"
)

type WalletService struct {
	WalletRepo      repository.WalletRepository
	TransactionRepo repository.TransactionRepository
}

// validateDepositAmount validates that the deposit amount is positive
func (s *WalletService) validateDepositAmount(amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("deposit amount must be positive")
	}
	return nil
}

// validateWithdrawAmount validates that the withdraw amount is positive and sufficient
func (s *WalletService) validateWithdrawAmount(amount decimal.Decimal, currentBalance decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("withdraw amount must be positive")
	}
	if currentBalance.LessThan(amount) {
		return fmt.Errorf("insufficient balance for withdrawal")
	}
	return nil
}

// validateTransferAmount validates transfer amount and wallets
func (s *WalletService) validateTransferAmount(amount decimal.Decimal, fromWalletID, toWalletID uuid.UUID) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("transfer amount must be positive")
	}
	if fromWalletID == toWalletID {
		return fmt.Errorf("cannot transfer to the same wallet")
	}
	return nil
}

func (s *WalletService) Deposit(ctx context.Context, walletID uuid.UUID, amount decimal.Decimal) (*models.Wallet, error) {
	// Validate input
	if err := s.validateDepositAmount(amount); err != nil {
		return nil, err
	}

	// Begin database transaction for atomicity
	tx, err := s.WalletRepo.BeginTx(ctx)
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
	wallet, err := s.WalletRepo.GetWalletByIDWithTx(ctx, tx, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	// Update balance
	newBalance := wallet.Balance.Add(amount)
	err = s.WalletRepo.UpdateBalanceWithTx(ctx, tx, walletID, newBalance)
	if err != nil {
		return nil, fmt.Errorf("failed to update wallet balance: %w", err)
	}

	// Record transaction
	transaction := &models.Transaction{
		WalletID:    walletID,
		Type:        TransactionTypeDeposit,
		Amount:      amount,
		Description: nil, // Optional description can be added later
	}

	err = s.TransactionRepo.CreateTransactionWithTx(ctx, tx, transaction)
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

func (s *WalletService) Withdraw(ctx context.Context, walletID uuid.UUID, amount decimal.Decimal) (*models.Wallet, error) {
	// Validate input
	wallet, err := s.WalletRepo.GetWalletByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	if err := s.validateWithdrawAmount(amount, wallet.Balance); err != nil {
		return nil, err
	}

	// Begin database transaction for atomicity
	tx, err := s.WalletRepo.BeginTx(ctx)
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

	// Update balance
	newBalance := wallet.Balance.Sub(amount)
	err = s.WalletRepo.UpdateBalanceWithTx(ctx, tx, walletID, newBalance)
	if err != nil {
		return nil, fmt.Errorf("failed to update wallet balance: %w", err)
	}

	// Record transaction
	transaction := &models.Transaction{
		WalletID:    walletID,
		Type:        TransactionTypeWithdraw,
		Amount:      amount,
		Description: nil, // Optional description can be added later
	}

	err = s.TransactionRepo.CreateTransactionWithTx(ctx, tx, transaction)
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

func (s *WalletService) GetBalance(ctx context.Context, walletID uuid.UUID) (*models.Wallet, error) {
	wallet, err := s.WalletRepo.GetWalletByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return wallet, nil
}

func (s *WalletService) GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*models.Wallet, error) {
	wallet, err := s.WalletRepo.GetWalletByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet for user: %w", err)
	}

	return wallet, nil
}

// Transfer money between wallets atomically
func (s *WalletService) Transfer(ctx context.Context, fromWalletID, toWalletID uuid.UUID, amount decimal.Decimal, description string) error {
	// Validate input
	if err := s.validateTransferAmount(amount, fromWalletID, toWalletID); err != nil {
		return err
	}

	// Begin database transaction for atomicity
	tx, err := s.WalletRepo.BeginTx(ctx)
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
	fromWallet, err := s.WalletRepo.GetWalletByIDWithTx(ctx, tx, fromWalletID)
	if err != nil {
		return fmt.Errorf("failed to get source wallet: %w", err)
	}

	// Lock and get destination wallet
	toWallet, err := s.WalletRepo.GetWalletByIDWithTx(ctx, tx, toWalletID)
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
	err = s.WalletRepo.UpdateBalanceWithTx(ctx, tx, fromWalletID, newFromBalance)
	if err != nil {
		return fmt.Errorf("failed to update source wallet balance: %w", err)
	}

	err = s.WalletRepo.UpdateBalanceWithTx(ctx, tx, toWalletID, newToBalance)
	if err != nil {
		return fmt.Errorf("failed to update destination wallet balance: %w", err)
	}

	// Create reference ID for linking both transaction records
	referenceID := uuid.New()

	// Create outbound transaction record
	outTransaction := &models.Transaction{
		WalletID:    fromWalletID,
		Type:        TransactionTypeTransfer + "_out",
		Amount:      amount,
		ReferenceID: &referenceID,
		Description: &description,
	}

	err = s.TransactionRepo.CreateTransactionWithTx(ctx, tx, outTransaction)
	if err != nil {
		return fmt.Errorf("failed to create outbound transaction: %w", err)
	}

	// Create inbound transaction record
	inTransaction := &models.Transaction{
		WalletID:    toWalletID,
		Type:        TransactionTypeTransfer + "_in",
		Amount:      amount,
		ReferenceID: &referenceID,
		Description: &description,
	}

	err = s.TransactionRepo.CreateTransactionWithTx(ctx, tx, inTransaction)
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
func (s *WalletService) GetTransactionHistory(ctx context.Context, walletID uuid.UUID) ([]*models.Transaction, error) {
	// First verify the wallet exists
	_, err := s.WalletRepo.GetWalletByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	// Get transaction history
	transactions, err := s.TransactionRepo.GetTransactionsByWalletID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction history: %w", err)
	}

	return transactions, nil
}
