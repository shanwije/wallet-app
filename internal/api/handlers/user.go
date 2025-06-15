package handlers

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/shanwije/wallet-app/internal/service"
	"github.com/shanwije/wallet-app/pkg/errors"
	"github.com/shanwije/wallet-app/pkg/logger"
)

type UserHandler struct {
	UserService *service.UserService
}

type createUserRequest struct {
	Name string `json:"name"`
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		UserService: userService,
	}
}

// CreateUser creates a user with wallet
// @Summary Create user
// @Tags users
// @Accept json
// @Produce json
// @Param user body createUserRequest true "User details"
// @Success 201 {object} models.UserWithWallet
// @Router /api/v1/users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("Failed to decode request", zap.Error(err))
		errors.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if req.Name == "" {
		log.Warn("User creation failed: empty name provided")
		errors.RespondWithError(w, http.StatusBadRequest, "Name is required")
		return
	}

	user, err := h.UserService.CreateUser(r.Context(), req.Name)
	if err != nil {
		log.Error("Failed to create user", zap.Error(err), zap.String("name", req.Name))
		errors.RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	log.Info("User created successfully", zap.String("user_id", user.ID.String()), zap.String("name", user.Name))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
