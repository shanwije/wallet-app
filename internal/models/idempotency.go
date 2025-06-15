package models

import (
	"time"
)

type IdempotencyKey struct {
	Key         string    `db:"key"`
	WalletID    string    `db:"wallet_id"`
	Response    []byte    `db:"response_body"`
	StatusCode  int       `db:"status_code"`
	CreatedAt   time.Time `db:"created_at"`
}
