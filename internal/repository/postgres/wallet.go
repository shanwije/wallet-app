package postgres

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shanwije/wallet-app/internal/models"
)

type WalletRepository struct {
	db *sqlx.DB
}

func NewWalletRepository(db *sqlx.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) CreateWallet(userID uuid.UUID) (*models.Wallet, error) {
	wallet := &models.Wallet{
		ID:      uuid.New(),
		UserID:  userID,
		Balance: 0.0,
	}

	query := `
		INSERT INTO wallets (id, user_id, balance) 
		VALUES ($1, $2, $3) 
		RETURNING created_at`

	err := r.db.QueryRow(query, wallet.ID, wallet.UserID, wallet.Balance).Scan(&wallet.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	return wallet, nil
}

func (r *WalletRepository) GetWalletByUserID(userID uuid.UUID) (*models.Wallet, error) {
	wallet := &models.Wallet{}
	query := `SELECT id, user_id, balance, created_at FROM wallets WHERE user_id = $1`

	err := r.db.Get(wallet, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("wallet not found")
		}
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return wallet, nil
}

func (r *WalletRepository) GetWalletByID(id uuid.UUID) (*models.Wallet, error) {
	wallet := &models.Wallet{}
	query := `SELECT id, user_id, balance, created_at FROM wallets WHERE id = $1`

	err := r.db.Get(wallet, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("wallet not found")
		}
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return wallet, nil
}

func (r *WalletRepository) UpdateBalance(id uuid.UUID, balance float64) error {
	query := `UPDATE wallets SET balance = $1 WHERE id = $2`
	
	result, err := r.db.Exec(query, balance, id)
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
