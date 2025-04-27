package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func generateToken(t *testing.T, secretKey string, role string, exp time.Time) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"role": role,
		"exp":  exp.Unix(),
	})
	tokenString, err := token.SignedString([]byte(secretKey))
	require.NoError(t, err)
	return tokenString
}

func TestCheckAuth(t *testing.T) {
	jwtSecretKey = "test_secret"

	validToken := generateToken(t, jwtSecretKey, "admin", time.Now().Add(time.Hour))
	expiredToken := generateToken(t, jwtSecretKey, "admin", time.Now().Add(-time.Hour*25))

	tests := []struct {
		name            string
		authHeader      string
		wantStatus      int
		wantResponseSub string
	}{
		{
			name:            "no auth header",
			authHeader:      "",
			wantStatus:      http.StatusUnauthorized,
			wantResponseSub: "Authorization header required",
		},
		{
			name:            "invalid token",
			authHeader:      "Bearer invalidtoken",
			wantStatus:      http.StatusUnauthorized,
			wantResponseSub: "error while parsing token",
		},
		{
			name:            "expired token",
			authHeader:      "Bearer " + expiredToken,
			wantStatus:      http.StatusUnauthorized,
			wantResponseSub: "error while parsing token",
		},
		{
			name:            "valid token",
			authHeader:      "Bearer " + validToken,
			wantStatus:      http.StatusOK,
			wantResponseSub: "OK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := CheckAuth()(http.HandlerFunc(dummyHandler))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			mw.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			require.Equal(t, tt.wantStatus, resp.StatusCode)
			require.Contains(t, string(body), tt.wantResponseSub)
		})
	}
}
