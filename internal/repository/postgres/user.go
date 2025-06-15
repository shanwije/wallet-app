package postgres

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shanwije/wallet-app/internal/models"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(name string) (*models.User, error) {
	user := &models.User{
		ID:   uuid.New(),
		Name: name,
	}

	query := `
		INSERT INTO users (id, name) 
		VALUES ($1, $2) 
		RETURNING created_at`

	err := r.db.QueryRow(query, user.ID, user.Name).Scan(&user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, name, created_at FROM users WHERE id = $1`

	err := r.db.Get(user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetUserWithWallet(id uuid.UUID) (*models.UserWithWallet, error) {
	var userWithWallet models.UserWithWallet
	query := `
		SELECT 
			u.id, u.name, u.created_at,
			w.id as wallet_id, w.user_id as wallet_user_id, w.balance, w.created_at as wallet_created_at
		FROM users u
		LEFT JOIN wallets w ON u.id = w.user_id
		WHERE u.id = $1`

	row := r.db.QueryRow(query, id)

	var walletID sql.NullString
	var walletUserID sql.NullString
	var balance sql.NullFloat64
	var walletCreatedAt sql.NullTime

	err := row.Scan(
		&userWithWallet.ID, &userWithWallet.Name, &userWithWallet.CreatedAt,
		&walletID, &walletUserID, &balance, &walletCreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user with wallet: %w", err)
	}

	// If wallet exists, populate it
	if walletID.Valid {
		walletUUID, _ := uuid.Parse(walletID.String)
		userUUID, _ := uuid.Parse(walletUserID.String)
		userWithWallet.Wallet = models.Wallet{
			ID:        walletUUID,
			UserID:    userUUID,
			Balance:   balance.Float64,
			CreatedAt: walletCreatedAt.Time,
		}
	}

	return &userWithWallet, nil
}
