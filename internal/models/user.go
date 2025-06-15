package models

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type UserWithWallet struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Wallet    Wallet    `json:"wallet"`
	CreatedAt time.Time `json:"created_at"`
}
