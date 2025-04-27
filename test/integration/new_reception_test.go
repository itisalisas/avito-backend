package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/handlers"
	"github.com/itisalisas/avito-backend/internal/service/product"
	"github.com/itisalisas/avito-backend/internal/service/pvz"
	"github.com/itisalisas/avito-backend/internal/service/reception"
	"github.com/itisalisas/avito-backend/internal/storage"
)

func setupTestRouter() *chi.Mux {
	db := storage.DBTestSetup()

	pvzRepo := storage.NewPvzRepository(db)
	receptionRepo := storage.NewReceptionRepository(db)
	productRepo := storage.NewProductRepository(db)

	pvzService := pvz.NewPvzService(pvzRepo)
	receptionService := reception.NewReceptionService(receptionRepo)
	productService := product.NewProductService(productRepo, receptionRepo)

	pvzHandler := handlers.NewPvzHandler(pvzService)
	receptionHandler := handlers.NewReceptionHandler(receptionService)
	productHandler := handlers.NewProductHandler(productService)

	r := chi.NewRouter()

	r.HandleFunc("POST /pvz", pvzHandler.AddPvz)
	r.HandleFunc("/receptions", receptionHandler.AddReception)
	r.HandleFunc("/pvz/{pvzId}/close_last_reception", receptionHandler.CloseLastReception)
	r.HandleFunc("/products", productHandler.AddProduct)

	return r
}

func TestCreatePvzAndReception(t *testing.T) {
	router := setupTestRouter()

	ts := httptest.NewServer(router)
	defer ts.Close()

	pvzData := map[string]interface{}{
		"city": dto.Москва,
	}
	pvzJSON, err := json.Marshal(pvzData)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Post(ts.URL+"/pvz", "application/json", bytes.NewBuffer(pvzJSON))
	if err != nil {
		t.Fatal(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		require.NoError(t, err)
	}(resp.Body)

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected status code to be 201")
	var pvzInfo dto.PVZ
	if json.NewDecoder(resp.Body).Decode(&pvzInfo) != nil {
		t.Fatal(err)
	}

	receptionData := map[string]interface{}{
		"pvzId": pvzInfo.Id,
	}
	receptionJSON, err := json.Marshal(receptionData)
	if err != nil {
		t.Fatal(err)
	}
	resp, err = http.Post(ts.URL+"/receptions", "application/json", bytes.NewBuffer(receptionJSON))
	if err != nil {
		t.Fatal(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		require.NoError(t, err)
	}(resp.Body)

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected status code to be 201")
	var receptionInfo dto.Reception
	if json.NewDecoder(resp.Body).Decode(&receptionInfo) != nil {
		t.Fatal(err)
	}

	for i := 0; i < 50; i++ {
		productData := map[string]interface{}{
			"pvzId": pvzInfo.Id,
			"type":  dto.ProductTypeОдежда,
		}
		productJSON, err := json.Marshal(productData)
		if err != nil {
			t.Fatal(err)
		}
		resp, err = http.Post(ts.URL+"/products", "application/json", bytes.NewBuffer(productJSON))
		if err != nil {
			t.Fatal(err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			require.NoError(t, err)
		}(resp.Body)

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected status code to be 201 for product")
	}

	url := fmt.Sprintf("%s/pvz/%s/close_last_reception", ts.URL, pvzInfo.Id)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		t.Fatal(err)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		require.NoError(t, err)
	}(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code to be 200 for closing reception")
}
