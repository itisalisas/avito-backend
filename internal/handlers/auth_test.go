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

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/models"
)

type stubAuthService struct {
	RegisterFunc   func(ctx context.Context, request dto.PostRegisterJSONRequestBody) (*models.User, error)
	LoginFunc      func(ctx context.Context, request dto.PostLoginJSONRequestBody) (*dto.Token, error)
	DummyLoginFunc func(request dto.PostDummyLoginJSONRequestBody) (*dto.Token, error)
}

func (s *stubAuthService) Register(ctx context.Context, request dto.PostRegisterJSONRequestBody) (*models.User, error) {
	return s.RegisterFunc(ctx, request)
}
func (s *stubAuthService) Login(ctx context.Context, request dto.PostLoginJSONRequestBody) (*dto.Token, error) {
	return s.LoginFunc(ctx, request)
}
func (s *stubAuthService) DummyLogin(request dto.PostDummyLoginJSONRequestBody) (*dto.Token, error) {
	return s.DummyLoginFunc(request)
}

func TestAuthHandler_Register(t *testing.T) {
	invalidJSON := []byte(`{"email":}`)
	tests := []struct {
		name           string
		body           []byte
		serviceReturn  *models.User
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
			name:           "role error -> 400",
			body:           []byte(`{"email":"a@b.c","password":"p","role":"wrong"}`),
			serviceErr:     models.ErrIncorrectUserRole,
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: models.ErrIncorrectUserRole.Error(),
		},
		{
			name:           "email in use -> 400",
			body:           []byte(`{"email":"x@y.z","password":"p","role":"moderator"}`),
			serviceErr:     models.ErrEmailAlreadyInUse,
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: models.ErrEmailAlreadyInUse.Error(),
		},
		{
			name:           "internal error -> 500",
			body:           []byte(`{"email":"u@v.w","password":"p","role":"moderator"}`),
			serviceErr:     errors.New("boom"),
			wantStatus:     http.StatusInternalServerError,
			wantBodySubstr: "boom",
		},
		{
			name:           "success -> 201",
			body:           []byte(`{"email":"u@v.w","password":"p","role":"moderator"}`),
			serviceReturn:  &models.User{ID: openapi_types.UUID{}, Email: "u@v.w", Role: dto.UserRole("moderator")},
			wantStatus:     http.StatusCreated,
			wantBodySubstr: `"email":"u@v.w"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &stubAuthService{
				RegisterFunc: func(ctx context.Context, req dto.PostRegisterJSONRequestBody) (*models.User, error) {
					return tt.serviceReturn, tt.serviceErr
				},
			}
			h := NewAuthHandler(stub)

			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(tt.body))
			w := httptest.NewRecorder()

			h.Register(w, req)
			resp := w.Result()
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				require.NoError(t, err)
			}(resp.Body)

			require.Equal(t, tt.wantStatus, resp.StatusCode)
			var got map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&got)
			b, _ := json.Marshal(got)
			require.Contains(t, string(b), tt.wantBodySubstr)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		body           []byte
		serviceToken   dto.Token
		serviceErr     error
		wantStatus     int
		wantBodySubstr string
	}{
		{
			name:           "invalid JSON",
			body:           []byte(`{"login":}`),
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: "Invalid request",
		},
		{
			name:           "user not found -> 401",
			body:           []byte(`{"email":"a@b.c","password":"p"}`),
			serviceErr:     models.ErrUserNotFound,
			wantStatus:     http.StatusUnauthorized,
			wantBodySubstr: models.ErrUserNotFound.Error(),
		},
		{
			name:           "wrong password -> 401",
			body:           []byte(`{"email":"a@b.c","password":"p"}`),
			serviceErr:     models.ErrWrongPassword,
			wantStatus:     http.StatusUnauthorized,
			wantBodySubstr: models.ErrWrongPassword.Error(),
		},
		{
			name:           "internal error -> 500",
			body:           []byte(`{"email":"a@b.c","password":"p"}`),
			serviceErr:     errors.New("fail"),
			wantStatus:     http.StatusInternalServerError,
			wantBodySubstr: "fail",
		},
		{
			name:           "success -> 200",
			body:           []byte(`{"email":"u@v.w","password":"p"}`),
			serviceToken:   "token123",
			wantStatus:     http.StatusOK,
			wantBodySubstr: `"token123"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &stubAuthService{
				LoginFunc: func(ctx context.Context, req dto.PostLoginJSONRequestBody) (*dto.Token, error) {
					return &tt.serviceToken, tt.serviceErr
				},
			}
			h := NewAuthHandler(stub)

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(tt.body))
			w := httptest.NewRecorder()

			h.Login(w, req)
			resp := w.Result()
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				require.NoError(t, err)
			}(resp.Body)

			require.Equal(t, tt.wantStatus, resp.StatusCode)
			respBody, _ := io.ReadAll(resp.Body)
			require.Contains(t, string(respBody), tt.wantBodySubstr)
		})
	}
}

func TestAuthHandler_DummyLogin(t *testing.T) {
	tests := []struct {
		name           string
		body           []byte
		serviceToken   dto.Token
		serviceErr     error
		wantStatus     int
		wantBodySubstr string
	}{
		{
			name:           "invalid JSON",
			body:           []byte(`{"role":}`),
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: "Invalid request",
		},
		{
			name:           "wrong role",
			body:           []byte(`{"email":"a@b.c","role":"x"}`),
			serviceErr:     models.ErrIncorrectUserRole,
			wantStatus:     http.StatusBadRequest,
			wantBodySubstr: models.ErrIncorrectUserRole.Error(),
		},
		{
			name:           "internal err",
			body:           []byte(`{"email":"a@b.c","role":"user"}`),
			serviceErr:     errors.New("err"),
			wantStatus:     http.StatusInternalServerError,
			wantBodySubstr: "err",
		},
		{
			name:           "success",
			body:           []byte(`{"email":"a@b.c","role":"user"}`),
			serviceToken:   "tok",
			wantStatus:     http.StatusOK,
			wantBodySubstr: `"tok"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &stubAuthService{
				DummyLoginFunc: func(req dto.PostDummyLoginJSONRequestBody) (*dto.Token, error) {
					return &tt.serviceToken, tt.serviceErr
				},
			}
			h := NewAuthHandler(stub)

			req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewReader(tt.body))
			w := httptest.NewRecorder()

			h.DummyLogin(w, req)
			resp := w.Result()
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				require.NoError(t, err)
			}(resp.Body)

			require.Equal(t, tt.wantStatus, resp.StatusCode)
			respBody, _ := io.ReadAll(resp.Body)
			require.Contains(t, string(respBody), tt.wantBodySubstr)
		})
	}
}
