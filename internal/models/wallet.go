package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"time"
)

type Wallet struct {
	ID        uuid.UUID       `db:"id" json:"id"`
	UserID    uuid.UUID       `db:"user_id" json:"user_id"`
	Balance   decimal.Decimal `db:"balance" json:"balance"`
	CreatedAt time.Time       `db:"created_at" json:"created_at"`
}
