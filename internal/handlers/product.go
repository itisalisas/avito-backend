package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/models"
	"github.com/itisalisas/avito-backend/internal/service/product"
	"github.com/itisalisas/avito-backend/internal/utils"
	"github.com/itisalisas/avito-backend/pkg/metrics"
)

type ProductHandler struct {
	productService product.ServiceInterface
}

func NewProductHandler(productService product.ServiceInterface) *ProductHandler {
	return &ProductHandler{productService: productService}
}

func (h *ProductHandler) AddProduct(w http.ResponseWriter, r *http.Request) {
	var request dto.PostProductsJSONRequestBody

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.WriteResponse(w, utils.Error("Invalid request: "+err.Error()), http.StatusBadRequest)
		return
	}

	addedProduct, err := h.productService.AddProduct(r.Context(), request)

	switch {
	case errors.Is(err, models.ErrIncorrectProductType) || errors.Is(err, models.ErrReceptionNotFound):
		utils.WriteResponse(w, utils.Error(err.Error()), http.StatusBadRequest)
	case err != nil:
		utils.WriteResponse(w, utils.Error(err.Error()), http.StatusInternalServerError)
	default:
		utils.WriteResponse(w, addedProduct, http.StatusCreated)
		metrics.ProductsAdded.Inc()
	}
}

func (h *ProductHandler) DeleteLastProduct(w http.ResponseWriter, r *http.Request) {
	pvzIdStr := r.PathValue("pvzId")
	pvzId, err := uuid.Parse(pvzIdStr)
	if err != nil {
		utils.WriteResponse(w, utils.Error("Invalid request: "+err.Error()), http.StatusBadRequest)
		return
	}

	err = h.productService.DeleteLastProduct(r.Context(), pvzId)
	switch {
	case err != nil:
		utils.WriteResponse(w, utils.Error(err.Error()), http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusOK)
	}
}
