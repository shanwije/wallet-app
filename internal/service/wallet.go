package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/shanwije/wallet-app/internal/models"
	"github.com/shanwije/wallet-app/internal/repository"
	"github.com/shopspring/decimal"
)

// Transaction type constants for better readability
const (
	TransactionTypeDeposit     = "deposit"
	TransactionTypeWithdraw    = "withdraw"
	TransactionTypeTransferOut = "transfer_out"
	TransactionTypeTransferIn  = "transfer_in"
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

	// Validate input amount and sufficient balance
	if err := s.validateWithdrawAmount(amount, wallet.Balance); err != nil {
		return nil, err
	}

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

// transferExecution handles the actual transfer logic within a transaction
func (s *WalletService) transferExecution(ctx context.Context, tx *sql.Tx, fromWalletID, toWalletID uuid.UUID, amount decimal.Decimal, description string) error {
	// Lock and get both wallets
	fromWallet, toWallet, err := s.lockAndGetWallets(ctx, tx, fromWalletID, toWalletID)
	if err != nil {
		return err
	}

	// Validate sufficient balance
	if fromWallet.Balance.LessThan(amount) {
		return fmt.Errorf("insufficient balance")
	}

	// Update balances
	if err := s.updateTransferBalances(ctx, tx, fromWalletID, toWalletID, fromWallet.Balance, toWallet.Balance, amount); err != nil {
		return err
	}

	// Create transaction records
	return s.createTransferRecords(ctx, tx, fromWalletID, toWalletID, amount, description)
}

// lockAndGetWallets locks and retrieves both wallets for transfer
func (s *WalletService) lockAndGetWallets(ctx context.Context, tx *sql.Tx, fromWalletID, toWalletID uuid.UUID) (*models.Wallet, *models.Wallet, error) {
	fromWallet, err := s.WalletRepo.GetWalletByIDWithTx(ctx, tx, fromWalletID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get source wallet: %w", err)
	}

	toWallet, err := s.WalletRepo.GetWalletByIDWithTx(ctx, tx, toWalletID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get destination wallet: %w", err)
	}

	return fromWallet, toWallet, nil
}

// updateTransferBalances updates both wallet balances
func (s *WalletService) updateTransferBalances(ctx context.Context, tx *sql.Tx, fromWalletID, toWalletID uuid.UUID, fromBalance, toBalance, amount decimal.Decimal) error {
	newFromBalance := fromBalance.Sub(amount)
	newToBalance := toBalance.Add(amount)

	if err := s.WalletRepo.UpdateBalanceWithTx(ctx, tx, fromWalletID, newFromBalance); err != nil {
		return fmt.Errorf("failed to update source wallet balance: %w", err)
	}

	if err := s.WalletRepo.UpdateBalanceWithTx(ctx, tx, toWalletID, newToBalance); err != nil {
		return fmt.Errorf("failed to update destination wallet balance: %w", err)
	}

	return nil
}

// createTransferRecords creates both transaction records for the transfer
func (s *WalletService) createTransferRecords(ctx context.Context, tx *sql.Tx, fromWalletID, toWalletID uuid.UUID, amount decimal.Decimal, description string) error {
	referenceID := uuid.New()

	outTransaction := &models.Transaction{
		WalletID:    fromWalletID,
		Type:        TransactionTypeTransferOut,
		Amount:      amount,
		ReferenceID: &referenceID,
		Description: &description,
	}

	if err := s.TransactionRepo.CreateTransactionWithTx(ctx, tx, outTransaction); err != nil {
		return fmt.Errorf("failed to create outbound transaction: %w", err)
	}

	inTransaction := &models.Transaction{
		WalletID:    toWalletID,
		Type:        TransactionTypeTransferIn,
		Amount:      amount,
		ReferenceID: &referenceID,
		Description: &description,
	}

	if err := s.TransactionRepo.CreateTransactionWithTx(ctx, tx, inTransaction); err != nil {
		return fmt.Errorf("failed to create inbound transaction: %w", err)
	}

	return nil
}

// Transfer money between wallets atomically
func (s *WalletService) Transfer(ctx context.Context, fromWalletID, toWalletID uuid.UUID, amount decimal.Decimal, description string) error {
	if err := s.validateTransferAmount(amount, fromWalletID, toWalletID); err != nil {
		return err
	}

	tx, err := s.WalletRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil && tx != nil {
			tx.Rollback()
		}
	}()

	if err := s.transferExecution(ctx, tx, fromWalletID, toWalletID, amount, description); err != nil {
		return err
	}

	if tx != nil {
		if err := tx.Commit(); err != nil {
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
