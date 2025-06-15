package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestUserModel(t *testing.T) {
	t.Run("User creation with valid data", func(t *testing.T) {
		user := User{
			ID:        uuid.New(),
			Name:      "John Doe",
			CreatedAt: time.Now().UTC(),
		}

		assert.NotEqual(t, uuid.Nil, user.ID)
		assert.Equal(t, "John Doe", user.Name)
		assert.False(t, user.CreatedAt.IsZero())
	})

	t.Run("UserWithWallet contains wallet data", func(t *testing.T) {
		userID := uuid.New()
		walletID := uuid.New()
		now := time.Now().UTC()

		userWithWallet := UserWithWallet{
			ID:   userID,
			Name: "Jane Doe",
			Wallet: Wallet{
				ID:        walletID,
				UserID:    userID,
				Balance:   decimal.NewFromFloat(100.50),
				CreatedAt: now,
			},
			CreatedAt: now,
		}

		assert.Equal(t, userID, userWithWallet.ID)
		assert.Equal(t, "Jane Doe", userWithWallet.Name)
		assert.Equal(t, walletID, userWithWallet.Wallet.ID)
		assert.Equal(t, userID, userWithWallet.Wallet.UserID)
		assert.True(t, userWithWallet.Wallet.Balance.Equal(decimal.NewFromFloat(100.50)))
	})
}

func TestWalletModel(t *testing.T) {
	t.Run("Wallet creation with zero balance", func(t *testing.T) {
		userID := uuid.New()
		wallet := Wallet{
			ID:        uuid.New(),
			UserID:    userID,
			Balance:   decimal.Zero,
			CreatedAt: time.Now().UTC(),
		}

		assert.NotEqual(t, uuid.Nil, wallet.ID)
		assert.Equal(t, userID, wallet.UserID)
		assert.True(t, wallet.Balance.IsZero())
		assert.False(t, wallet.CreatedAt.IsZero())
	})

	t.Run("Wallet balance precision handling", func(t *testing.T) {
		wallet := Wallet{
			ID:      uuid.New(),
			UserID:  uuid.New(),
			Balance: decimal.NewFromFloat(123.456789),
		}

		// Decimal should maintain precision
		assert.True(t, wallet.Balance.Equal(decimal.NewFromFloat(123.456789)))

		// Test rounding to 2 decimal places (if needed in business logic)
		rounded := wallet.Balance.Round(2)
		expected := decimal.NewFromFloat(123.46)
		assert.True(t, rounded.Equal(expected))
	})

	t.Run("Wallet balance operations", func(t *testing.T) {
		wallet := Wallet{
			ID:      uuid.New(),
			UserID:  uuid.New(),
			Balance: decimal.NewFromFloat(100.00),
		}

		// Test addition
		depositAmount := decimal.NewFromFloat(50.25)
		newBalance := wallet.Balance.Add(depositAmount)
		expected := decimal.NewFromFloat(150.25)
		assert.True(t, newBalance.Equal(expected))

		// Test subtraction
		withdrawAmount := decimal.NewFromFloat(25.50)
		newBalance = wallet.Balance.Sub(withdrawAmount)
		expected = decimal.NewFromFloat(74.50)
		assert.True(t, newBalance.Equal(expected))

		// Test negative balance detection
		largeWithdraw := decimal.NewFromFloat(200.00)
		newBalance = wallet.Balance.Sub(largeWithdraw)
		assert.True(t, newBalance.IsNegative())
	})
}

