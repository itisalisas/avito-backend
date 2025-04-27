package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/models"
	"github.com/itisalisas/avito-backend/internal/service/auth"
	"github.com/itisalisas/avito-backend/internal/utils"
)

type AuthHandler struct {
	authService auth.ServiceInterface
}

func NewAuthHandler(authService auth.ServiceInterface) *AuthHandler {
	return &AuthHandler{
		authService: authService,
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
	case errors.Is(err, models.ErrIncorrectUserRole) || errors.Is(err, models.ErrEmailAlreadyInUse) || errors.Is(err, models.ErrEmptyEmailOrPassword):
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
	case errors.Is(err, models.ErrUserNotFound) || errors.Is(err, models.ErrWrongPassword) || errors.Is(err, models.ErrEmptyEmailOrPassword):
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
