package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/shanwije/wallet-app/internal/models"
	"github.com/shopspring/decimal"
)

type UserRepository interface {
	CreateUser(ctx context.Context, name string) (*models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserWithWallet(ctx context.Context, id uuid.UUID) (*models.UserWithWallet, error)
}

type WalletRepository interface {
	CreateWallet(ctx context.Context, userID uuid.UUID) (*models.Wallet, error)
	GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*models.Wallet, error)
	GetWalletByID(ctx context.Context, id uuid.UUID) (*models.Wallet, error)
	UpdateBalance(ctx context.Context, id uuid.UUID, balance decimal.Decimal) error
	// Transaction support for atomic operations
	BeginTx(ctx context.Context) (*sql.Tx, error)
	UpdateBalanceWithTx(ctx context.Context, tx *sql.Tx, id uuid.UUID, balance decimal.Decimal) error
	GetWalletByIDWithTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) (*models.Wallet, error)
}

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, transaction *models.Transaction) error
	CreateTransactionWithTx(ctx context.Context, tx *sql.Tx, transaction *models.Transaction) error
	GetTransactionsByWalletID(ctx context.Context, walletID uuid.UUID) ([]*models.Transaction, error)
}
