package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/models"
	"github.com/itisalisas/avito-backend/internal/service/reception"
	"github.com/itisalisas/avito-backend/internal/utils"
	"github.com/itisalisas/avito-backend/pkg/metrics"
)

type ReceptionHandler struct {
	receptionService reception.ServiceInterface
}

func NewReceptionHandler(receptionService reception.ServiceInterface) *ReceptionHandler {
	return &ReceptionHandler{receptionService: receptionService}
}

func (h *ReceptionHandler) AddReception(w http.ResponseWriter, r *http.Request) {
	var request dto.PostReceptionsJSONRequestBody

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.WriteResponse(w, utils.Error("Invalid request: "+err.Error()), http.StatusBadRequest)
		return
	}

	addedReception, err := h.receptionService.AddReception(r.Context(), request)

	switch {
	case errors.Is(err, models.ErrReceptionNotClosed):
		utils.WriteResponse(w, utils.Error(err.Error()), http.StatusBadRequest)
	case err != nil:
		utils.WriteResponse(w, utils.Error(err.Error()), http.StatusInternalServerError)
	default:
		utils.WriteResponse(w, addedReception, http.StatusCreated)
		metrics.OrderReceptionsCreated.Inc()
	}
}

func (h *ReceptionHandler) CloseLastReception(w http.ResponseWriter, r *http.Request) {
	pvzIdStr := r.PathValue("pvzId")
	pvzId, err := uuid.Parse(pvzIdStr)
	if err != nil {
		utils.WriteResponse(w, utils.Error("Invalid request: "+err.Error()), http.StatusBadRequest)
	}

	closedReception, err := h.receptionService.CloseLastReception(r.Context(), pvzId)
	switch {
	case err != nil:
		utils.WriteResponse(w, utils.Error(err.Error()), http.StatusInternalServerError)
	default:
		utils.WriteResponse(w, closedReception, http.StatusOK)
	}
}
