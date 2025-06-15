package repository

import (
	"github.com/google/uuid"
	"github.com/shanwije/wallet-app/internal/models"
	"github.com/shopspring/decimal"
)

type UserRepository interface {
	CreateUser(name string) (*models.User, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	GetUserWithWallet(id uuid.UUID) (*models.UserWithWallet, error)
}

type WalletRepository interface {
	CreateWallet(userID uuid.UUID) (*models.Wallet, error)
	GetWalletByUserID(userID uuid.UUID) (*models.Wallet, error)
	GetWalletByID(id uuid.UUID) (*models.Wallet, error)
	UpdateBalance(id uuid.UUID, balance decimal.Decimal) error
}