func TestTransactionModel(t *testing.T) {
	t.Run("Transaction creation with all fields", func(t *testing.T) {
		walletID := uuid.New()
		description := "Test deposit"
		transaction := Transaction{
			ID:          uuid.New(),
			WalletID:    walletID,
			Type:        "deposit",
			Amount:      decimal.NewFromFloat(100.50),
			ReferenceID: nil,
			Description: &description,
			CreatedAt:   time.Now().UTC(),
		}

		assert.NotEqual(t, uuid.Nil, transaction.ID)
		assert.Equal(t, walletID, transaction.WalletID)
		assert.Equal(t, "deposit", transaction.Type)
		assert.True(t, transaction.Amount.Equal(decimal.NewFromFloat(100.50)))
		assert.Nil(t, transaction.ReferenceID)
		assert.Equal(t, "Test deposit", *transaction.Description)
		assert.False(t, transaction.CreatedAt.IsZero())
	})

	t.Run("Transfer transaction with ReferenceID", func(t *testing.T) {
		fromWalletID := uuid.New()
		referenceID := uuid.New()
		description := "Transfer to friend"

		transaction := Transaction{
			ID:          uuid.New(),
			WalletID:    fromWalletID,
			Type:        "transfer_out",
			Amount:      decimal.NewFromFloat(75.25),
			ReferenceID: &referenceID,
			Description: &description,
			CreatedAt:   time.Now().UTC(),
		}

		assert.Equal(t, "transfer_out", transaction.Type)
		assert.NotNil(t, transaction.ReferenceID)
		assert.Equal(t, referenceID, *transaction.ReferenceID)
	})

	t.Run("Transaction type validation", func(t *testing.T) {
		validTypes := []string{"deposit", "withdraw", "transfer_in", "transfer_out"}

		for _, transactionType := range validTypes {
			transaction := Transaction{
				ID:       uuid.New(),
				WalletID: uuid.New(),
				Type:     transactionType,
				Amount:   decimal.NewFromFloat(10.00),
			}

			assert.Contains(t, validTypes, transaction.Type)
		}
	})

	t.Run("Transaction amount precision", func(t *testing.T) {
		transaction := Transaction{
			ID:       uuid.New(),
			WalletID: uuid.New(),
			Type:     "deposit",
			Amount:   decimal.NewFromFloat(99.999),
		}

		// Amount should maintain precision
		assert.True(t, transaction.Amount.Equal(decimal.NewFromFloat(99.999)))

		// Test business logic rounding (if applied)
		rounded := transaction.Amount.Round(2)
		expected := decimal.NewFromFloat(100.00)
		assert.True(t, rounded.Equal(expected))
	})
}

func TestDecimalHandling(t *testing.T) {
	t.Run("Zero decimal values", func(t *testing.T) {
		zero := decimal.Zero
		assert.True(t, zero.IsZero())
		assert.False(t, zero.IsPositive())
		assert.False(t, zero.IsNegative())
	})

	t.Run("Positive decimal values", func(t *testing.T) {
		positive := decimal.NewFromFloat(10.50)
		assert.False(t, positive.IsZero())
		assert.True(t, positive.IsPositive())
		assert.False(t, positive.IsNegative())
	})

	t.Run("Negative decimal values", func(t *testing.T) {
		negative := decimal.NewFromFloat(-5.25)
		assert.False(t, negative.IsZero())
		assert.False(t, negative.IsPositive())
		assert.True(t, negative.IsNegative())
	})

	t.Run("Decimal comparison", func(t *testing.T) {
		amount1 := decimal.NewFromFloat(100.00)
		amount2 := decimal.NewFromFloat(100.00)
		amount3 := decimal.NewFromFloat(99.99)

		assert.True(t, amount1.Equal(amount2))
		assert.False(t, amount1.Equal(amount3))
		assert.True(t, amount1.GreaterThan(amount3))
		assert.True(t, amount3.LessThan(amount1))
	})

	t.Run("Decimal arithmetic precision", func(t *testing.T) {
		// Test precision issues that might occur with float64
		amount1 := decimal.NewFromFloat(0.1)
		amount2 := decimal.NewFromFloat(0.2)
		sum := amount1.Add(amount2)
		expected := decimal.NewFromFloat(0.3)

		// This should be true with decimal, but might fail with float64
		assert.True(t, sum.Equal(expected))

		// Verify string representation is precise
		assert.Equal(t, "0.3", sum.String())
	})

	t.Run("Large number handling", func(t *testing.T) {
		large := decimal.NewFromFloat(999999999.99)
		addition := decimal.NewFromFloat(0.01)
		result := large.Add(addition)
		expected := decimal.NewFromFloat(1000000000.00)

		assert.True(t, result.Equal(expected))
	})
}
