package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Transaction types as constants
const (
	TransactionTypeDeposit     = "deposit"
	TransactionTypeWithdraw    = "withdraw"
	TransactionTypeTransferIn  = "transfer_in"
	TransactionTypeTransferOut = "transfer_out"
)

type Transaction struct {
	ID          uuid.UUID       `db:"id" json:"id"`
	WalletID    uuid.UUID       `db:"wallet_id" json:"wallet_id"`
	Type        string          `db:"type" json:"type"` // deposit, withdraw, transfer_in, transfer_out
	Amount      decimal.Decimal `db:"amount" json:"amount"`
	ReferenceID *uuid.UUID      `db:"reference_id" json:"reference_id,omitempty"`
	Description *string         `db:"description" json:"description,omitempty"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
}

// IsValidTransactionType validates transaction type
func IsValidTransactionType(txType string) bool {
	switch txType {
	case TransactionTypeDeposit, TransactionTypeWithdraw, TransactionTypeTransferIn, TransactionTypeTransferOut:
		return true
	default:
		return false
	}
}
