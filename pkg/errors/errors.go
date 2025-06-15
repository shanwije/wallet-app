package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Error codes for the application
const (
	// Validation errors
	ErrInvalidInput  = "INVALID_INPUT"
	ErrMissingField  = "MISSING_FIELD"
	ErrInvalidUUID   = "INVALID_UUID"
	ErrInvalidAmount = "INVALID_AMOUNT"

	// Business logic errors
	ErrInsufficientFunds  = "INSUFFICIENT_FUNDS"
	ErrWalletNotFound     = "WALLET_NOT_FOUND"
	ErrUserNotFound       = "USER_NOT_FOUND"
	ErrSameWalletTransfer = "SAME_WALLET_TRANSFER"

	// System errors
	ErrDatabaseConnection = "DATABASE_CONNECTION"
	ErrTransactionFailed  = "TRANSACTION_FAILED"
	ErrInternal           = "INTERNAL_ERROR"
)

// AppError represents an application error with context
type AppError struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Details    map[string]string `json:"details,omitempty"`
	HTTPStatus int               `json:"-"`
	Err        error             `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new application error
func New(code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Wrap creates a new application error wrapping an existing error
func Wrap(err error, code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
	}
}

// WithDetails adds additional context to an error
func (e *AppError) WithDetails(key, value string) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]string)
	}
	e.Details[key] = value
	return e
}

// Common error constructors
func InvalidInput(message string) *AppError {
	return New(ErrInvalidInput, message, http.StatusBadRequest)
}

func InsufficientFunds() *AppError {
	return New(ErrInsufficientFunds, "Insufficient funds for this operation", http.StatusBadRequest)
}

func WalletNotFound(walletID string) *AppError {
	return New(ErrWalletNotFound, "Wallet not found", http.StatusNotFound).
		WithDetails("wallet_id", walletID)
}

func UserNotFound(userID string) *AppError {
	return New(ErrUserNotFound, "User not found", http.StatusNotFound).
		WithDetails("user_id", userID)
}

func DatabaseError(err error) *AppError {
	return Wrap(err, ErrDatabaseConnection, "Database operation failed", http.StatusInternalServerError)
}

func InternalError(err error) *AppError {
	return Wrap(err, ErrInternal, "Internal server error", http.StatusInternalServerError)
}

// HTTP response utilities

// ErrorResponse represents a JSON error response
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

// RespondWithError sends a JSON error response
func RespondWithError(w http.ResponseWriter, httpStatus int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	response := ErrorResponse{
		Error: message,
	}

	json.NewEncoder(w).Encode(response)
}

// RespondWithAppError sends a JSON error response using an AppError
func RespondWithAppError(w http.ResponseWriter, appErr *AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.HTTPStatus)

	response := ErrorResponse{
		Error: appErr.Message,
		Code:  appErr.Code,
	}

	json.NewEncoder(w).Encode(response)
}
