package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/models"
	"github.com/itisalisas/avito-backend/internal/service"
	"github.com/itisalisas/avito-backend/internal/utils"
	"github.com/itisalisas/avito-backend/pkg/metrics"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type PvzHandler struct {
	pvzService *service.PvzService
}

func NewPvzHandler(db *sql.DB) *PvzHandler {
	return &PvzHandler{pvzService: service.NewPvzService(db)}
}

func (h *PvzHandler) GetPvz(w http.ResponseWriter, r *http.Request) {
	params, err := parseGetPvzParams(r.URL.Query())
	if err != nil {
		utils.WriteResponse(w, utils.Error(err.Error()), http.StatusBadRequest)
		return
	}

	page := uint64(*params.Page)
	limit := uint64(*params.Limit)

	pvzList, err := h.pvzService.GetPvzList(r.Context(), params.StartDate, params.EndDate, page, limit)
	switch {
	case err != nil:
		utils.WriteResponse(w, utils.Error(err.Error()), http.StatusInternalServerError)
	default:
		utils.WriteResponse(w, pvzList, http.StatusOK)
	}
}

func (h *PvzHandler) AddPvz(w http.ResponseWriter, r *http.Request) {
	var pvz dto.PostPvzJSONRequestBody

	if err := json.NewDecoder(r.Body).Decode(&pvz); err != nil {
		utils.WriteResponse(w, utils.Error("Invalid request: "+err.Error()), http.StatusBadRequest)
		return
	}

	err := h.pvzService.AddPvz(r.Context(), &pvz)

	switch {
	case errors.Is(err, models.ErrIncorrectCity):
		utils.WriteResponse(w, utils.Error(err.Error()), http.StatusBadRequest)
	case err != nil:
		utils.WriteResponse(w, utils.Error(err.Error()), http.StatusInternalServerError)
	default:
		utils.WriteResponse(w, pvz, http.StatusCreated)
		metrics.PVZCreated.Inc()
	}
}

func parseGetPvzParams(query url.Values) (*dto.GetPvzParams, error) {
	params := &dto.GetPvzParams{}

	if startDateStr := query.Get("startDate"); startDateStr != "" {
		t, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			return nil, errors.New("invalid startDate format")
		}
		params.StartDate = &t
	}

	if endDateStr := query.Get("endDate"); endDateStr != "" {
		t, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			return nil, errors.New("invalid endDate format")
		}
		params.EndDate = &t
	}

	if params.StartDate != nil && params.EndDate != nil && params.StartDate.After(*params.EndDate) {
		return nil, errors.New("startDate must be before endDate")
	}

	if pageStr := query.Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			return nil, errors.New("invalid page format")
		}
		params.Page = &page
	} else {
		defaultPage := 1
		params.Page = &defaultPage
	}

	if limitStr := query.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 30 {
			return nil, errors.New("invalid limit format")
		}
		params.Limit = &limit
	} else {
		defaultLimit := 10
		params.Limit = &defaultLimit
	}

	return params, nil
}
