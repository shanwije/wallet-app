package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/shanwije/wallet-app/internal/service"
	"github.com/shanwije/wallet-app/pkg/errors"
	"github.com/shanwije/wallet-app/pkg/logger"
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

type transferRequest struct {
	ToWalletID  string  `json:"to_wallet_id"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description,omitempty"`
}

// NewWalletHandler creates a new WalletHandler
func NewWalletHandler(walletService *service.WalletService) *WalletHandler {
	return &WalletHandler{
		WalletService: walletService,
	}
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
	log := logger.FromContext(r.Context())
	ctx := r.Context()
	walletIDStr := chi.URLParam(r, "id")
	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		log.Error("Invalid wallet ID in deposit request", zap.Error(err), zap.String("id", walletIDStr))
		errors.RespondWithError(w, http.StatusBadRequest, "Invalid wallet ID")
		return
	}

	var req depositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("Failed to decode deposit request", zap.Error(err))
		errors.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Convert float64 to decimal for precise calculations
	amount := decimal.NewFromFloat(req.Amount)

	wallet, err := h.WalletService.Deposit(ctx, walletID, amount)
	if err != nil {
		log.Error("Deposit failed", zap.Error(err),
			zap.String("wallet_id", walletID.String()),
			zap.String("amount", amount.String()))
		errors.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Info("Deposit successful",
		zap.String("wallet_id", walletID.String()),
		zap.String("amount", amount.String()),
		zap.String("new_balance", wallet.Balance.String()))

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
	log := logger.FromContext(r.Context())
	ctx := r.Context()
	walletIDStr := chi.URLParam(r, "id")
	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		log.Error("Invalid wallet ID in withdraw request", zap.Error(err), zap.String("id", walletIDStr))
		errors.RespondWithError(w, http.StatusBadRequest, "Invalid wallet ID")
		return
	}

	var req withdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("Failed to decode withdraw request", zap.Error(err))
		errors.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Convert float64 to decimal for precise calculations
	amount := decimal.NewFromFloat(req.Amount)

	wallet, err := h.WalletService.Withdraw(ctx, walletID, amount)
	if err != nil {
		log.Error("Withdraw failed", zap.Error(err),
			zap.String("wallet_id", walletID.String()),
			zap.String("amount", amount.String()))
		errors.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Info("Withdraw successful",
		zap.String("wallet_id", walletID.String()),
		zap.String("amount", amount.String()),
		zap.String("new_balance", wallet.Balance.String()))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wallet)
}

// Transfer moves money from one wallet to another
// @Summary Transfer between wallets
// @Tags wallets
// @Accept json
// @Produce json
// @Param id path string true "Wallet ID"
// @Param transfer body transferRequest true "Transfer details"
// @Success 200 {object} models.Wallet
// @Router /api/v1/wallets/{id}/transfer [post]
func (h *WalletHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fromWalletIDStr := chi.URLParam(r, "id")
	fromWalletID, err := uuid.Parse(fromWalletIDStr)
	if err != nil {
		http.Error(w, "Invalid source wallet ID", http.StatusBadRequest)
		return
	}

	var req transferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	toWalletID, err := uuid.Parse(req.ToWalletID)
	if err != nil {
		http.Error(w, "Invalid destination wallet ID", http.StatusBadRequest)
		return
	}

	// Convert float64 to decimal for precise calculations
	amount := decimal.NewFromFloat(req.Amount)

	err = h.WalletService.Transfer(ctx, fromWalletID, toWalletID, amount, req.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Transfer completed successfully",
	})
}

// GetBalance gets wallet balance
// @Summary Get wallet balance
// @Tags wallets
// @Produce json
// @Param id path string true "Wallet ID"
// @Success 200 {object} models.Wallet
// @Router /api/v1/wallets/{id}/balance [get]
func (h *WalletHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	walletIDStr := chi.URLParam(r, "id")
	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		http.Error(w, "Invalid wallet ID", http.StatusBadRequest)
		return
	}

	wallet, err := h.WalletService.GetBalance(ctx, walletID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wallet)
}

// GetTransactionHistory gets transaction history for a wallet
// @Summary Get wallet transaction history
// @Tags wallets
// @Produce json
// @Param id path string true "Wallet ID"
// @Success 200 {array} models.Transaction
// @Router /api/v1/wallets/{id}/transactions [get]
func (h *WalletHandler) GetTransactionHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	walletIDStr := chi.URLParam(r, "id")
	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		http.Error(w, "Invalid wallet ID", http.StatusBadRequest)
		return
	}

	transactions, err := h.WalletService.GetTransactionHistory(ctx, walletID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}
