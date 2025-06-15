package models

import (
	"time"
	"github.com/google/uuid"
)

type Transaction struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	WalletID    uuid.UUID  `db:"wallet_id" json:"wallet_id"`
	Type        string     `db:"type" json:"type"` // deposit, withdraw, transfer_in, transfer_out
	Amount      float64    `db:"amount" json:"amount"`
	ReferenceID *uuid.UUID `db:"reference_id" json:"reference_id,omitempty"`
	Description *string    `db:"description" json:"description,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
}
