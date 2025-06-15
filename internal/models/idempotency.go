package models

import (
	"github.com/google/uuid"
	"time"
)

type IdempotencyKey struct {
	Key        string    `db:"key" json:"key"`
	WalletID   uuid.UUID `db:"wallet_id" json:"wallet_id"`
	Response   []byte    `db:"response_body" json:"response_body"`
	StatusCode int       `db:"status_code" json:"status_code"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}
