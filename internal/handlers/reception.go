package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/models"
	"github.com/itisalisas/avito-backend/internal/service"
	"github.com/itisalisas/avito-backend/internal/utils"
	"github.com/itisalisas/avito-backend/pkg/metrics"
	"net/http"
)

type ReceptionHandler struct {
	receptionService *service.ReceptionService
}

func NewReceptionHandler(db *sql.DB) *ReceptionHandler {
	return &ReceptionHandler{receptionService: service.NewReceptionService(db)}
}

func (h *ReceptionHandler) AddReception(w http.ResponseWriter, r *http.Request) {
	var request dto.PostReceptionsJSONRequestBody

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.WriteResponse(w, utils.Error("Invalid request: "+err.Error()), http.StatusBadRequest)
		return
	}

	reception, err := h.receptionService.AddReception(r.Context(), request)

	switch {
	case errors.Is(err, models.ErrReceptionNotClosed):
		utils.WriteResponse(w, utils.Error(err.Error()), http.StatusBadRequest)
	case err != nil:
		utils.WriteResponse(w, utils.Error(err.Error()), http.StatusInternalServerError)
	default:
		utils.WriteResponse(w, reception, http.StatusCreated)
		metrics.OrderReceptionsCreated.Inc()
	}
}

func (h *ReceptionHandler) CloseLastReception(w http.ResponseWriter, r *http.Request) {
	pvzIdStr := r.PathValue("pvzId")
	pvzId, err := uuid.Parse(pvzIdStr)
	if err != nil {
		utils.WriteResponse(w, utils.Error("Invalid request: "+err.Error()), http.StatusBadRequest)
	}

	reception, err := h.receptionService.CloseLastReception(r.Context(), pvzId)
	switch {
	case err != nil:
		utils.WriteResponse(w, utils.Error(err.Error()), http.StatusInternalServerError)
	default:
		utils.WriteResponse(w, reception, http.StatusOK)
	}
}
