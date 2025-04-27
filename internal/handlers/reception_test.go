package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/models"
)

type stubReceptionService struct {
	AddReceptionFunc       func(ctx context.Context, request dto.PostReceptionsJSONRequestBody) (*dto.Reception, error)
	CloseLastReceptionFunc func(ctx context.Context, pvzID uuid.UUID) (*dto.Reception, error)
}

func (s *stubReceptionService) AddReception(ctx context.Context, request dto.PostReceptionsJSONRequestBody) (*dto.Reception, error) {
	return s.AddReceptionFunc(ctx, request)
}

func (s *stubReceptionService) CloseLastReception(ctx context.Context, pvzID uuid.UUID) (*dto.Reception, error) {
	return s.CloseLastReceptionFunc(ctx, pvzID)
}

func TestReceptionHandler_AddReception(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		serviceReturn  *dto.Reception
		serviceErr     error
		wantStatus     int
		wantBodySubstr string
	}{
		{
			name:           "invalid JSON",
			requestBody:    "invalid_json",
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: "Invalid request",
		},
		{
			name:           "reception not closed error",
			requestBody:    dto.PostReceptionsJSONRequestBody{PvzId: uuid.New()},
			serviceErr:     models.ErrReceptionNotClosed,
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: models.ErrReceptionNotClosed.Error(),
		},
		{
			name:           "internal err",
			requestBody:    dto.PostReceptionsJSONRequestBody{PvzId: uuid.New()},
			serviceErr:     errors.New("db error"),
			wantStatus:     http.StatusInternalServerError,
			wantBodySubstr: "db error",
		},
		{
			name:           "success",
			requestBody:    dto.PostReceptionsJSONRequestBody{PvzId: uuid.New()},
			serviceReturn:  &dto.Reception{PvzId: uuid.New()},
			wantStatus:     http.StatusCreated,
			wantBodySubstr: `"pvzId`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &stubReceptionService{
				AddReceptionFunc: func(ctx context.Context, req dto.PostReceptionsJSONRequestBody) (*dto.Reception, error) {
					return tt.serviceReturn, tt.serviceErr
				},
			}
			h := NewReceptionHandler(stub)

			var bodyBytes []byte
			if str, ok := tt.requestBody.(string); ok {
				bodyBytes = []byte(str)
			} else {
				bodyBytes, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/receptions", bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			h.AddReception(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			respBody, _ := io.ReadAll(resp.Body)
			require.Equal(t, tt.wantStatus, resp.StatusCode)
			require.Contains(t, string(respBody), tt.wantBodySubstr)
		})
	}
}

func TestReceptionHandler_CloseLastReception(t *testing.T) {
	tests := []struct {
		name           string
		pvzID          string
		serviceReturn  dto.Reception
		serviceErr     error
		wantStatus     int
		wantBodySubstr string
	}{
		{
			name:           "invalid UUID",
			pvzID:          "invalid-uuid",
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: "Invalid request",
		},
		{
			name:           "internal err",
			pvzID:          uuid.New().String(),
			serviceErr:     errors.New("db error"),
			wantStatus:     http.StatusInternalServerError,
			wantBodySubstr: "db error",
		},
		{
			name:           "success",
			pvzID:          uuid.New().String(),
			serviceReturn:  dto.Reception{Status: dto.Close},
			wantStatus:     http.StatusOK,
			wantBodySubstr: `"status":"close"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &stubReceptionService{
				CloseLastReceptionFunc: func(ctx context.Context, pvzID uuid.UUID) (*dto.Reception, error) {
					return &tt.serviceReturn, tt.serviceErr
				},
			}
			h := NewReceptionHandler(stub)

			req := httptest.NewRequest(http.MethodPatch, "/receptions/"+tt.pvzID+"/close", nil)
			req.SetPathValue("pvzId", tt.pvzID)

			w := httptest.NewRecorder()

			h.CloseLastReception(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			respBody, _ := io.ReadAll(resp.Body)
			require.Equal(t, tt.wantStatus, resp.StatusCode)
			require.Contains(t, string(respBody), tt.wantBodySubstr)
		})
	}
}
