package models

import "time"

type Transaction struct {
	ID          string    `db:"id"`
	WalletID    string    `db:"wallet_id"`
	Type        string    `db:"type"` // deposit, withdraw, transfer_in, transfer_out
	Amount      float64   `db:"amount"`
	ReferenceID *string   `db:"reference_id"`
	Description *string   `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
}
