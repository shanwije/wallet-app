package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shanwije/wallet-app/internal/models"
)

type TransactionRepository struct {
	db *sqlx.DB
}

func NewTransactionRepository(db *sqlx.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) CreateTransaction(ctx context.Context, transaction *models.Transaction) error {
	transaction.ID = uuid.New()

	query := `
		INSERT INTO transactions (id, wallet_id, type, amount, reference_id, description) 
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING created_at`

	err := r.db.QueryRowContext(ctx, query,
		transaction.ID,
		transaction.WalletID,
		transaction.Type,
		transaction.Amount,
		transaction.ReferenceID,
		transaction.Description,
	).Scan(&transaction.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

func (r *TransactionRepository) CreateTransactionWithTx(ctx context.Context, tx *sql.Tx, transaction *models.Transaction) error {
	transaction.ID = uuid.New()

	query := `
		INSERT INTO transactions (id, wallet_id, type, amount, reference_id, description) 
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING created_at`

	err := tx.QueryRowContext(ctx, query,
		transaction.ID,
		transaction.WalletID,
		transaction.Type,
		transaction.Amount,
		transaction.ReferenceID,
		transaction.Description,
	).Scan(&transaction.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

func (r *TransactionRepository) GetTransactionsByWalletID(ctx context.Context, walletID uuid.UUID) ([]*models.Transaction, error) {
	var transactions []*models.Transaction

	query := `
		SELECT id, wallet_id, type, amount, reference_id, description, created_at 
		FROM transactions 
		WHERE wallet_id = $1 
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		transaction := &models.Transaction{}
		err := rows.Scan(
			&transaction.ID,
			&transaction.WalletID,
			&transaction.Type,
			&transaction.Amount,
			&transaction.ReferenceID,
			&transaction.Description,
			&transaction.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, transaction)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("transaction rows error: %w", err)
	}

	return transactions, nil
}
