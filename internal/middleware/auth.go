package middleware

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/itisalisas/avito-backend/internal/utils"
	"net/http"
	"os"
	"strings"
	"time"
)

type contextKey string

const userRoleKey contextKey = "userRole"

var jwtSecretKey = os.Getenv("JWT_SECRET_KEY")

func CheckAuth() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				utils.WriteResponse(w, utils.Error("Authorization header required"), http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenStr == "" {
				utils.WriteResponse(w, utils.Error("Authorization header required"), http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("error while generating token")
				}
				return []byte(jwtSecretKey), nil
			})

			if err != nil || !token.Valid {
				utils.WriteResponse(w, utils.Error("Error while generating token"), http.StatusInternalServerError)
				return
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				expTime, err := claims.GetExpirationTime()
				if err != nil {
					utils.WriteResponse(w, utils.Error(err.Error()), http.StatusInternalServerError)
					return
				}
				if expTime.Before(time.Now()) {
					utils.WriteResponse(w, utils.Error("Token expired"), http.StatusUnauthorized)
					return
				}
				ctx := context.WithValue(r.Context(), userRoleKey, claims["role"])
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				utils.WriteResponse(w, utils.Error("Token invalid"), http.StatusUnauthorized)
			}
		})
	}
}
