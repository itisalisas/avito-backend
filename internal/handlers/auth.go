package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/models"
	"github.com/itisalisas/avito-backend/internal/service"
	"github.com/itisalisas/avito-backend/internal/utils"
	"net/http"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{
		authService: service.NewAuthService(db),
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var request dto.PostRegisterJSONRequestBody

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.WriteResponse(w, utils.Error("Invalid request: "+err.Error()), http.StatusBadRequest)
		return
	}

	user, err := h.authService.Register(r.Context(), request)

	switch {
	case errors.Is(err, models.ErrIncorrectUserRole) || errors.Is(err, models.ErrEmailAlreadyInUse):
		utils.WriteResponse(w, utils.Error(err.Error()), http.StatusBadRequest)
	case err != nil:
		utils.WriteResponse(w, utils.Error(err.Error()), http.StatusInternalServerError)
	default:
		utils.WriteResponse(w, user, http.StatusCreated)
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request dto.PostLoginJSONRequestBody

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.WriteResponse(w, utils.Error("Invalid request: "+err.Error()), http.StatusBadRequest)
		return
	}

	token, err := h.authService.Login(r.Context(), request)
	switch {
	case errors.Is(err, models.ErrUserNotFound) || errors.Is(err, models.ErrWrongPassword):
		utils.WriteResponse(w, utils.Error(err.Error()), http.StatusUnauthorized)
	case err != nil:
		utils.WriteResponse(w, utils.Error(err.Error()), http.StatusInternalServerError)
	default:
		utils.WriteResponse(w, token, http.StatusOK)
	}
}

func (h *AuthHandler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	var request dto.PostDummyLoginJSONRequestBody

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.WriteResponse(w, utils.Error("Invalid request: "+err.Error()), http.StatusBadRequest)
		return
	}

	token, err := h.authService.DummyLogin(request)

	switch {
	case errors.Is(err, models.ErrIncorrectUserRole):
		utils.WriteResponse(w, utils.Error(err.Error()), http.StatusBadRequest)
	case err != nil:
		utils.WriteResponse(w, utils.Error(err.Error()), http.StatusInternalServerError)
	default:
		utils.WriteResponse(w, token, http.StatusOK)
	}
}
