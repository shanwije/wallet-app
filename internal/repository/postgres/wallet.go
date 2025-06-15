package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shanwije/wallet-app/internal/models"
	"github.com/shopspring/decimal"
)

type WalletRepository struct {
	db *sqlx.DB
}

func NewWalletRepository(db *sqlx.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) CreateWallet(ctx context.Context, userID uuid.UUID) (*models.Wallet, error) {
	wallet := &models.Wallet{
		ID:      uuid.New(),
		UserID:  userID,
		Balance: decimal.Zero,
	}

	query := `
		INSERT INTO wallets (id, user_id, balance) 
		VALUES ($1, $2, $3) 
		RETURNING created_at`

	err := r.db.QueryRowContext(ctx, query, wallet.ID, wallet.UserID, wallet.Balance).Scan(&wallet.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	return wallet, nil
}

func (r *WalletRepository) GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*models.Wallet, error) {
	wallet := &models.Wallet{}
	query := `SELECT id, user_id, balance, created_at FROM wallets WHERE user_id = $1`

	err := r.db.GetContext(ctx, wallet, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("wallet not found for user ID: %s", userID)
		}
		return nil, fmt.Errorf("failed to get wallet by user ID: %w", err)
	}

	return wallet, nil
}

func (r *WalletRepository) GetWalletByID(ctx context.Context, id uuid.UUID) (*models.Wallet, error) {
	wallet := &models.Wallet{}
	query := `SELECT id, user_id, balance, created_at FROM wallets WHERE id = $1`

	err := r.db.GetContext(ctx, wallet, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("wallet not found")
		}
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return wallet, nil
}

func (r *WalletRepository) UpdateBalance(ctx context.Context, id uuid.UUID, balance decimal.Decimal) error {
	query := `UPDATE wallets SET balance = $1 WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, balance, id)
	if err != nil {
		return fmt.Errorf("failed to update wallet balance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("wallet not found")
	}

	return nil
}

// Transaction support methods
func (r *WalletRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func (r *WalletRepository) UpdateBalanceWithTx(ctx context.Context, tx *sql.Tx, id uuid.UUID, balance decimal.Decimal) error {
	query := `UPDATE wallets SET balance = $1 WHERE id = $2`

	result, err := tx.ExecContext(ctx, query, balance, id)
	if err != nil {
		return fmt.Errorf("failed to update wallet balance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("wallet not found")
	}

	return nil
}

func (r *WalletRepository) GetWalletByIDWithTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) (*models.Wallet, error) {
	wallet := &models.Wallet{}
	query := `SELECT id, user_id, balance, created_at FROM wallets WHERE id = $1 FOR UPDATE`

	err := tx.QueryRowContext(ctx, query, id).Scan(&wallet.ID, &wallet.UserID, &wallet.Balance, &wallet.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("wallet not found")
		}
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return wallet, nil
}
