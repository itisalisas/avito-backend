package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/models"
)

type stubPvzService struct {
	GetPvzListFunc func(ctx context.Context, startTime, endTime *time.Time, page, limit uint64) ([]*models.ExtendedPvz, error)
	AddPvzFunc     func(ctx context.Context, pvz *dto.PostPvzJSONRequestBody) (*dto.PVZ, error)
}

func (s *stubPvzService) GetPvzList(ctx context.Context, startTime, endTime *time.Time, page, limit uint64) ([]*models.ExtendedPvz, error) {
	return s.GetPvzListFunc(ctx, startTime, endTime, page, limit)
}

func (s *stubPvzService) AddPvz(ctx context.Context, pvz *dto.PostPvzJSONRequestBody) (*dto.PVZ, error) {
	return s.AddPvzFunc(ctx, pvz)
}

func TestPvzHandler_GetPvz(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		serviceReturn  []*models.ExtendedPvz
		serviceErr     error
		wantStatus     int
		wantBodySubstr string
	}{
		{
			name:           "invalid startDate format",
			queryParams:    "?startDate=invalid",
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: "invalid startDate format",
		},
		{
			name:           "startDate after endDate",
			queryParams:    "?startDate=2025-04-10T00:00:00Z&endDate=2025-04-01T00:00:00Z",
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: "startDate must be before endDate",
		},
		{
			name:           "invalid page format",
			queryParams:    "?page=zero",
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: "invalid page format",
		},
		{
			name:           "invalid limit format",
			queryParams:    "?limit=1000",
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: "invalid limit format",
		},
		{
			name:           "internal err",
			queryParams:    "",
			serviceErr:     errors.New("db error"),
			wantStatus:     http.StatusInternalServerError,
			wantBodySubstr: "db error",
		},
		{
			name:           "success",
			queryParams:    "?page=1&limit=2",
			serviceReturn:  []*models.ExtendedPvz{{PVZ: dto.PVZ{City: dto.СанктПетербург}}},
			wantStatus:     http.StatusOK,
			wantBodySubstr: `"city":"Санкт-Петербург"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &stubPvzService{
				GetPvzListFunc: func(ctx context.Context, startTime, endTime *time.Time, page, limit uint64) ([]*models.ExtendedPvz, error) {
					return tt.serviceReturn, tt.serviceErr
				},
			}
			h := NewPvzHandler(stub)

			req := httptest.NewRequest(http.MethodGet, "/pvz"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			h.GetPvz(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			require.Equal(t, tt.wantStatus, resp.StatusCode)

			respBody, _ := io.ReadAll(resp.Body)
			require.Contains(t, string(respBody), tt.wantBodySubstr)
		})
	}
}

func TestPvzHandler_AddPvz(t *testing.T) {
	invalidJSON := []byte(`qwerty`)

	tests := []struct {
		name           string
		body           []byte
		serviceErr     error
		serviceReturn  *dto.PVZ
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
			name:           "incorrect city",
			body:           []byte(`{"city":"Москва"}`),
			serviceErr:     models.ErrIncorrectCity,
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: models.ErrIncorrectCity.Error(),
		},
		{
			name:           "internal err",
			body:           []byte(`{"city":"Москва"}`),
			serviceErr:     errors.New("insert failed"),
			wantStatus:     http.StatusInternalServerError,
			wantBodySubstr: "insert failed",
		},
		{
			name:       "success",
			body:       []byte(`{"city":"Москва"}`),
			wantStatus: http.StatusCreated,
			serviceReturn: &dto.PVZ{
				City: dto.Москва,
			},
			wantBodySubstr: `"city":"Москва"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &stubPvzService{
				AddPvzFunc: func(ctx context.Context, pvz *dto.PostPvzJSONRequestBody) (*dto.PVZ, error) {
					return tt.serviceReturn, tt.serviceErr
				},
			}
			h := NewPvzHandler(stub)

			req := httptest.NewRequest(http.MethodPost, "/pvz", bytes.NewReader(tt.body))
			w := httptest.NewRecorder()

			h.AddPvz(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			require.Equal(t, tt.wantStatus, resp.StatusCode)

			respBody, _ := io.ReadAll(resp.Body)
			require.Contains(t, string(respBody), tt.wantBodySubstr)
		})
	}
}
