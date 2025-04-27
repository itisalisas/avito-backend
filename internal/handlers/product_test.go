package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/models"
)

type stubProductService struct {
	AddProductFunc        func(ctx context.Context, req dto.PostProductsJSONRequestBody) (*dto.Product, error)
	DeleteLastProductFunc func(ctx context.Context, pvzId uuid.UUID) error
}

func (s *stubProductService) AddProduct(ctx context.Context, req dto.PostProductsJSONRequestBody) (*dto.Product, error) {
	return s.AddProductFunc(ctx, req)
}

func (s *stubProductService) DeleteLastProduct(ctx context.Context, pvzId uuid.UUID) error {
	return s.DeleteLastProductFunc(ctx, pvzId)
}

func TestProductHandler_AddProduct(t *testing.T) {
	invalidJSON := []byte(`qwerty`)

	tests := []struct {
		name           string
		body           []byte
		serviceReturn  *dto.Product
		serviceErr     error
		wantStatus     int
		wantBodySubstr string
	}{
		{
			name:           "invalid JSON",
			body:           invalidJSON,
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: "Invalid request",
		},
		{
			name:           "incorrect product type",
			body:           []byte(`{"type":"wrong","pvzId":"00000000-0000-0000-0000-000000000000"}`),
			serviceErr:     models.ErrIncorrectProductType,
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: models.ErrIncorrectProductType.Error(),
		},
		{
			name:           "reception not found",
			body:           []byte(`{"type":"some","pvzId":"00000000-0000-0000-0000-000000000000"}`),
			serviceErr:     models.ErrReceptionNotFound,
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: models.ErrReceptionNotFound.Error(),
		},
		{
			name:           "internal err",
			body:           []byte(`{"type":"some","pvzId":"00000000-0000-0000-0000-000000000000"}`),
			serviceErr:     errors.New("fail"),
			wantStatus:     http.StatusInternalServerError,
			wantBodySubstr: "fail",
		},
		{
			name:           "success",
			body:           []byte(`{"type":"electronics","pvzId":"00000000-0000-0000-0000-000000000000"}`),
			serviceReturn:  &dto.Product{Id: &openapi_types.UUID{}, Type: "electronics"},
			wantStatus:     http.StatusCreated,
			wantBodySubstr: `"type":"electronics"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &stubProductService{
				AddProductFunc: func(ctx context.Context, req dto.PostProductsJSONRequestBody) (*dto.Product, error) {
					return tt.serviceReturn, tt.serviceErr
				},
			}
			h := NewProductHandler(stub)

			req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(tt.body))
			w := httptest.NewRecorder()

			h.AddProduct(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			require.Equal(t, tt.wantStatus, resp.StatusCode)

			respBody, _ := io.ReadAll(resp.Body)
			require.Contains(t, string(respBody), tt.wantBodySubstr)
		})
	}
}

func TestProductHandler_DeleteLastProduct(t *testing.T) {
	validPvzId := uuid.New()
	invalidPvzId := "qwerty"

	tests := []struct {
		name           string
		pvzId          string
		serviceErr     error
		wantStatus     int
		wantBodySubstr string
	}{
		{
			name:           "invalid UUID",
			pvzId:          invalidPvzId,
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: "Invalid request",
		},
		{
			name:           "internal err",
			pvzId:          validPvzId.String(),
			serviceErr:     errors.New("delete error"),
			wantStatus:     http.StatusInternalServerError,
			wantBodySubstr: "delete error",
		},
		{
			name:           "success",
			pvzId:          validPvzId.String(),
			wantStatus:     http.StatusOK,
			wantBodySubstr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &stubProductService{
				DeleteLastProductFunc: func(ctx context.Context, pvzId uuid.UUID) error {
					return tt.serviceErr
				},
			}
			h := NewProductHandler(stub)

			req := httptest.NewRequest(http.MethodDelete, "/products/"+tt.pvzId, nil)

			req.SetPathValue("pvzId", tt.pvzId)

			w := httptest.NewRecorder()

			h.DeleteLastProduct(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			require.Equal(t, tt.wantStatus, resp.StatusCode)

			respBody, _ := io.ReadAll(resp.Body)
			if tt.wantBodySubstr != "" {
				require.Contains(t, string(respBody), tt.wantBodySubstr)
			} else {
				require.Empty(t, respBody)
			}
		})
	}
}
