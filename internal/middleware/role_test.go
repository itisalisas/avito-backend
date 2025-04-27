package middleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
)

func dummyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		return
	}
}

func TestCheckRole(t *testing.T) {
	tests := []struct {
		name       string
		userRole   string
		allowed    []dto.PostRegisterJSONBodyRole
		wantStatus int
	}{
		{
			name:       "role allowed",
			userRole:   "admin",
			allowed:    []dto.PostRegisterJSONBodyRole{"admin", "manager"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "role forbidden",
			userRole:   "client",
			allowed:    []dto.PostRegisterJSONBodyRole{"admin", "manager"},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := CheckRole(tt.allowed...)(http.HandlerFunc(dummyHandler))

			ctx := context.WithValue(context.Background(), userRoleKey, tt.userRole)
			req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)

			w := httptest.NewRecorder()
			mw.ServeHTTP(w, req)

			resp := w.Result()
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				require.NoError(t, err)
			}(resp.Body)

			require.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}
