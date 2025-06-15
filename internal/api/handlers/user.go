package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/shanwije/wallet-app/internal/service"
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
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	user, err := h.UserService.CreateUser(req.Name)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
