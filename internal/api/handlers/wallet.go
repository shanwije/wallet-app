package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shanwije/wallet-app/internal/service"
	"github.com/shopspring/decimal"
)

type WalletHandler struct {
	WalletService *service.WalletService
}

type depositRequest struct {
	Amount float64 `json:"amount"`
}

type withdrawRequest struct {
	Amount float64 `json:"amount"`
}

// Deposit adds money to a wallet
// @Summary Deposit to wallet
// @Tags wallets
// @Accept json
// @Produce json
// @Param id path string true "Wallet ID"
// @Param deposit body depositRequest true "Deposit details"
// @Success 200 {object} models.Wallet
// @Router /api/v1/wallets/{id}/deposit [post]
func (h *WalletHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	walletIDStr := chi.URLParam(r, "id")
	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		http.Error(w, "Invalid wallet ID", http.StatusBadRequest)
		return
	}

	var req depositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Convert float64 to decimal for precise calculations
	amount := decimal.NewFromFloat(req.Amount)

	wallet, err := h.WalletService.Deposit(walletID, amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wallet)
}

// Withdraw removes money from a wallet
// @Summary Withdraw from wallet
// @Tags wallets
// @Accept json
// @Produce json
// @Param id path string true "Wallet ID"
// @Param withdraw body withdrawRequest true "Withdraw details"
// @Success 200 {object} models.Wallet
// @Router /api/v1/wallets/{id}/withdraw [post]
func (h *WalletHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	walletIDStr := chi.URLParam(r, "id")
	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		http.Error(w, "Invalid wallet ID", http.StatusBadRequest)
		return
	}

	var req withdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Convert float64 to decimal for precise calculations
	amount := decimal.NewFromFloat(req.Amount)

	wallet, err := h.WalletService.Withdraw(walletID, amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wallet)
}

// GetBalance gets wallet balance
// @Summary Get wallet balance
// @Tags wallets
// @Produce json
// @Param id path string true "Wallet ID"
// @Success 200 {object} models.Wallet
// @Router /api/v1/wallets/{id}/balance [get]
func (h *WalletHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	walletIDStr := chi.URLParam(r, "id")
	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		http.Error(w, "Invalid wallet ID", http.StatusBadRequest)
		return
	}

	wallet, err := h.WalletService.GetBalance(walletID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wallet)
}
